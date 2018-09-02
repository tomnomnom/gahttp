package gahttp

import (
	"crypto/tls"
	"io"
	"net/http"
	"sync"
	"time"
)

func getDefaultClient() *http.Client {

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &http.Client{
		Transport: transport,
	}
}

type ProcFn func(*http.Request, *http.Response, error)

type request struct {
	req *http.Request
	fn  ProcFn
}

type Pipeline struct {
	concurrency int

	client *http.Client
	reqs   chan request

	running bool
	wg      sync.WaitGroup

	rl          *rateLimiter
	rateLimited bool
}

func New(concurrency int) *Pipeline {
	return &Pipeline{
		concurrency: concurrency,

		client: getDefaultClient(),
		reqs:   make(chan request),

		running: false,

		rl:          newRateLimiter(0),
		rateLimited: false,
	}
}

func NewWithClient(concurrency int, client *http.Client) *Pipeline {
	p := New(concurrency)
	p.client = client
	return p
}

func (p *Pipeline) SetRateLimit(d time.Duration) {
	if p.running {
		return
	}

	if d == 0 {
		p.rateLimited = false
	} else {
		p.rateLimited = true
	}

	p.rl.delay = d
}

func (p *Pipeline) SetClient(c *http.Client) {
	if p.running {
		return
	}
	p.client = c
}

func (p *Pipeline) SetConcurrency(c int) {
	if p.running {
		return
	}
	p.concurrency = c
}

func (p *Pipeline) Do(r *http.Request, fn ProcFn) {
	if !p.running {
		p.Run()
	}

	p.reqs <- request{r, fn}
}

func (p *Pipeline) Get(u string, fn ProcFn) error {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}
	p.Do(req, fn)
	return nil
}

func (p *Pipeline) Post(u string, b io.Reader, fn ProcFn) error {
	req, err := http.NewRequest("GET", u, b)
	if err != nil {
		return err
	}
	p.Do(req, fn)
	return nil
}

func (p *Pipeline) Done() {
	close(p.reqs)
}

func (p *Pipeline) Run() {
	if p.running {
		return
	}
	p.running = true

	// launch workers
	for i := 0; i < p.concurrency; i++ {
		p.wg.Add(1)
		go func() {
			for r := range p.reqs {
				if p.rateLimited {
					p.rl.Block(r.req.URL.Hostname())
				}

				resp, err := p.client.Do(r.req)
				r.fn(r.req, resp, err)
			}
			p.wg.Done()
		}()
	}
}

func (p *Pipeline) Wait() {
	p.wg.Wait()
}

func CloseBody(fn ProcFn) ProcFn {
	return func(req *http.Request, resp *http.Response, err error) {
		fn(req, resp, err)

		if resp == nil {
			return
		}
		if resp.Body != nil {
			resp.Body.Close()
		}
	}
}

func IfNoError(fn ProcFn) ProcFn {
	return func(req *http.Request, resp *http.Response, err error) {
		if err != nil {
			return
		}
		fn(req, resp, err)
	}
}

func Wrap(fn ProcFn, middleware ...func(ProcFn) ProcFn) ProcFn {
	for _, m := range middleware {
		fn = m(fn)
	}
	return fn
}

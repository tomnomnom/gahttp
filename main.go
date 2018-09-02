package gahttp

import (
	"io"
	"net/http"
	"sync"
	"time"
)

// getDefaultClient returns the default HTTP client
func getDefaultClient() *http.Client {
	return &http.Client{}
}

// a ProcFn is a function that processes an HTTP response.
// The HTTP request is provided for context, along with any
// error that occurred.
type ProcFn func(*http.Request, *http.Response, error)

// a request wraps a Go HTTP request struct, and a ProcFn
// to process its result
type request struct {
	req *http.Request
	fn  ProcFn
}

// a Pipeline is the main component of the gahttp package.
// It orchestrates making requests, optionally rate limiting them
type Pipeline struct {
	concurrency int

	client *http.Client
	reqs   chan request

	running bool
	wg      sync.WaitGroup

	rl          *rateLimiter
	rateLimited bool
}

// New returns a new *Pipeline for the provided concurrency level
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

// NewWithClient returns a new *Pipeline for the provided concurrency
// level, and uses the provided *http.Client to make requests
func NewWithClient(concurrency int, client *http.Client) *Pipeline {
	p := New(concurrency)
	p.client = client
	return p
}

// SetRateLimit sets the delay between requests to a given hostname
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

// SetRateLimitMillis sets the delay between request to a given hostname
// in milliseconds. This function is provided as a convenience, to make
// it easy to accept integer values as command line arguments.
func (p *Pipeline) SetRateLimitMillis(m int) {
	p.SetRateLimit(time.Duration(m * 1000000))
}

// SetClient sets the HTTP client used by the pipeline to make HTTP
// requests. It can only be set before the pipeline is running
func (p *Pipeline) SetClient(c *http.Client) {
	if p.running {
		return
	}
	p.client = c
}

// SetConcurrency sets the concurrency level for the pipeline.
// It can only be set before the pipeline is running
func (p *Pipeline) SetConcurrency(c int) {
	if p.running {
		return
	}
	p.concurrency = c
}

// Do is the pipeline's generic request function; similar to
// http.DefaultClient.Do(), but it also accepts a ProcFn which
// will be called when the request has been executed
func (p *Pipeline) Do(r *http.Request, fn ProcFn) {
	if !p.running {
		p.Run()
	}

	p.reqs <- request{r, fn}
}

// Get is a convenience wrapper around the Do() function for making
// HTTP GET requests. It accepts a URL and the ProcFn to process
// the response.
func (p *Pipeline) Get(u string, fn ProcFn) error {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}
	p.Do(req, fn)
	return nil
}

// Post is a convenience wrapper around the Do() function for making
// HTTP POST requests. It accepts a URL, an io.Reader for the POST
// body, and a ProcFn to process the response.
func (p *Pipeline) Post(u string, body io.Reader, fn ProcFn) error {
	req, err := http.NewRequest("GET", u, body)
	if err != nil {
		return err
	}
	p.Do(req, fn)
	return nil
}

// Done should be called to signal to the pipeline that all requests
// that will be made have been enqueued. This closes the internal
// channel used to send requests to the workers that are executing
// the HTTP requests.
func (p *Pipeline) Done() {
	close(p.reqs)
}

// Run puts the pipeline into a running state. It launches the
// worker processes that execute the HTTP requests. Run() is
// called automatically by Do(), Get() and Post(), so it's often
// not necessary to call it directly.
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

// Wait blocks until all requests in the pipeline have been executed
func (p *Pipeline) Wait() {
	p.wg.Wait()
}

// CloseBody wraps a ProcFn and returns a version of it that automatically
// closed the response body
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

// Wrap accepts a ProcFn and applies any number of 'middleware' functions
// to it (e.g. the CloseBody function).
func Wrap(fn ProcFn, middleware ...func(ProcFn) ProcFn) ProcFn {
	for _, m := range middleware {
		fn = m(fn)
	}
	return fn
}

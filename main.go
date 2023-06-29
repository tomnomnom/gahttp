package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/fcynic3/gahttp"
	"golang.org/x/net/html"
)

func extractTitle(req *http.Request, resp *http.Response, err error) {
	if err != nil {
		return
	}

	z := html.NewTokenizer(resp.Body)

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		t := z.Token()

		if t.Type == html.StartTagToken && t.Data == "title" {
			if z.Next() == html.TextToken {
				title := strings.TrimSpace(z.Token().Data)
				fmt.Printf("%s (%s)\n", title, req.URL)
				break
			}
		}

	}
}

func main() {
	var (
		concurrency int
		proxyURL    string
	)

	flag.IntVar(&concurrency, "c", 20, "Concurrency")
	flag.StringVar(&proxyURL, "proxy", "", "Proxy URL")
	flag.Parse()

	p := gahttp.NewPipeline()
	p.SetConcurrency(concurrency)

	if proxyURL != "" {
		proxyURL, err := url.Parse(proxyURL)
		if err != nil {
			fmt.Println("Failed to parse proxy URL:", err)
			return
		}
		p.HTTPClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	extractFn := gahttp.Wrap(extractTitle, gahttp.CloseBody)

	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		req, err := http.NewRequest("GET", sc.Text(), nil)
		if err != nil {
			fmt.Println("Failed to create request:", err)
			continue
		}
		p.Do(req, extractFn)
	}
	p.Done()

	p.Wait()
}

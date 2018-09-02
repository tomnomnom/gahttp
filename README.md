# gahttp

Async/concurrent HTTP requests for Go.

An attempt to handle some of the boilerplate of doing concurrent HTTP requests in Go.

Work in progress.


## Example

```golang
package main

import (
	"fmt"
	"net/http"

	"github.com/tomnomnom/gahttp"
)

func printStatus(req *http.Request, resp *http.Response, err error) {
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return
	}
	fmt.Printf("%s: %s\n", req.URL, resp.Status)
}

func main() {
	p := gahttp.New(20)

	urls := []string{
		"http://example.com",
		"http://example.net",
		"http://example.org",
	}

	for _, u := range urls {
		p.Get(u, printStatus)
	}
	p.Done()

	p.Wait()
}

```

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

## TODO

* Optional rate limiting
* Functions to return commonly used clients (e.g. ignore cert errors, don't follow redirects)
* `DoneAndWait()` func?
* Tests (lol)
* Helper for writing responses to channel? (e.g. `func ChanWriter() (chan *Response, procFn)`)
    - For when you don't want to do the work concurrently

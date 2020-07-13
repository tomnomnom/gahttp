# gahttp

Async/concurrent HTTP requests for Go with rate-limiting.

Work in progress.


## Example

```golang
package main

import (
    "fmt"
    "net/http"
    "time"

    "github.com/tomnomnom/gahttp"
)

func printStatus(req *http.Request, resp *http.Response, err error) {
    if err != nil {
        return
    }
    fmt.Printf("%s: %s\n", req.URL, resp.Status)
}

func main() {
    p := gahttp.NewPipeline()
    p.SetConcurrency(20)
    p.SetRateLimit(time.Second * 1)

    urls := []string{
        "http://example.com",
        "http://example.com",
        "http://example.com",
        "http://example.net",
        "http://example.org",
    }
    
    actualHost := "127.0.0.1"

    for _, u := range urls {
        p.Get(u, gahttp.Wrap(printStatus, gahttp.CloseBody))

        // ... or 

        p.GetFromHost(u, actualHost, gahttp.Wrap(printStatus, gahttp.CloseBody))
    }
    p.Done()

    p.Wait()
}
```

## TODO

* `DoneAndWait()` func?
* Helper for writing responses to channel? (e.g. `func ChanWriter() (chan *Response, procFn)`)
    - For when you don't want to do the work concurrently
* Actually handle timeouts / provide context interface for cancellation etc?

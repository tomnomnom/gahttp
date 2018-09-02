package gahttp

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSmoke(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "a response")
	}))
	defer ts.Close()

	p := New(2)

	p.Get(ts.URL, func(req *http.Request, resp *http.Response, err error) {
		if err != nil {
			t.Fatalf("want non-nil error passed to fn; have %s", err)
		}

		if resp == nil {
			t.Fatalf("resp should not be nil")
		}

		if resp.Body == nil {
			t.Fatalf("resp body should not be nil")
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("should have no error reading from body")
		}

		if strings.TrimSpace(string(b)) != "a response" {
			t.Errorf("want 'a response' read from resp.Body; have '%s'", b)
		}
	})
	p.Done()
	p.Wait()
}

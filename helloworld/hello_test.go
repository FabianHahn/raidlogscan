package helloworld

import (
	"net/http/httptest"
	"testing"
)

func TestHelloGet(t *testing.T) {

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	helloGet(rr, req)

	want := "Hello, World!"
	if got := rr.Body.String(); got != want {
		t.Errorf("helloGet() = %q, want %q", got, want)
	}
}


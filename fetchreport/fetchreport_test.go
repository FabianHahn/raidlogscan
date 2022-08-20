package fetchreport

import (
	"fmt"
	"net/http/httptest"
	"testing"
)

const (
	testReportCode = "xkyT2H6VGM3PqZ1p"
)

func TestHelloGet(t *testing.T) {

	req := httptest.NewRequest("GET", fmt.Sprintf("/?code=%v", testReportCode), nil)
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	fetchReport(rr, req)

	t.Log(rr.Body.String())
}

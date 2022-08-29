package accountstats

import (
	"fmt"
	"net/http/httptest"
	"testing"
)

const (
	testAccountName = "Jaythe"
)

func TestAccountStats(t *testing.T) {
	req := httptest.NewRequest("GET", fmt.Sprintf("/?account_name=%v", testAccountName), nil)
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	accountStats(rr, req)

	t.Log(rr.Body.String())
}

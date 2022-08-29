package claimaccount

import (
	"fmt"
	"net/http/httptest"
	"testing"
)

const (
	testAccount  = "Jaythe"
	testPlayerId = "38937027"
)

func TestClaimAccount(t *testing.T) {
	req := httptest.NewRequest("GET", fmt.Sprintf("/?account_name=%v&player_id=%v", testAccount, testPlayerId), nil)
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	claimAccount(rr, req)

	t.Log(rr.Body.String())
}

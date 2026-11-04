package server

import (
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestToolCallQueryValidation(t *testing.T) {
	for _, raw := range []string{"limit=0", "limit=51", "limit=x", "type=anything", "status=pending", "limit=2&limit=3", "url=other", "cursor=garbage"} {
		if _, err := parseToolCallQuery(httptest.NewRequest("GET", "/?"+raw, nil), "session"); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	data, _ := json.Marshal(toolCallCursor{Scope: "session", Q: "Bash", Before: 12, Snapshot: 20})
	cursor := base64.RawURLEncoding.EncodeToString(data)
	r := httptest.NewRequest("GET", "/?q=Bash&cursor="+cursor, nil)
	q, err := parseToolCallQuery(r, "session")
	if err != nil || q.Before != 12 || q.Limit != 20 {
		t.Fatalf("query %+v %v", q, err)
	}
	if _, err = parseToolCallQuery(r, "other"); err == nil {
		t.Fatal("cross session cursor accepted")
	}
	if _, err = parseToolCallQuery(httptest.NewRequest("GET", "/?q=Read&cursor="+cursor, nil), "session"); err == nil {
		t.Fatal("changed filter accepted")
	}
}

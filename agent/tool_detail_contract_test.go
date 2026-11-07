package agent

import (
	"encoding/json"
	"testing"
)

func TestToolDetailInputValidation(t *testing.T) {
	for _, raw := range []string{``, `null`, `[]`, `{} {}`, `{"unknown":1}`, `{"field":null}`, `{"offset":"1"}`, `{"max_chars":0}`, `{"max_chars":24001}`, `{"max_chars":-1}`, `{"offset":-1}`, `{"index":-1}`, `{"field":"bad"}`, `{"index":1.5}`} {
		t.Run(raw, func(t *testing.T) {
			if ValidateToolDetailInput(json.RawMessage(raw)) == nil {
				t.Fatal("invalid request accepted")
			}
		})
	}
	for _, raw := range []string{`{}`, `{"field":"/a~1b","max_chars":1}`, `{"index":20,"offset":0,"max_chars":24000}`} {
		if err := ValidateToolDetailInput(json.RawMessage(raw)); err != nil {
			t.Fatal(err)
		}
	}
}

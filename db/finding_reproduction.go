package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Ported from SERIES db/finding_reproduction.go. The asset-approval host check
// (validateReproductionTrafficTx) is intentionally NOT ported: it depends on
// task_asset_links / approval states that ARTEX does not have. Steps may only
// reference traffic already bound to this finding (see Resolve).

// FindingReproduction is immutable, recorded at discovery time. Wire evidence is
// kept verbatim; only explicit traffic references are resolved, never fuzzy matches.
type FindingReproduction struct {
	EntryPoint    string                    `json:"entry_point,omitempty"`
	Prerequisites []string                  `json:"prerequisites,omitempty"`
	Steps         []FindingReproductionStep `json:"steps"`
}

type FindingReproductionStep struct {
	Title       string       `json:"title"`
	Action      string       `json:"action"`
	Result      string       `json:"result"`
	Proves      string       `json:"proves,omitempty"`
	Request     string       `json:"request,omitempty"`
	Response    string       `json:"response,omitempty"`
	Command     string       `json:"command,omitempty"`
	Output      string       `json:"output,omitempty"`
	TrafficRefs []TrafficRef `json:"traffic_refs,omitempty"`
	BindingIDs  []string     `json:"binding_ids,omitempty"` // server-owned, scoped to this finding
}

func DecodeFindingReproduction(raw json.RawMessage) (*FindingReproduction, error) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}
	if !json.Valid(raw) {
		return nil, fmt.Errorf("reproduction 必须为有效JSON")
	}
	if len(raw) > 1024*1024 {
		return nil, fmt.Errorf("reproduction 最多1MiB，请用流量引用保留长报文")
	}
	var out FindingReproduction
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("reproduction 格式错误: %w", err)
	}
	if err := out.ValidateInput(); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *FindingReproduction) ValidateInput() error {
	if r == nil {
		return nil
	}
	raw, err := json.Marshal(r)
	if err != nil || len(raw) > 1024*1024 {
		return fmt.Errorf("reproduction 最多1MiB")
	}
	if len(r.Steps) == 0 || len(r.Steps) > 50 || len(r.Prerequisites) > 50 {
		return fmt.Errorf("reproduction.steps 必须为1..50步，前置条件最多50条")
	}
	for i, step := range r.Steps {
		if strings.TrimSpace(step.Title) == "" || strings.TrimSpace(step.Action) == "" || strings.TrimSpace(step.Result) == "" {
			return fmt.Errorf("reproduction.steps[%d] 需要 title/action/result；缺失的真实结果须明确说明", i)
		}
		if len(step.BindingIDs) > 0 {
			return fmt.Errorf("reproduction.steps[%d].binding_ids 由服务端生成，请使用真实 traffic_refs", i)
		}
		if _, err := NormalizeTrafficRefs(step.TrafficRefs); err != nil {
			return fmt.Errorf("reproduction.steps[%d].traffic_refs: %w", i, err)
		}
	}
	return nil
}

func (r *FindingReproduction) TrafficReferences() []TrafficRef {
	var refs []TrafficRef
	if r != nil {
		for _, step := range r.Steps {
			refs = append(refs, step.TrafficRefs...)
		}
	}
	return refs
}

// Resolve returns a copy; failed transactions must not mutate the caller's input.
func (r *FindingReproduction) Resolve(bindings []FindingTrafficBinding) (json.RawMessage, error) {
	if r == nil {
		return nil, nil
	}
	raw, _ := json.Marshal(r)
	var out FindingReproduction
	_ = json.Unmarshal(raw, &out)
	byTraffic := map[string]string{}
	for _, b := range bindings {
		byTraffic[b.Snapshot.SourceTrafficID] = fmt.Sprint(b.ID)
	}
	for i := range out.Steps {
		step := &out.Steps[i]
		seen := map[string]bool{}
		for _, ref := range step.TrafficRefs {
			id := byTraffic[strings.TrimSpace(ref.TrafficID)]
			if id == "" {
				return nil, fmt.Errorf("reproduction.steps[%d] 的流量未绑定到当前漏洞", i)
			}
			if !seen[id] {
				step.BindingIDs = append(step.BindingIDs, id)
				seen[id] = true
			}
		}
		step.TrafficRefs = nil
	}
	return json.Marshal(out)
}

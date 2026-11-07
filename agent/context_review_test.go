package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/norma/llm"
	actool "github.com/Autumn-27/norma/tool"
	"github.com/neouks/series/db"
	"iter"
	"strings"
	"testing"
)

type contextReviewProvider struct {
	captureUsageProvider
	requests *[]llm.CompletionRequest
}

func (p contextReviewProvider) Stream(ctx context.Context, req llm.CompletionRequest) iter.Seq2[llm.StreamEvent, error] {
	*p.requests = append(*p.requests, req)
	return p.captureUsageProvider.Stream(ctx, req)
}

func TestContextHistoryDoesNotInterpretEvidenceAsPlatformIDs(t *testing.T) {
	d := testDB(t)
	defer d.Close()
	task, e := d.CreateTask("history IDs", "goal", nil, 0, 0)
	if e != nil {
		t.Fatal(e)
	}
	defer d.DeleteTask(task.ID)
	s := d.Exploration(task.ExplorationID)
	id, e := s.AddNode(db.KindFact, map[string]any{"summary": "business evidence"}, 0, "confirmed", "test", nil)
	if e != nil {
		t.Fatal(e)
	}
	body := fmt.Sprintf(`{"items":[{"node_id":%d,"kind":"fact","payload":{"entries":[{"field":"/detail","value":{"asset_id":999999999,"node_id":999999998,"evidence":"KEEP BUSINESS JSON","number":9007199254740993}}]}}]}`, id)
	req := llm.CompletionRequest{Messages: []llm.Message{
		{Role: llm.RoleAssistant, Content: []llm.ContentBlock{{Type: llm.BlockToolUse, ID: "read", Name: "context_get", Input: json.RawMessage(`{}`)}}},
		{Role: llm.RoleUser, Content: []llm.ContentBlock{llm.ToolResultText("read", body, false)}},
	}}
	out, e := FilterTaskAssetRequest(WithRunInfo(t.Context(), RunInfo{AgentKey: "planner", TaskID: task.ID}), d.Assets(), task.ID, req)
	if e != nil {
		t.Fatal(e)
	}
	result := out.Messages[1].Content[0].Content[0].Text
	if !strings.Contains(result, "KEEP BUSINESS JSON") || !strings.Contains(result, "9007199254740993") {
		t.Fatal("valid evidence was removed/changed:", result)
	}
	if req.Messages[1].Content[0].Content[0].Text != body {
		t.Fatal("audit mutated")
	}
}

func TestAssetReadRequiresTrustedWorkerContext(t *testing.T) {
	d := testDB(t)
	defer d.Close()
	task, e := d.CreateTaskWithOptions("trusted worker read", "goal", db.TaskCreateOptions{AssetApprovalTemplate: "explicit_targets"})
	if e != nil {
		t.Fatal(e)
	}
	defer d.DeleteTask(task.ID)
	id, e := d.Assets().UpsertRootDomain(db.UpsertRootDomainReq{Domain: fmt.Sprintf("trusted-%d.test", task.ID), TaskID: task.ID, AgentDiscovered: true})
	if e != nil {
		t.Fatal(e)
	}
	ts := NewToolSet(d.Exploration(task.ExplorationID), "worker")
	ts.SetTaskID(task.ID)
	ts.SetAssetStore(d.Assets(), d.Companies())
	ts.workerExecution = true
	for _, ri := range []RunInfo{{}, {TaskID: task.ID, AgentKey: "planner"}, {TaskID: task.ID + 1, AgentKey: "worker", IntentID: 1}, {TaskID: task.ID, AgentKey: "worker"}, {TaskID: task.ID, AgentKey: "worker", IntentID: 1}} {
		ctx := WithRunInfo(t.Context(), ri)
		r, e := ts.assetGet().Call(ctx, json.RawMessage(fmt.Sprintf(`{"ids":[%d]}`, id)), nil)
		if e != nil || r.IsError {
			t.Fatal(r, e)
		}
		valid := ri.TaskID == task.ID && ri.AgentKey == "worker" && ri.IntentID > 0
		if strings.Contains(r.Flatten(), `"details"`) != valid {
			t.Fatal("asset read role mismatch", ri, r.Flatten())
		}
		r, e = ts.assetSearch().Call(ctx, json.RawMessage(fmt.Sprintf(`{"dsl":"domain==trusted-%d.test"}`, task.ID)), nil)
		if e != nil || r.IsError {
			t.Fatal(r, e)
		}
		if strings.Contains(r.Flatten(), `"target"`) != valid {
			t.Fatal("asset search role mismatch", ri, r.Flatten())
		}
	}
}

func TestStructuredCapturePreservesExactAuthorizationID(t *testing.T) {
	r, _ := jsonResult(map[string]any{"node_id": int64(9007199254740993), "summary": strings.Repeat("evidence", 2000)})
	out := captureStructuredResult("context_get", r, &actool.ToolContext{MaxOutputChars: 1000})
	var data struct {
		IDs []int64 `json:"node_ids"`
	}
	if e := json.Unmarshal([]byte(out.Flatten()), &data); e != nil {
		t.Fatal(e)
	}
	if len(data.IDs) != 1 || data.IDs[0] != 9007199254740993 {
		t.Fatal("capture changed authorization identity", out.Flatten())
	}
}

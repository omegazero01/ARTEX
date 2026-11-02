package agent

import (
	"context"
	"iter"
	"testing"

	"github.com/Autumn-27/norma/harness"
	"github.com/Autumn-27/norma/llm"
)

const testSettlePrompt = "预算已用完，请收尾。"

func TestSettlementGateLatchesOnWrapUpPrompt(t *testing.T) {
	g := NewSettlementGate(&harness.Settlement{Prompt: testSettlePrompt}, "Bash")
	if g.Blocked("Bash") {
		t.Fatal("gate must not block before the wrap-up prompt appears")
	}
	g.observe([]llm.Message{llm.UserText("开始执行意图")})
	if g.Blocked("Bash") {
		t.Fatal("ordinary user messages must not arm the gate")
	}
	g.observe([]llm.Message{llm.UserText("开始执行意图"), llm.UserText(testSettlePrompt)})
	if !g.Settling() || !g.Blocked("Bash") {
		t.Fatalf("expected Bash blocked after wrap-up prompt, settling=%v", g.Settling())
	}
	if g.Blocked("read_file") {
		t.Fatal("only guarded tools may be blocked")
	}
}

func TestSettlementGateMatchesPrefix(t *testing.T) {
	g := NewSettlementGate(&harness.Settlement{Prompt: testSettlePrompt}, "Bash")
	g.observe([]llm.Message{llm.UserText(testSettlePrompt + "\n（系统补充）")})
	if !g.Blocked("Bash") {
		t.Fatal("prefix match should arm the gate")
	}
}

func TestSettlementGateUsesPromptByReason(t *testing.T) {
	g := NewSettlementGate(&harness.Settlement{
		Prompt: "per-run 收尾",
		PromptByReason: map[harness.TerminalReason]string{
			harness.ReasonTimeout: "任务到点收尾",
		},
	}, "Bash")
	g.observe([]llm.Message{llm.UserText("任务到点收尾")})
	if !g.Blocked("Bash") {
		t.Fatal("PromptByReason text should arm the gate")
	}
}

func TestSettlementGateInertWithoutPromptOrGuard(t *testing.T) {
	cases := []*SettlementGate{
		NewSettlementGate(nil, "Bash"),
		NewSettlementGate(&harness.Settlement{}, "Bash"),
		NewSettlementGate(&harness.Settlement{Prompt: testSettlePrompt}),
	}
	for i, g := range cases {
		if g.Enabled() {
			t.Fatalf("case %d: gate should be inert", i)
		}
		g.observe([]llm.Message{llm.UserText(testSettlePrompt)})
		if g.Blocked("Bash") {
			t.Fatalf("case %d: inert gate must never block", i)
		}
	}
}

type countingProvider struct {
	streams   int
	completes int
}

func (p *countingProvider) Stream(context.Context, llm.CompletionRequest) iter.Seq2[llm.StreamEvent, error] {
	p.streams++
	return func(yield func(llm.StreamEvent, error) bool) {}
}

func (p *countingProvider) Complete(context.Context, llm.CompletionRequest) (llm.Message, string, llm.Usage, error) {
	p.completes++
	return llm.Message{Role: llm.RoleAssistant}, "stop", llm.Usage{}, nil
}

func TestSettlementGateWrapDepsObservesRequests(t *testing.T) {
	prov := &countingProvider{}
	g := NewSettlementGate(&harness.Settlement{Prompt: testSettlePrompt}, "Bash")
	var deps harness.QueryDeps
	g.WrapDeps(&deps, prov)
	if deps.CallModel == nil || deps.CallModelSync == nil {
		t.Fatal("WrapDeps must install both call paths from the provider")
	}
	req := llm.CompletionRequest{Messages: []llm.Message{llm.UserText(testSettlePrompt)}}
	for range deps.CallModel(context.Background(), req) {
	}
	if _, _, _, err := deps.CallModelSync(context.Background(), req); err != nil {
		t.Fatalf("sync path: %v", err)
	}
	if prov.streams != 1 || prov.completes != 1 {
		t.Fatalf("provider not called through: streams=%d completes=%d", prov.streams, prov.completes)
	}
	if !g.Blocked("Bash") {
		t.Fatal("wrapped requests should arm the gate")
	}
}

func TestSettlementBlocksFromContext(t *testing.T) {
	g := NewSettlementGate(&harness.Settlement{Prompt: testSettlePrompt}, "Bash")
	ctx := WithSettlementGate(context.Background(), g)
	if SettlementBlocks(ctx, "Bash") {
		t.Fatal("fresh context must allow tools")
	}
	g.observe([]llm.Message{llm.UserText(testSettlePrompt)})
	if !SettlementBlocks(ctx, "Bash") {
		t.Fatal("gate in context should block Bash once settling")
	}
	if SettlementBlocks(context.Background(), "Bash") {
		t.Fatal("context without a gate must allow tools")
	}
}

package agent

import (
	"context"
	"iter"
	"strings"
	"sync/atomic"

	"github.com/Autumn-27/norma/harness"
	"github.com/Autumn-27/norma/llm"
)

// SettlementGate keeps the tool list stable during a run's wrap-up phase.
//
// The SDK's Settlement.DisabledTools removes tools from the *request* when a run
// hits its budget. The tool array sits at the front of the request, so removing
// one entry rewrites the prefix and invalidates the provider's prompt cache for
// every remaining turn of that run (measured on SERIES: 5.6%-6.5% of task cost).
//
// Instead of hiding the tool, this gate watches the model requests: the SDK
// appends the wrap-up prompt verbatim as the trailing user message, so the first
// request whose last message equals that prompt marks the run as settling. From
// then on PreToolUse refuses the guarded tools — same safety, byte-stable prefix.
//
// A nil or empty gate is inert, so callers can leave it unset.
type SettlementGate struct {
	prompts  []string
	guard    map[string]bool
	settling atomic.Bool
}

// NewSettlementGate arms a gate for s. guard lists the tools that must not run
// during wrap-up; an empty list (or a settlement without prompts) makes the gate
// inert.
func NewSettlementGate(s *harness.Settlement, guard ...string) *SettlementGate {
	g := &SettlementGate{guard: make(map[string]bool, len(guard))}
	if s == nil || len(guard) == 0 {
		return g
	}
	if p := strings.TrimSpace(s.Prompt); p != "" {
		g.prompts = append(g.prompts, p)
	}
	for _, p := range s.PromptByReason {
		if p = strings.TrimSpace(p); p != "" {
			g.prompts = append(g.prompts, p)
		}
	}
	for _, name := range guard {
		if name = strings.TrimSpace(name); name != "" {
			g.guard[name] = true
		}
	}
	return g
}

// Enabled reports whether the gate can ever latch (prompts + guarded tools).
func (g *SettlementGate) Enabled() bool {
	return g != nil && len(g.prompts) > 0 && len(g.guard) > 0
}

// Blocked reports whether tool must be refused right now. It stays false until
// the wrap-up prompt has actually been observed.
func (g *SettlementGate) Blocked(tool string) bool {
	return g.Enabled() && g.settling.Load() && g.guard[tool]
}

// Settling reports whether the wrap-up prompt has been seen (for tests/logs).
func (g *SettlementGate) Settling() bool { return g != nil && g.settling.Load() }

// WrapDeps makes the gate observe every model request. Provider defaults are
// applied here because harness.QueryDeps.withDefaults would otherwise install
// Provider.Stream/Complete after we wrap and bypass the observation entirely.
func (g *SettlementGate) WrapDeps(d *harness.QueryDeps, prov llm.Provider) {
	if d == nil || !g.Enabled() || prov == nil {
		return
	}
	stream := d.CallModel
	if stream == nil {
		stream = prov.Stream
	}
	syncCall := d.CallModelSync
	if syncCall == nil {
		syncCall = prov.Complete
	}
	d.CallModel = func(ctx context.Context, req llm.CompletionRequest) iter.Seq2[llm.StreamEvent, error] {
		g.observe(req.Messages)
		return stream(ctx, req)
	}
	d.CallModelSync = func(ctx context.Context, req llm.CompletionRequest) (llm.Message, string, llm.Usage, error) {
		g.observe(req.Messages)
		return syncCall(ctx, req)
	}
}

// observe latches "settling" the first time the SDK's wrap-up text shows up as
// the trailing user message. Matching is exact-or-prefix: the SDK appends the
// prompt verbatim (harness/query.go beginSettlement → llm.UserText(prompt)),
// while a prefix also covers SDKs that append their own suffix.
func (g *SettlementGate) observe(msgs []llm.Message) {
	if !g.Enabled() || g.settling.Load() || len(msgs) == 0 {
		return
	}
	last := msgs[len(msgs)-1]
	if last.Role != llm.RoleUser {
		return
	}
	text := strings.TrimSpace(messageText(last))
	if text == "" {
		return
	}
	for _, p := range g.prompts {
		if text == p || strings.HasPrefix(text, p) {
			g.settling.Store(true)
			return
		}
	}
}

type settlementGateKey struct{}

// WithSettlementGate carries the gate to hooks, which receive the run context
// but not the worker.
func WithSettlementGate(ctx context.Context, g *SettlementGate) context.Context {
	return context.WithValue(ctx, settlementGateKey{}, g)
}

// SettlementBlocks reports whether tool is refused by the gate attached to ctx.
// Hooks call this from PreToolUse; no gate in ctx = allow.
func SettlementBlocks(ctx context.Context, tool string) bool {
	g, _ := ctx.Value(settlementGateKey{}).(*SettlementGate)
	return g.Blocked(tool)
}

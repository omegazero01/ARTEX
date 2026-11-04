package agent

import (
	"strings"
	"testing"

	actool "github.com/Autumn-27/norma/tool"
)

func TestBashPromptOverridePerHost(t *testing.T) {
	if _, ok := bashPromptOverrideFor("linux"); ok {
		t.Fatal("non-Windows hosts must keep the SDK prompt")
	}
	if _, ok := bashPromptOverrideFor("darwin"); ok {
		t.Fatal("non-Windows hosts must keep the SDK prompt")
	}
	prompt, ok := bashPromptOverrideFor("windows")
	if !ok || prompt == "" {
		t.Fatal("Windows hosts must replace the SDK prompt")
	}
	for _, bad := range []string{"chain steps with", "heredoc"} {
		if strings.Contains(prompt, bad) {
			t.Fatalf("replacement prompt still advertises %q: %s", bad, prompt)
		}
	}
	// Capability guard: the replacement must keep the advice that still applies
	// on Windows instead of emptying the prompt.
	for _, want := range []string{"write the script to a file", "timeout", "interactive"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("replacement prompt dropped %q: %s", want, prompt)
		}
	}
}

func TestRegistrySchemaDropsSDKPromptAfterOverride(t *testing.T) {
	base := actool.NewBash()
	before := actool.NewRegistry(base).Schemas()[0].Description
	if !strings.Contains(before, "chain steps with") {
		t.Fatalf("SDK schema no longer carries the generic prompt: %q", before)
	}
	after := actool.NewRegistry(OverrideToolPrompt(base, windowsBashPrompt)).Schemas()[0].Description
	if strings.Contains(after, "chain steps with") || strings.Contains(after, "heredoc") {
		t.Fatalf("override must remove the SDK prompt from the rendered schema:\n%s", after)
	}
	if !strings.Contains(after, "Executes a shell command") {
		t.Fatalf("base description must survive the override:\n%s", after)
	}
	if !strings.Contains(after, "write the script to a file") {
		t.Fatalf("replacement guidance must reach the rendered schema:\n%s", after)
	}
}

func TestApplyHostPromptOverridesTouchesOnlyBash(t *testing.T) {
	tools := actool.DefaultTools()
	var otherName, otherBefore string
	for _, t := range tools {
		if t.Name() != "Bash" {
			otherName, otherBefore = t.Name(), actool.NewRegistry(t).Schemas()[0].Description
			break
		}
	}
	if otherName == "" {
		t.Skip("SDK exposes no non-Bash tool to compare against")
	}
	before := len(tools)
	tools = applyHostPromptOverrides(tools)
	if len(tools) != before {
		t.Fatalf("tool count changed: %d -> %d", before, len(tools))
	}
	for _, tool := range tools {
		if tool.Name() != "Bash" {
			continue
		}
		desc := actool.NewRegistry(tool).Schemas()[0].Description
		if strings.Contains(desc, "chain steps with") || strings.Contains(desc, "heredoc") {
			t.Fatalf("Bash still advertises Unix syntax:\n%s", desc)
		}
	}
	for _, tool := range tools {
		if tool.Name() != otherName {
			continue
		}
		if got := actool.NewRegistry(tool).Schemas()[0].Description; got != otherBefore {
			t.Fatalf("%s description changed unexpectedly", otherName)
		}
	}
}

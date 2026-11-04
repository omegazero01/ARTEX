package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestIsLegacyReporterPromptMatchesKnownHashOnly(t *testing.T) {
	if IsLegacyReporterPrompt("") {
		t.Fatal("empty prompt must not be legacy")
	}
	if IsLegacyReporterPrompt(ReporterDefaultPrompt) {
		t.Fatal("current default must not be treated as legacy")
	}
	// 临时把已知哈希指向一段测试文本，验证精确命中才升级。
	old := PreviousContextPromptHashes["reporter"]
	sum := sha256.Sum256([]byte("legacy reporter prompt"))
	PreviousContextPromptHashes["reporter"] = hex.EncodeToString(sum[:])
	t.Cleanup(func() { PreviousContextPromptHashes["reporter"] = old })
	if !IsLegacyReporterPrompt("legacy reporter prompt") {
		t.Fatal("exact old default must be detected")
	}
	if IsLegacyReporterPrompt("legacy reporter prompt ") {
		t.Fatal("modified prompt must not be detected")
	}
}

package server

// Ported from SERIES `server/shell_usage_test.go`. The persisted-attribution case
// uses ARTEX's live-DB convention (ARTEX_PG_DSN) instead of SERIES' modeServer
// fixture; the original TestCustomToolDirectAndDeferredUsage is not shell
// metering and stays out of scope.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	actool "github.com/Autumn-27/norma/tool"

	"github.com/Autumn-27/artex/agent"
	"github.com/Autumn-27/artex/db"
)

func TestShellUsageCommands(t *testing.T) {
	rows := []*db.Tool{{Key: "scan", Executable: "nmap", Kind: "shell", Enabled: true, Agents: []string{"worker"}}, {Key: "fetch", Executable: "/opt/bin/curl", Kind: "shell", Enabled: true, Agents: []string{"worker"}}, {Key: "hidden", Executable: "ffuf", Kind: "shell", Enabled: true, Agents: []string{"planner"}}}
	for _, tt := range []struct {
		command string
		want    []string
	}{
		{"nmap a; nmap b | /opt/bin/curl x", []string{"fetch", "scan"}},
		{"X=1 env -u HOME A=b sudo -u root command -- nmap a", []string{"scan"}},
		{"echo nmap; # nmap a\n echo 'nmap b'; f() { nmap c; }; bash -c 'nmap d'", nil},
		{"$tool a; ${tool} b; $(printf nmap) c", nil},
		{"false && nmap a", []string{"scan"}},
		{"'/opt/bin/curl' x; command -v nmap; ffuf x", []string{"fetch"}},
		{"nmap '$(bad)'", []string{"scan"}},
		{"nmap '", nil},
	} {
		t.Run(tt.command, func(t *testing.T) {
			got, _ := shellUsageKeys(tt.command, rows, "worker")
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
	rows = append(rows, &db.Tool{Key: "duplicate", Executable: "nmap", Kind: "shell", Enabled: true, Agents: []string{"worker"}})
	if got, conflicts := shellUsageKeys("nmap a", rows, "worker"); len(got) != 0 || len(conflicts) != 1 {
		t.Fatal(got, conflicts)
	}
}

func TestShellUsageDirectoryAttribution(t *testing.T) {
	rows := []*db.Tool{
		{Key: "gogo", Executable: "gogo_linux_amd64", Directory: "/root/pentestools", Kind: "shell", Enabled: true, Agents: []string{"worker"}},
		{Key: "elsewhere", Executable: "gogo_linux_amd64", Directory: "/opt/other", Kind: "shell", Enabled: true, Agents: []string{"worker"}},
		{Key: "disabled", Executable: "other", Directory: "/root/pentestools", Kind: "shell", Agents: []string{"worker"}},
		{Key: "unbound", Executable: "extra", Directory: "/root/pentestools", Kind: "shell", Enabled: true, Agents: []string{"planner"}},
	}
	for _, tt := range []struct {
		name, command string
		want          []string
	}{
		{"background example", "cd /root/artex/data/tasks/123/i14278 && /root/pentestools/gogo_linux_amd64 -i 221.2.93.248 -p 1-65535 --http -j -o gogo_full.json --timeout 3 --retry 1 2>&1 | tail -50", []string{"gogo"}},
		{"quoted path", "'/root/pentestools/gogo_linux_amd64' -i target", []string{"gogo"}},
		{"same basename elsewhere", "/opt/other/gogo_linux_amd64 target", []string{"elsewhere"}},
		{"unconfigured program in directory", "/root/pentestools/unknown target", nil},
		{"conditional request", "false && /root/pentestools/gogo_linux_amd64 target", []string{"gogo"}},
		{"repeated in one request", "/root/pentestools/gogo_linux_amd64 a; /root/pentestools/gogo_linux_amd64 b", []string{"gogo"}},
		{"not executable positions", "echo '/root/pentestools/gogo_linux_amd64'; # /root/pentestools/gogo_linux_amd64\nf() { /root/pentestools/gogo_linux_amd64; }; bash -c '/root/pentestools/gogo_linux_amd64'", nil},
		{"disabled and unbound", "/root/pentestools/other; /root/pentestools/extra", nil},
		{"invalid shell", "'/root/pentestools/gogo_linux_amd64", nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, conflicts := shellUsageKeys(tt.command, rows, "worker")
			if !reflect.DeepEqual(got, tt.want) || len(conflicts) != 0 {
				t.Fatalf("got %v, conflicts %v; want %v", got, conflicts, tt.want)
			}
		})
	}
	if got, conflicts := shellUsageKeys("gogo_linux_amd64 target", rows, "worker"); len(got) != 0 || len(conflicts) != 1 {
		t.Fatalf("ambiguous bare command counted: %v, %v", got, conflicts)
	}
	if got, conflicts := shellUsageKeys("gogo_linux_amd64 target", rows[:1], "worker"); !reflect.DeepEqual(got, []string{"gogo"}) || len(conflicts) != 0 {
		t.Fatalf("unique bare command not counted: %v, %v", got, conflicts)
	}
	relative := []*db.Tool{{Key: "scan", Executable: "scan_linux", Directory: "pentestools", Kind: "shell", Enabled: true, Agents: []string{"worker"}}}
	if got, _ := shellUsageKeys("/data/pentestools/scan_linux target", relative, "worker", "/data"); !reflect.DeepEqual(got, []string{"scan"}) {
		t.Fatalf("relative configured directory not resolved: %v", got)
	}

	// A configured directory must not steal another declaration's absolute path.
	rows = append(rows, &db.Tool{Key: "duplicate", Executable: "/root/pentestools/gogo_linux_amd64", Kind: "shell", Enabled: true, Agents: []string{"worker"}})
	if got, conflicts := shellUsageKeys("/root/pentestools/gogo_linux_amd64 target", rows, "worker"); len(got) != 0 || len(conflicts) != 1 {
		t.Fatalf("ambiguous command counted: %v, %v", got, conflicts)
	}
}

// Windows resolves command names case-insensitively; POSIX does not. The fold is
// platform-scoped so the Linux contract (distinct /usr/bin/Curl) is preserved.
func TestShellUsageCaseFolding(t *testing.T) {
	rows := []*db.Tool{
		{Key: "curl_bare", Executable: "curl.exe", Kind: "shell", Enabled: true, Agents: []string{"worker"}},
		{Key: "curl_path", Executable: "/opt/bin/curl", Kind: "shell", Enabled: true, Agents: []string{"worker"}},
	}
	for _, tt := range []struct {
		name, command string
		want          []string
	}{
		{"bare name", "CURL.EXE --version", []string{"curl_bare"}},
		{"absolute path", "/opt/bin/CURL --version", []string{"curl_path"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := shellUsageKeys(tt.command, rows, "worker")
			want := tt.want
			if runtime.GOOS != "windows" {
				want = nil
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v want %v on %s", got, want, runtime.GOOS)
			}
		})
	}
}

func TestShellUsageMeterAndFailure(t *testing.T) {
	recorder := &recordingToolUsage{}
	base := actool.Build(actool.Spec{Name: "Bash", Run: func(context.Context, json.RawMessage, *actool.ToolContext) (actool.Result, error) {
		return actool.Errorf("failed command"), nil
	}})
	ri := agent.RunInfo{TaskID: 7, ExplorationID: 8, IntentID: 9, SessionID: "s"}
	tool := &shellMeteredTool{CoreTool: meterTool(base, recorder, "Bash", "worker", ri), recorder: recorder, declarations: []*db.Tool{{Key: "gogo", Executable: "gogo_linux_amd64", Directory: "/root/pentestools", Kind: "shell", Enabled: true, Agents: []string{"worker"}}}, agentKey: "worker", ri: ri}
	r, e := tool.Call(t.Context(), json.RawMessage(`{"command":"cd /root/artex/data/tasks/123/i14278 && /root/pentestools/gogo_linux_amd64 -i 221.2.93.248 -p 1-65535 --http -j -o gogo_full.json --timeout 3 --retry 1 2>&1 | tail -50; /root/pentestools/gogo_linux_amd64 --version","run_in_background":true}`), nil)
	if e != nil || !r.IsError || len(recorder.rows) != 2 {
		t.Fatal(r, e, recorder.rows)
	}
	if recorder.rows[0].ToolKey != "gogo" || recorder.rows[0].SessionID != "s" || recorder.rows[0].TaskID != 7 || recorder.rows[0].IntentID != 9 || recorder.rows[1].ToolKey != "Bash" {
		t.Fatal(recorder.rows)
	}
}

func TestShellUsageValidationAndLedgerFailure(t *testing.T) {
	recorder := &recordingToolUsage{}
	executed := 0
	base := actool.Build(actool.Spec{Name: "Bash", Run: func(context.Context, json.RawMessage, *actool.ToolContext) (actool.Result, error) {
		executed++
		return actool.Text("ok"), nil
	}})
	tool := &shellMeteredTool{CoreTool: meterTool(base, recorder, "Bash", "worker", agent.RunInfo{}), recorder: recorder, declarations: []*db.Tool{{Key: "gogo", Executable: "gogo_linux_amd64", Directory: "/root/pentestools", Kind: "shell", Enabled: true, Agents: []string{"worker"}}}, agentKey: "worker"}
	// ARTEX 的 DecorateTool 只替换描述/schema 并注入缺省值，不做必填校验
	// （SERIES 在 PrepareInput 里校验；移植缺口记在 PORTING.md 第五节），
	// 所以这里只验证"台账失败不打断执行"。
	validated := agent.DecorateTool(tool, "", map[string]any{"type": "object", "properties": map[string]any{"command": map[string]any{"type": "string"}}})
	recorder.err = errors.New("ledger unavailable")
	if result, err := validated.Call(t.Context(), json.RawMessage(`{"command":"/root/pentestools/gogo_linux_amd64 target"}`), nil); err != nil || result.IsError || executed != 1 {
		t.Fatalf("metering failure interrupted execution: result=%+v err=%v executed=%d", result, err, executed)
	}
}

// Persisted attribution needs a real database; it runs when ARTEX_PG_DSN is set,
// same convention as TestMCPUsageLedgerLive. No task row is required: the ledger
// only stores the attribution columns.
func TestShellUsagePersistedAttribution(t *testing.T) {
	dsn := os.Getenv("ARTEX_PG_DSN")
	if dsn == "" {
		t.Skip("需要 ARTEX_PG_DSN")
	}
	d, err := db.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	// Close after the row cleanup: t.Cleanup is LIFO, and a defer would close
	// the pool before the delete runs (silently leaving rows behind).
	t.Cleanup(func() { d.Close() })
	key := "test_shell_usage_persist"
	sessionID := "shell-usage-persist"
	t.Cleanup(func() {
		_, _ = d.Exec(`DELETE FROM tool_usage WHERE tool_key=$1 OR session_id=$2`, key, sessionID)
		_ = d.DeleteCustomTool(key)
		_, _ = d.Exec(`DELETE FROM tool_usage WHERE tool_key=$1 OR session_id=$2`, key, sessionID)
	})
	_ = d.DeleteCustomTool(key)
	declaration := &db.Tool{Key: key, Executable: "gogo_linux_amd64", Directory: "/root/pentestools", UsageHelp: "gogo -i target", WhenToUse: "端口扫描", Kind: "shell", Enabled: true, Agents: []string{"worker"}}
	if err := d.CreateCustomTool(declaration); err != nil {
		t.Fatal(err)
	}
	stored, err := d.GetTool(key)
	if err != nil || stored == nil {
		t.Fatalf("GetTool=%v err=%v", stored, err)
	}
	if stored.Executable != declaration.Executable || stored.Directory != declaration.Directory || stored.UsageHelp != declaration.UsageHelp || stored.WhenToUse != declaration.WhenToUse {
		t.Fatalf("shell fields not persisted: %+v", stored)
	}
	taskID, _ := strconv.ParseInt("7", 10, 64)
	ri := agent.RunInfo{TaskID: taskID, ExplorationID: 8, SessionID: sessionID}
	base := actool.Build(actool.Spec{Name: "Bash", Run: func(context.Context, json.RawMessage, *actool.ToolContext) (actool.Result, error) {
		return actool.Text("submitted"), nil
	}})
	tool := &shellMeteredTool{CoreTool: meterTool(base, d, "Bash", "worker", ri), recorder: d, declarations: []*db.Tool{declaration}, agentKey: "worker", ri: ri}
	command := fmt.Sprintf(`{"command":"cd /root/artex/data/tasks/%d/i14278 && /root/pentestools/gogo_linux_amd64 -i 221.2.93.248 2>&1 | tail -50","run_in_background":true}`, taskID)
	if result, err := tool.Call(t.Context(), json.RawMessage(command), nil); err != nil || result.IsError {
		t.Fatal(result, err)
	}
	var customCalls, bashCalls int
	if err := d.QueryRow(`SELECT count(*) FROM tool_usage WHERE tool_key=$1 AND session_id=$2 AND agent_key='worker'`, key, sessionID).Scan(&customCalls); err != nil {
		t.Fatal(err)
	}
	if err := d.QueryRow(`SELECT count(*) FROM tool_usage WHERE tool_key='Bash' AND session_id=$1`, sessionID).Scan(&bashCalls); err != nil {
		t.Fatal(err)
	}
	if customCalls != 1 || bashCalls != 1 {
		t.Fatalf("persisted custom/Bash calls = %d/%d, want 1/1", customCalls, bashCalls)
	}
	// The declaration reaches the model through Bash's description: ToolResolve
	// must apply shellToolNote for the bound agent, and non-bound agents must not
	// see it. Restore the global resolver the test replaces.
	oldResolve := agent.ToolResolve
	t.Cleanup(func() { agent.ToolResolve = oldResolve })
	wireTools(d, nil)
	bash := actool.Build(actool.Spec{Name: "Bash", Description: "run", Run: func(context.Context, json.RawMessage, *actool.ToolContext) (actool.Result, error) {
		return actool.Text("ok"), nil
	}})
	resolved := agent.ToolResolve(t.Context(), "worker", []actool.CoreTool{bash})
	var got actool.CoreTool
	for _, t := range resolved {
		if t.Name() == "Bash" {
			got = t
		}
	}
	if got == nil {
		t.Fatalf("Bash missing from resolved tools: %v", resolved)
	}
	desc := got.Description()
	for _, want := range []string{"可执行命令：gogo_linux_amd64", "所在目录：/root/pentestools", "用法帮助：gogo -i target", "何时调用：端口扫描"} {
		if !strings.Contains(desc, want) {
			t.Fatalf("resolved Bash description missing %q: %s", want, desc)
		}
	}

	// The declared guidance reaches the model through Bash's description.
	note := shellToolNote([]*db.Tool{declaration}, "worker")
	for _, want := range []string{"可执行命令：gogo_linux_amd64", "所在目录：/root/pentestools", "用法帮助：gogo -i target", "何时调用：端口扫描"} {
		if !strings.Contains(note, want) {
			t.Fatalf("Bash note missing %q:\n%s", want, note)
		}
	}
}

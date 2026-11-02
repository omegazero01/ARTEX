package server

import (
	"context"
	"encoding/json"
	"log"
	"path"
	"path/filepath"
	"sort"
	"strings"

	actool "github.com/Autumn-27/norma/tool"
	mvdan "mvdan.cc/sh/v3/syntax"

	"github.com/Autumn-27/artex/agent"
	"github.com/Autumn-27/artex/db"
)

// shellPathClean normalizes both separators so a Windows caller's C:\\tools\\x and a
// bash-style /opt/x compare consistently.
func shellPathClean(p string) string {
	return path.Clean(strings.ReplaceAll(p, "\\", "/"))
}

// shellPathAbs recognizes POSIX absolute paths on every OS plus native Windows paths.
func shellPathAbs(p string) bool {
	return strings.HasPrefix(p, "/") || filepath.IsAbs(p)
}

func shellPathBase(p string) string {
	return path.Base(shellPathClean(p))
}

// Counts submitted command requests, not processes or iterations. Never expands
// variables, evaluates substitutions, or inspects scripts / bash -c strings.
func shellCommands(command string) []string {
	f, err := mvdan.NewParser().Parse(strings.NewReader(command), "")
	if err != nil {
		return nil
	}
	names := map[string]bool{}
	mvdan.Walk(f, func(n mvdan.Node) bool {
		switch node := n.(type) {
		case *mvdan.FuncDecl:
			return false
		case *mvdan.CallExpr:
			args := []string{}
			for _, word := range node.Args {
				value, ok := literalShellWord(word)
				if !ok {
					break
				}
				args = append(args, value)
			}
			if cmd := unwrapShellCommand(args); cmd != "" {
				names[cmd] = true
			}
			// Do not inspect substitutions embedded in arguments.
			return false
		}
		return true
	})
	out := []string{}
	for n := range names {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func literalShellWord(w *mvdan.Word) (string, bool) {
	var out strings.Builder
	for _, p := range w.Parts {
		switch v := p.(type) {
		case *mvdan.Lit:
			// Backslash escapes are deliberately not guessed.
			if strings.Contains(v.Value, "\\") {
				return "", false
			}
			out.WriteString(v.Value)
		case *mvdan.SglQuoted:
			out.WriteString(v.Value)
		case *mvdan.DblQuoted:
			value, ok := literalShellWord(&mvdan.Word{Parts: v.Parts})
			if !ok {
				return "", false
			}
			out.WriteString(value)
		default:
			return "", false
		}
	}
	return out.String(), true
}

func unwrapShellCommand(args []string) string {
	for len(args) > 0 {
		original := args[0]
		name := shellPathBase(original)
		args = args[1:]
		if name != "env" && name != "sudo" && name != "command" {
			return original
		}
		for len(args) > 0 {
			a := args[0]
			if a == "--" {
				args = args[1:]
				break
			}
			if name == "env" && strings.Contains(a, "=") && !strings.HasPrefix(a, "-") {
				args = args[1:]
				continue
			}
			if !strings.HasPrefix(a, "-") {
				break
			}
			args = args[1:]
			switch {
			case name == "command" && (a == "-v" || a == "-V"):
				return ""
			case name == "command" && a == "-p":
			case name == "env" && (a == "-i" || a == "--ignore-environment"):
			case name == "sudo" && (a == "-n" || a == "-E" || a == "-H"):
			case name == "env" && (a == "-u" || a == "--unset" || a == "-C" || a == "--chdir"), name == "sudo" && (a == "-u" || a == "-g" || a == "--user" || a == "--group"):
				if len(args) == 0 {
					return ""
				}
				args = args[1:]
			default:
				return ""
			}
		}
		if len(args) > 0 && shellPathBase(args[0]) != "env" && shellPathBase(args[0]) != "sudo" && shellPathBase(args[0]) != "command" {
			return args[0]
		}
	}
	return ""
}

func shellUsageKeys(command string, rows []*db.Tool, agentKey string, configRoot ...string) (keys, ambiguous []string) {
	root := ""
	if len(configRoot) > 0 {
		root = configRoot[0]
	}
	matched := map[string]bool{}
	for _, cmd := range shellCommands(command) {
		candidates := []string{}
		for _, row := range rows {
			if row.System || row.Kind != "shell" || !row.Enabled || !contains(row.Agents, agentKey) {
				continue
			}
			if shellCommandMatches(cmd, row, root) {
				candidates = append(candidates, row.Key)
			}
		}
		if len(candidates) == 1 {
			matched[candidates[0]] = true
		} else if len(candidates) > 1 {
			ambiguous = append(ambiguous, strings.Join(candidates, ", "))
		}
	}
	for k := range matched {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return
}

// A directory narrows absolute-path calls to that exact configured executable.
// A bare command has no reliable resolved path at this layer, so it continues
// to match by name. Relative paths are not resolved against an unobserved cd.
func shellCommandMatches(cmd string, row *db.Tool, configRoot string) bool {
	target := row.ShellCommand()
	if shellPathAbs(target) {
		return shellPathAbs(cmd) && shellPathClean(target) == shellPathClean(cmd)
	}
	if strings.ContainsAny(target, "/\\") {
		return false
	}
	if cmd == target {
		return true
	}
	if !shellPathAbs(cmd) {
		return strings.TrimSpace(row.Directory) == "" && shellPathBase(cmd) == target
	}
	directory := strings.TrimSpace(row.Directory)
	if directory == "" {
		return shellPathBase(cmd) == target
	}
	if !shellPathAbs(directory) && shellPathAbs(configRoot) {
		directory = path.Join(shellPathClean(configRoot), directory)
	}
	return shellPathAbs(directory) && path.Join(shellPathClean(directory), target) == shellPathClean(cmd)
}

type shellMeteredTool struct {
	actool.CoreTool
	recorder     toolUsageRecorder
	declarations []*db.Tool
	agentKey     string
	ri           agent.RunInfo
	configRoot   string
}

func (m *shellMeteredTool) Call(ctx context.Context, in json.RawMessage, tc *actool.ToolContext) (actool.Result, error) {
	var input struct {
		Command string `json:"command"`
	}
	if json.Unmarshal(in, &input) == nil && ctx.Err() == nil {
		keys, conflicts := shellUsageKeys(input.Command, m.declarations, m.agentKey, m.configRoot)
		for _, conflict := range conflicts {
			log.Printf("[toolusage] ambiguous shell declarations: %s", conflict)
		}
		for _, key := range keys {
			if err := m.recorder.InsertToolUsage(&db.ToolUsage{ToolKey: key, AgentKey: m.agentKey, TaskID: m.ri.TaskID, ExplorationID: m.ri.ExplorationID, IntentID: m.ri.IntentID, SessionID: m.ri.SessionID}); err != nil {
				log.Printf("[toolusage] shell %s: %v", key, err)
			}
		}
	}
	return m.CoreTool.Call(ctx, in, tc)
}

package db

import (
	"strings"
	"testing"
)

func TestBuildTaskDSLWhereKeepsPlaceholderOrderAndAuthorization(t *testing.T) {
	where, args, err := buildTaskDSLWhere(42, "domain=example.test", "service", "false", ApprovalApproved)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"type=$2",
		"current_link.task_id=$3",
		"relation.task_id=$3",
		"current_link.tested",
		"source_link.tested",
		"task_asset_effectively_approved($3,assets.id)",
	} {
		if !strings.Contains(where, want) {
			t.Fatalf("where missing %q: %s", want, where)
		}
	}
	if len(args) != 4 || args[1] != "service" || args[2] != int64(42) || args[3] != false {
		t.Fatalf("unexpected args: %#v", args)
	}

	blockedWhere, _, err := buildTaskDSLWhere(42, "domain=example.test", "", "all", "blocked")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(blockedWhere, "task_asset_effective_approval_state($2,assets.id)='blocked'") {
		t.Fatalf("blocked where does not use effective tombstones: %s", blockedWhere)
	}
}

func TestTaskDSLApprovalFields(t *testing.T) {
	for _, field := range []string{"status", "approval_state"} {
		for _, state := range []string{"approved", "pending", "blocked", "revoked"} {
			where, args, err := buildTaskDSLWhere(42, field+"=="+state, "", "all", ApprovalApproved)
			if err != nil || !strings.Contains(where, "task_asset_effective_approval_state($1,assets.id) = $2") || !strings.Contains(where, "task_asset_effectively_approved($3,assets.id)") {
				t.Fatalf("%s=%s: %s %v", field, state, where, err)
			}
			if args[0] != int64(42) || args[1] != state {
				t.Fatalf("args=%v", args)
			}
		}
	}
	for _, query := range []string{"status==200", "status_code==200"} {
		where, _, err := buildTaskDSLWhere(42, query, "", "all", ApprovalApproved)
		if err != nil || !strings.Contains(where, "status_code = $1") {
			t.Fatalf("%s: %s %v", query, where, err)
		}
	}
	for _, query := range []string{"approval_state==unknown", "approval_state>approved", "status>approved"} {
		if _, _, err := buildTaskDSLWhere(42, query, "", "all", ApprovalApproved); err == nil {
			t.Fatal("accepted", query)
		}
	}
	for _, query := range []string{"approval_state==approved", "status==approved", "task_id==42 AND status==approved"} {
		n, err := ParseDSL(query)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := buildDSLWhere(n); err == nil {
			t.Fatal("global query accepted approval filter", query)
		}
	}
	if _, _, err := buildTaskDSLWhere(42, "", "", "all", ApprovalApproved); err != nil {
		t.Fatal(err)
	}
}

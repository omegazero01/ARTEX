package db

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestManualAssetBlockArchiveCompatibility(t *testing.T) {
	dsn := testDSN(t)
	d, err := Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	task, err := d.CreateTask("block archive", "goal", nil, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer d.DeleteTask(task.ID)
	root, err := d.Assets().UpsertRootDomain(UpsertRootDomainReq{Domain: fmt.Sprintf("block-archive-%d.test", task.ID), TaskID: task.ID})
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Assets().BlockTaskAssets(task.ID, []int64{root}, "operator", "keep audit"); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	reopened.Close()
	var raw []byte
	if err := d.QueryRow(`SELECT json_agg(b) FROM task_asset_blocks b WHERE task_id=$1`, task.ID).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	for _, legacy := range []bool{false, true} {
		tx, err := d.Begin()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(`DELETE FROM task_asset_blocks WHERE task_id=$1`, task.ID); err != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		payload := raw
		if legacy {
			var records []map[string]any
			if err := json.Unmarshal(raw, &records); err != nil {
				tx.Rollback()
				t.Fatal(err)
			}
			for _, record := range records {
				delete(record, "block_kind")
			}
			payload, err = json.Marshal(records)
			if err != nil {
				tx.Rollback()
				t.Fatal(err)
			}
		}
		if err := insertArchiveRows(tx, "task_asset_blocks", payload); err != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		var kind string
		if err := tx.QueryRow(`SELECT block_kind FROM task_asset_blocks WHERE task_id=$1`, task.ID).Scan(&kind); err != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		want := "manual"
		if legacy {
			want = "deleted"
		}
		if kind != want {
			tx.Rollback()
			t.Fatalf("kind=%s want %s", kind, want)
		}
		tx.Rollback()
	}
	if err := d.EnsureLLMRecordsTable(); err != nil {
		t.Fatal(err)
	}
	if err := d.EnsureLLMUsageTable(); err != nil {
		t.Fatal(err)
	}
	if err := d.SetPaused(task.ID, true); err != nil {
		t.Fatal(err)
	}
	archive, err := d.QueueTaskArchive(task.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Exec(`DELETE FROM task_archives WHERE id=$1`, archive.ID)
	if _, err := d.ClaimTaskArchiveJob(t.Context()); err != nil {
		t.Fatal(err)
	}
	store := d.Exploration(task.ExplorationID)
	member, err := store.AddNode(KindFact, map[string]any{"summary": "archived cold member"}, 0, "confirmed", "worker", nil)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := store.AddDigest(map[string]any{"body": "archived cold body"}, []int64{member})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`UPDATE explorations SET round_no=42 WHERE id=$1`, task.ExplorationID); err != nil {
		t.Fatal(err)
	}
	snapshot, err := d.SnapshotTaskArchive(task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.CompleteTaskArchive(archive.ID, snapshot, "/tmp/manual-block-test.tar.zst", "test", 100, 50); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`UPDATE explorations SET round_no=0 WHERE id=$1`, task.ExplorationID); err != nil {
		t.Fatal(err)
	}
	if _, err := d.QueueTaskArchiveRestore(archive.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := d.ClaimTaskArchiveJob(t.Context()); err != nil {
		t.Fatal(err)
	}
	if _, err := d.RestoreTaskArchive(archive.ID, snapshot, 0); err != nil {
		t.Fatal(err)
	}
	if round, err := store.RoundNo(); err != nil || round != 42 {
		t.Fatalf("restored round=%d err=%v", round, err)
	}
	if members, err := store.DigestMembers(digest); err != nil || len(members) != 1 || members[0] != member {
		t.Fatalf("restored digest members=%v err=%v", members, err)
	}
	var state, kind string
	if err := d.QueryRow(`SELECT l.approval_state,b.block_kind FROM task_asset_links l JOIN task_asset_blocks b
ON b.task_id=l.task_id AND b.asset_id=l.asset_id WHERE l.task_id=$1 AND l.asset_id=$2`, task.ID, root).Scan(&state, &kind); err != nil {
		t.Fatal(err)
	}
	if state != ApprovalBlocked || kind != "manual" {
		t.Fatalf("archive lost block: %s %s", state, kind)
	}
	if err := d.Assets().ApproveTaskAssets(task.ID, []int64{root}, "user", "restore"); err != nil {
		t.Fatal(err)
	}
}

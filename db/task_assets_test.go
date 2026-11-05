package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestRegisterTaskAssetScopesCreatesAssetsAndPersistsTextScope(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Skipf("postgres unavailable (%v) - skipping", err)
	}
	defer d.Close()

	task, err := d.CreateTask("manual scope registration", "goal", nil, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.DeleteTask(task.ID) })
	suffix := time.Now().UnixNano()
	domain := fmt.Sprintf("manual-scope-%d.example.test", suffix)
	ip := fmt.Sprintf("fc%02x:%x:%x::19", byte(suffix), uint16(suffix>>8), uint16(suffix>>24))
	cidrPrefix := fmt.Sprintf("fd%02x:%x:%x", byte(suffix), uint16(suffix>>16), uint16(suffix>>32))
	cidr := cidrPrefix + "::/64"
	cidrIP := cidrPrefix + "::23"
	icp := fmt.Sprintf("京 ICP 备 %d 号-1", suffix)
	cidrAssetID, err := d.Assets().UpsertIP(UpsertIPReq{IP: cidrIP})
	if err != nil {
		t.Fatal(err)
	}
	icpAssetID, err := d.Assets().UpsertApp(UpsertAppReq{Name: fmt.Sprintf("manual ICP app %d", suffix), ICP: icp})
	if err != nil {
		t.Fatal(err)
	}
	keywordAssetID, err := d.Assets().UpsertApp(UpsertAppReq{Name: fmt.Sprintf("Acme Security %d", suffix)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Assets().AttachAssetsToTask(task.ID, []int64{cidrAssetID}, "prepare tombstone"); err != nil {
		t.Fatal(err)
	}
	if detached, err := d.Assets().DetachAssetFromTask(task.ID, cidrAssetID); err != nil || !detached {
		t.Fatalf("prepare cidr tombstone: detached=%v err=%v", detached, err)
	}
	inputs := []ScopeInput{
		{Kind: "domain", Value: domain},
		{Kind: "ip", Value: ip},
		{Kind: "cidr", Value: cidr},
		{Kind: "icp", Value: " " + icp + " "},
		{Kind: "keyword", Value: fmt.Sprintf(" Acme   Security   %d ", suffix)},
	}

	first, err := d.Assets().RegisterTaskAssetScopes(task.ID, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if first.Requested != 5 || first.AssetsLinked != 2 || first.AssetsExisting != 0 || first.ScopesAdded != 5 || first.ScopesExisting != 0 {
		t.Fatalf("unexpected first mutation: %+v", first)
	}
	assets, err := d.Assets().QueryByTask(task.ID, "", 10, 0)
	if err != nil || len(assets) != 4 {
		t.Fatalf("task assets=%+v err=%v", assets, err)
	}
	assetIDs := make([]int64, 0, len(assets))
	linked := make(map[int64]bool, len(assets))
	for _, asset := range assets {
		assetIDs = append(assetIDs, asset.ID)
		linked[asset.ID] = true
		if asset.TaskSource != "manual" || asset.TaskSourceSummary != manualTaskScopeSummary {
			t.Fatalf("unexpected task asset provenance: %+v", asset)
		}
		if asset.ApprovalState != ApprovalApproved {
			t.Fatalf("manual scope asset was not approved: %+v", asset)
		}
	}
	if !linked[cidrAssetID] || !linked[icpAssetID] || linked[keywordAssetID] {
		t.Fatalf("scope authorization mismatch: cidr=%v icp=%v keyword=%v", linked[cidrAssetID], linked[icpAssetID], linked[keywordAssetID])
	}
	var cidrBlocks int
	if err := d.QueryRow(`SELECT count(*) FROM task_asset_blocks WHERE task_id=$1 AND asset_id=$2`, task.ID, cidrAssetID).Scan(&cidrBlocks); err != nil {
		t.Fatal(err)
	}
	if cidrBlocks != 0 {
		t.Fatalf("manual CIDR registration left %d tombstones", cidrBlocks)
	}
	assetIDs = append(assetIDs, keywordAssetID)
	t.Cleanup(func() { _, _ = d.Assets().DeleteByIDs(assetIDs) })
	scopes, err := d.Assets().ListTaskScope(task.ID)
	if err != nil || len(scopes) != 5 {
		t.Fatalf("task scope=%+v err=%v", scopes, err)
	}
	values := map[string]string{}
	for _, scope := range scopes {
		values[scope.Kind] = scope.Value
	}
	if values["icp"] != NormalizeICP(inputs[3].Value) || values["keyword"] != strings.TrimSpace(inputs[4].Value) {
		t.Fatalf("text scope not normalized: %+v", values)
	}

	second, err := d.Assets().RegisterTaskAssetScopes(task.ID, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if second.AssetsExisting != 2 || second.AssetsLinked != 0 || second.ScopesExisting != 5 || second.ScopesAdded != 0 {
		t.Fatalf("unexpected idempotent mutation: %+v", second)
	}

	rollbackDomain := fmt.Sprintf("rollback-%d.example.test", suffix)
	_, err = d.Assets().RegisterTaskAssetScopes(task.ID, []ScopeInput{
		{Kind: "domain", Value: rollbackDomain},
		{Kind: "keyword", Value: strings.Repeat("x", MaxCompanyScopeRawRunes+1)},
	})
	if !errors.Is(err, ErrTaskAssetInvalid) {
		t.Fatalf("oversized scope error=%v", err)
	}
	var rollbackAssets int
	if err := d.QueryRow(`SELECT count(*) FROM assets WHERE type='root_domain' AND domain=$1`, rollbackDomain).Scan(&rollbackAssets); err != nil {
		t.Fatal(err)
	}
	if rollbackAssets != 0 {
		t.Fatalf("invalid request created %d assets", rollbackAssets)
	}
}

func TestTaskAssetAttachDetachPreservesGlobalAssetAndAnchors(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Skipf("postgres unavailable (%v) - skipping", err)
	}
	defer d.Close()

	task, err := d.CreateTask("asset editing", "goal", nil, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.DeleteTask(task.ID) })
	assetID, err := d.Assets().UpsertRootDomain(UpsertRootDomainReq{Domain: fmt.Sprintf("manual-%d.example.test", time.Now().UnixNano())})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = d.Assets().DeleteByIDs([]int64{assetID}) })

	mutation, err := d.Assets().AttachAssetsToTask(task.ID, []int64{assetID, assetID}, "授权资产清单第 3 项")
	if err != nil {
		t.Fatal(err)
	}
	if mutation.Requested != 1 || mutation.Attached != 1 || mutation.Existing != 0 {
		t.Fatalf("unexpected first mutation: %+v", mutation)
	}
	assets, err := d.Assets().QueryByTask(task.ID, "", 10, 0)
	if err != nil || len(assets) != 1 {
		t.Fatalf("task assets=%+v err=%v", assets, err)
	}
	if assets[0].TaskSource != "manual" || assets[0].TaskSourceSummary != "授权资产清单第 3 项" {
		t.Fatalf("unexpected provenance: %+v", assets[0])
	}

	intentID, err := d.Exploration(task.ExplorationID).AddIntent(map[string]any{"summary": "test asset"}, 5, []int64{assetID}, "human")
	if err != nil {
		t.Fatal(err)
	}
	detached, err := d.Assets().DetachAssetFromTask(task.ID, assetID)
	if err != nil || !detached {
		t.Fatalf("detach=%v err=%v", detached, err)
	}
	if assets, err := d.Assets().QueryByTask(task.ID, "", 10, 0); err != nil || len(assets) != 1 || !assets[0].Blocked {
		t.Fatalf("detached task assets must retain a visible tombstone: assets=%+v err=%v", assets, err)
	}
	var global, anchors, links int
	_ = d.QueryRow(`SELECT count(*) FROM assets WHERE id=$1`, assetID).Scan(&global)
	_ = d.QueryRow(`SELECT count(*) FROM exploration_anchors WHERE node_id=$1 AND asset_id=$2`, intentID, assetID).Scan(&anchors)
	_ = d.QueryRow(`SELECT count(*) FROM task_asset_links WHERE task_id=$1 AND asset_id=$2`, task.ID, assetID).Scan(&links)
	if global != 1 || anchors != 1 || links != 0 {
		t.Fatalf("global=%d anchors=%d links=%d", global, anchors, links)
	}
}

func assetListContains(assets []*Asset, id int64) bool {
	for _, asset := range assets {
		if asset.ID == id {
			return true
		}
	}
	return false
}

func TestAssetKeyIsStableForNonHTTPServiceAndApp(t *testing.T) {
	port := 22
	service := &Asset{Type: "service", IP: "2001:0db8::7", Port: &port, ServiceName: " SSH "}
	if key, host := AssetKey(service); key != "service:2001:db8::7:22:ssh" || host != "2001:db8::7" {
		t.Fatalf("service AssetKey=(%q,%q)", key, host)
	}
	app := &Asset{Type: "app", BundleID: " COM.Example.App ", AppName: "ignored"}
	if key, _ := AssetKey(app); key != "app:com.example.app" {
		t.Fatalf("app AssetKey=%q", key)
	}
}

func TestAgentRediscoveryCanReattachServiceAfterRecordDeletion(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Skipf("postgres unavailable (%v) - skipping", err)
	}
	defer d.Close()

	suffix := time.Now().UnixNano()
	task, err := d.CreateTask(fmt.Sprintf("recreated tombstone %d", suffix), "goal", nil, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	domain := fmt.Sprintf("recreated-%d.example.test", suffix)
	firstID, err := d.Assets().UpsertOtherService(UpsertOtherServiceReq{Domain: domain, Port: 2222, ServiceName: "ssh"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = d.DeleteTask(task.ID)
		_, _ = d.Exec(`DELETE FROM assets WHERE type='service' AND domain=$1 AND port=2222`, domain)
	})
	if _, err := d.Assets().AttachAssetsToTask(task.ID, []int64{firstID}, "operator selected"); err != nil {
		t.Fatal(err)
	}
	if detached, err := d.Assets().DetachAssetFromTask(task.ID, firstID); err != nil || !detached {
		t.Fatalf("detach=%v err=%v", detached, err)
	}
	if _, err := d.Assets().DeleteByIDs([]int64{firstID}); err != nil {
		t.Fatal(err)
	}

	recreatedID, err := d.Assets().UpsertOtherService(UpsertOtherServiceReq{
		Domain: domain, Port: 2222, ServiceName: "ssh", TaskID: task.ID, AgentDiscovered: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if recreatedID == firstID {
		t.Fatal("global asset was not recreated with a new id")
	}
	state, err := d.Assets().RegisterAgentDiscoveredAsset(task.ID, recreatedID, "worker")
	if err != nil || state != ApprovalApproved {
		t.Fatalf("rediscovery state=%q err=%v", state, err)
	}
	var attached, linked bool
	if err := d.QueryRow(`SELECT $1=ANY(task_ids) FROM assets WHERE id=$2`, task.ID, recreatedID).Scan(&attached); err != nil {
		t.Fatal(err)
	}
	if err := d.QueryRow(`SELECT EXISTS(SELECT 1 FROM task_asset_links WHERE task_id=$1 AND asset_id=$2)`, task.ID, recreatedID).Scan(&linked); err != nil {
		t.Fatal(err)
	}
	if !attached || !linked {
		t.Fatalf("service rediscovery failed to attach: task_ids=%v link=%v", attached, linked)
	}
}

func TestIntentAssetsHonorsCanceledReadContext(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Skipf("postgres unavailable (%v) - skipping", err)
	}
	defer d.Close()
	task, err := d.CreateTask("canceled intent assets", "goal", nil, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.DeleteTask(task.ID) })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = d.Assets().WithReadContext(ctx).IntentAssets(task.ID)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("IntentAssets error=%v, want context.Canceled", err)
	}
}

func TestIntentAssetsPageLimitsIntentsBeforeJoiningAssets(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Skipf("postgres unavailable (%v) - skipping", err)
	}
	defer d.Close()
	task, err := d.CreateTask("paged intent assets", "goal", nil, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	assetID, err := d.Assets().UpsertRootDomain(UpsertRootDomainReq{
		Domain: fmt.Sprintf("intent-page-%d.example.test", time.Now().UnixNano()),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = d.DeleteTask(task.ID)
		_, _ = d.Assets().DeleteByIDs([]int64{assetID})
	})
	if _, err := d.Assets().AttachAssetsToTask(task.ID, []int64{assetID}, "test target"); err != nil {
		t.Fatal(err)
	}
	intentIDs := make([]int64, 0, 3)
	for i := 0; i < 3; i++ {
		intentID, err := d.Exploration(task.ExplorationID).AddIntent(
			map[string]any{"summary": fmt.Sprintf("intent %d", i)}, 1, []int64{assetID}, "planner",
		)
		if err != nil {
			t.Fatal(err)
		}
		intentIDs = append(intentIDs, intentID)
	}
	page, err := d.Assets().IntentAssetsPage(task.ID, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page) != 2 || page[0].IntentID != intentIDs[2] || page[1].IntentID != intentIDs[1] {
		t.Fatalf("newest intent asset page=%+v, want intent ids %v", page, intentIDs[1:])
	}
	page, err = d.Assets().IntentAssetsPage(task.ID, intentIDs[1], 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page) != 1 || page[0].IntentID != intentIDs[0] {
		t.Fatalf("older intent asset page=%+v, want intent id %d", page, intentIDs[0])
	}
}

func TestTaskCreationDirectAssetsHaveIndependentTestState(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Skipf("postgres unavailable (%v) - skipping", err)
	}
	defer d.Close()

	domain := fmt.Sprintf("direct-task-%d.example.test", time.Now().UnixNano())
	assetID, err := d.Assets().UpsertRootDomain(UpsertRootDomainReq{Domain: domain})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = d.Assets().DeleteByIDs([]int64{assetID}) })
	first, err := d.CreateTaskWithOptions("direct first", "goal", TaskCreateOptions{AssetIDs: []int64{assetID, assetID}})
	if err != nil {
		t.Fatal(err)
	}
	second, err := d.CreateTaskWithOptions("direct second", "goal", TaskCreateOptions{AssetIDs: []int64{assetID}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.DeleteTask(first.ID); _ = d.DeleteTask(second.ID) })

	page, err := d.Assets().QueryByTaskTested(first.ID, "root_domain", "false", 10, 0)
	if err != nil || len(page) != 1 || page[0].Tested == nil || *page[0].Tested || page[0].TaskSource != "direct" {
		t.Fatalf("initial direct task asset=%+v err=%v", page, err)
	}
	if err := d.Assets().MarkTaskAssetsTested(first.ID, []int64{assetID}, "worker"); err != nil {
		t.Fatal(err)
	}
	page, err = d.Assets().QueryByTaskTested(first.ID, "root_domain", "true", 10, 0)
	if err != nil || len(page) != 1 || page[0].Tested == nil || !*page[0].Tested || page[0].TestedBy != "worker" || page[0].TestedAt == nil {
		t.Fatalf("marked direct task asset=%+v err=%v", page, err)
	}
	page, err = d.Assets().QueryByTaskTested(second.ID, "root_domain", "false", 10, 0)
	if err != nil || len(page) != 1 || page[0].Tested == nil || *page[0].Tested {
		t.Fatalf("second task state leaked=%+v err=%v", page, err)
	}
	scopes, err := d.Assets().ListTaskScope(first.ID)
	if err != nil || len(scopes) != 1 || scopes[0].Kind != "root_domain" || scopes[0].Domain != domain {
		t.Fatalf("direct asset scope=%+v err=%v", scopes, err)
	}
}

func TestTaskCreationMissingAssetRollsBack(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Skipf("postgres unavailable (%v) - skipping", err)
	}
	defer d.Close()
	description := fmt.Sprintf("missing-direct-asset-%d", time.Now().UnixNano())
	if _, err := d.CreateTaskWithOptions(description, "goal", TaskCreateOptions{AssetIDs: []int64{1 << 62}}); !errors.Is(err, ErrTaskAssetAssetNotFound) {
		t.Fatalf("missing asset error=%v", err)
	}
	var tasks, explorations int
	if err := d.QueryRow(`SELECT count(*) FROM tasks WHERE description=$1`, description).Scan(&tasks); err != nil {
		t.Fatal(err)
	}
	if err := d.QueryRow(`SELECT count(*) FROM explorations WHERE description=$1`, description).Scan(&explorations); err != nil {
		t.Fatal(err)
	}
	if tasks != 0 || explorations != 0 {
		t.Fatalf("failed create leaked rows: tasks=%d explorations=%d", tasks, explorations)
	}
}

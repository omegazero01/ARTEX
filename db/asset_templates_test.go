package db

import (
	"testing"
)

func TestAgentAssetValidationAndUserURL(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	task, err := d.CreateTaskWithOptions("validation", "goal", TaskCreateOptions{AssetApprovalTemplate: "explicit_targets"})
	if err != nil {
		t.Fatal(err)
	}
	defer d.DeleteTask(task.ID)
	as := d.Assets()
	for _, host := range []string{"www_host", "${HOST}", "foo..example.test", "-bad.example.test", "host"} {
		if _, err := as.UpsertRootDomain(UpsertRootDomainReq{Domain: host, TaskID: task.ID, AgentDiscovered: true}); err == nil {
			t.Fatalf("accepted %q", host)
		}
	}
	for _, u := range []string{"https://www_host/x", "https://${HOST}/", "http://foo..test/"} {
		if _, err := as.UpsertHTTPService(UpsertHTTPServiceReq{URL: u, TaskID: task.ID, AgentDiscovered: true}); err == nil {
			t.Fatalf("accepted URL %q", u)
		}
	}
	var count int
	if err := d.QueryRow(`SELECT count(*) FROM task_asset_links WHERE task_id=$1`, task.ID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("invalid writes leaked: %d %v", count, err)
	}
	id, err := as.RegisterUserAsset(task.ID, func(scoped *AssetStore) (int64, error) {
		return scoped.UpsertHTTPService(UpsertHTTPServiceReq{URL: "https://www.manual-url.test/path", TaskID: task.ID})
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := as.ValidateTaskAssetsApproved(task.ID, []int64{id}); err != nil {
		t.Fatal(err)
	}
	if _, err := as.RegisterAgentDiscoveredAsset(task.ID, id, "worker"); err != nil {
		t.Fatal(err)
	}
	if err := as.ValidateTaskAssetsApproved(task.ID, []int64{id}); err != nil {
		t.Fatalf("user URL downgraded: %v", err)
	}
	if err := as.ValidateTaskHostsApproved(task.ID, []string{"www.manual-url.test"}); err != nil {
		t.Fatalf("tool layer blocked exact user host: %v", err)
	}
	apiID, err := as.RegisterUserAssetWithSource(task.ID, "api", "API exact provenance", func(scoped *AssetStore) (int64, error) {
		return scoped.UpsertRootDomain(UpsertRootDomainReq{Domain: "api-provenance.test", TaskID: task.ID})
	})
	if err != nil {
		t.Fatal(err)
	}
	var source, summary string
	if err := d.QueryRow(`SELECT source,source_summary FROM task_asset_links WHERE task_id=$1 AND asset_id=$2`, task.ID, apiID).Scan(&source, &summary); err != nil {
		t.Fatal(err)
	}
	if source != "api" || summary != "API exact provenance" {
		t.Fatalf("lost user entry provenance: source=%q summary=%q", source, summary)
	}
}

func TestTemplateDefaultsAndDescriptionEvidence(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	task, err := d.CreateTask("测试 https://target.evidence.test；禁止测试 excluded.evidence.test", "goal", nil, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer d.DeleteTask(task.ID)
	if task.AssetApprovalTemplate != "all_assets" {
		t.Fatal("wrong new default")
	}
	for _, entry := range []struct{ host, evidence string }{{"fake.evidence.test", "测试 https://target.evidence.test"}, {"evidence.test", "测试 https://target.evidence.test"}, {"excluded.evidence.test", "禁止测试 excluded.evidence.test"}} {
		if _, err := d.Assets().RegisterDescriptionAsset(task.ID, "host", entry.host, entry.evidence); err == nil {
			t.Fatalf("accepted ungrounded target %+v", entry)
		}
	}
	if _, err := d.Assets().RegisterDescriptionAsset(task.ID, "host", "excluded.evidence.test", "excluded.evidence.test"); err == nil {
		t.Fatal("partial quote bypassed denial")
	}
	if err := d.SetAssetApprovalTemplate(task.ID, "explicit_targets"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`UPDATE tasks SET first_run_at=now() WHERE id=$1`, task.ID); err != nil {
		t.Fatal(err)
	}
	if err := d.SetAssetApprovalTemplate(task.ID, "all_assets"); err == nil {
		t.Fatal("changed template after start")
	}
}

func TestDNSOwnerMetadataNotHost(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	task, err := d.CreateTask("DNS metadata", "goal", nil, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer d.DeleteTask(task.ID)
	id, err := d.Assets().UpsertSubdomain(UpsertSubdomainReq{Domain: "_ldap._tcp.dns-record.test", RecordType: "SRV", RecordValue: []string{"0 100 389 ldap.dns-record.test"}, TaskID: task.ID, AgentDiscovered: true})
	if err != nil {
		t.Fatal(err)
	}
	var domain string
	var count int
	if err := d.QueryRow(`SELECT domain,jsonb_array_length(extra->'dns_records') FROM assets WHERE id=$1`, id).Scan(&domain, &count); err != nil {
		t.Fatal(err)
	}
	if domain != "dns-record.test" || count != 1 {
		t.Fatalf("invalid DNS projection %s %d", domain, count)
	}
	if err := d.QueryRow(`SELECT count(*) FROM assets WHERE domain='_ldap._tcp.dns-record.test'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("DNS owner became a host: %d %v", count, err)
	}
}

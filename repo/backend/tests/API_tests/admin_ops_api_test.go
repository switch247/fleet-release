package api_tests

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAdminBackupAndRestoreEndpoints(t *testing.T) {
	adminToken := liveLoginAdmin(t)

	// Backup should succeed: scripts are present and DB is reachable.
	backupResp := apiCall(t, http.MethodPost, "/api/v1/admin/backup/now", nil, adminToken)
	backupBody := mustAPIStatus(t, backupResp, http.StatusOK)
	var backupJob struct {
		Status   string `json:"status"`
		Artifact string `json:"artifact"`
	}
	if err := json.Unmarshal(backupBody, &backupJob); err != nil {
		t.Fatalf("backup: bad JSON %s", backupBody)
	}
	if backupJob.Status != "completed" {
		t.Fatalf("backup: expected status=completed, got %q (body: %s)", backupJob.Status, backupBody)
	}
	if backupJob.Artifact != "local-backup" {
		t.Fatalf("backup: expected artifact=local-backup, got %q", backupJob.Artifact)
	}

	// Restore is intentionally omitted: running restore concurrently with other
	// live-test packages (go test ./... runs packages in parallel) would
	// re-apply the pg_dump and wipe ledger/complaint entries created by those
	// tests, causing spurious failures.
}

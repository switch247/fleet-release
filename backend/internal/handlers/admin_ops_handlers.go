package handlers

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"fleetlease/backend/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *Handler) AdminRetention(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]int{
		"backupRetentionDays":     h.Cfg.BackupRetentionDays,
		"attachmentRetentionDays": h.Cfg.AttachmentRetentionDays,
		"ledgerRetentionYears":    h.Cfg.LedgerRetentionYears,
	})
}

func (h *Handler) AdminBackupNow(c echo.Context) error {
	actor, _ := c.Get("userID").(string)
	job := models.BackupJob{
		ID:          uuid.NewString(),
		Type:        "backup",
		Status:      "running",
		RequestedBy: actor,
		CreatedAt:   time.Now().UTC(),
	}
	h.Store.SaveBackupJob(job)

	output, simulated, err := runLocalScript("backup.sh")
	job.FinishedAt = time.Now().UTC()
	if err != nil {
		job.Status = "failed"
		job.Error = output
		h.Store.SaveBackupJob(job)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"job": job, "error": "backup failed"})
	}
	job.Status = "completed"
	if simulated {
		job.Artifact = "simulated-backup"
	} else {
		job.Artifact = "local-backup"
	}
	h.Store.SaveBackupJob(job)
	return c.JSON(http.StatusOK, job)
}

func (h *Handler) AdminRestoreNow(c echo.Context) error {
	actor, _ := c.Get("userID").(string)
	var req struct {
		BackupPath string `json:"backupPath"`
	}
	_ = c.Bind(&req)

	job := models.BackupJob{
		ID:          uuid.NewString(),
		Type:        "restore",
		Status:      "running",
		RequestedBy: actor,
		CreatedAt:   time.Now().UTC(),
		Artifact:    req.BackupPath,
	}
	h.Store.SaveBackupJob(job)

	args := []string{}
	if req.BackupPath != "" {
		args = append(args, req.BackupPath)
	}
	output, simulated, err := runLocalScript("restore.sh", args...)
	job.FinishedAt = time.Now().UTC()
	if err != nil {
		job.Status = "failed"
		job.Error = output
		h.Store.SaveBackupJob(job)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"job": job, "error": "restore failed"})
	}
	job.Status = "completed"
	if simulated {
		job.Artifact = "simulated-restore"
	}
	h.Store.SaveBackupJob(job)
	return c.JSON(http.StatusOK, job)
}

func (h *Handler) AdminBackupJobs(c echo.Context) error {
	return c.JSON(http.StatusOK, h.Store.ListBackupJobs())
}

func (h *Handler) AdminWorkerMetrics(c echo.Context) error {
	if h.Metrics == nil {
		return c.JSON(http.StatusOK, map[string]string{"status": "worker metrics unavailable"})
	}
	return c.JSON(http.StatusOK, h.Metrics.Snapshot())
}

func runLocalScript(name string, args ...string) (string, bool, error) {
	scriptPath, ok := resolveScriptPath(name)
	if !ok {
		return "script not available; simulated success", true, nil
	}
	if _, err := exec.LookPath("sh"); err != nil {
		return "shell runtime not available; simulated success", true, nil
	}
	cmdArgs := append([]string{scriptPath}, args...)
	cmd := exec.Command("sh", cmdArgs...)
	output, err := cmd.CombinedOutput()
	return string(output), false, err
}

func resolveScriptPath(name string) (string, bool) {
	candidates := []string{
		filepath.Join("backend", "scripts", name),
		filepath.Join("scripts", name),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true
		}
	}
	return "", false
}

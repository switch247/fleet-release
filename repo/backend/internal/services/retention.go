package services

import (
	"log/slog"
	"os"
	"time"

	"fleetlease/backend/internal/config"
	"fleetlease/backend/internal/models"
	"fleetlease/backend/internal/store"

	"github.com/google/uuid"
)

type RetentionPurgeResult struct {
	ID                 string    `json:"id"`
	AttachmentsDeleted int       `json:"attachmentsDeleted"`
	LedgerDeleted      int       `json:"ledgerDeleted"`
	FileDeleteErrors   int       `json:"fileDeleteErrors"`
	CreatedAt          time.Time `json:"createdAt"`
}

func RunRetentionPurge(st store.Repository, cfg config.Config, logger *slog.Logger) RetentionPurgeResult {
	now := time.Now().UTC()
	attachmentCutoff := now.AddDate(0, 0, -cfg.AttachmentRetentionDays)
	ledgerCutoff := now.AddDate(-cfg.LedgerRetentionYears, 0, 0)

	removedAttachments := st.PurgeAttachmentsOlderThan(attachmentCutoff)
	fileErrors := 0
	for _, attachment := range removedAttachments {
		if attachment.Path == "" {
			continue
		}
		if err := os.Remove(attachment.Path); err != nil && !os.IsNotExist(err) {
			fileErrors++
			logger.Warn("retention_attachment_file_delete_failed", "attachmentID", attachment.ID, "path", attachment.Path, "error", err)
		}
	}
	ledgerDeleted := st.PurgeLedgerOlderThan(ledgerCutoff)

	result := RetentionPurgeResult{
		ID:                 uuid.NewString(),
		AttachmentsDeleted: len(removedAttachments),
		LedgerDeleted:      ledgerDeleted,
		FileDeleteErrors:   fileErrors,
		CreatedAt:          now,
	}
	st.SaveRetentionReport(models.RetentionReport{
		ID:                 result.ID,
		AttachmentsDeleted: result.AttachmentsDeleted,
		LedgerDeleted:      result.LedgerDeleted,
		FileDeleteErrors:   result.FileDeleteErrors,
		CreatedAt:          result.CreatedAt,
	})
	logger.Info(
		"retention_purge_completed",
		"attachmentsDeleted", result.AttachmentsDeleted,
		"ledgerDeleted", result.LedgerDeleted,
		"fileDeleteErrors", result.FileDeleteErrors,
		"attachmentCutoff", attachmentCutoff,
		"ledgerCutoff", ledgerCutoff,
	)
	return result
}

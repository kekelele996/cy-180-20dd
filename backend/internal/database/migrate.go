package database

import (
	"fmt"

	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/model"
	"gorm.io/gorm"
)

// Migrate 执行表结构迁移。
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.User{},
		&model.Project{},
		&model.Question{},
		&model.Recording{},
		&model.TimelineMarker{},
		&model.AuditLog{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	if err := backfillReviewStatus(db); err != nil {
		return fmt.Errorf("backfill recording review status: %w", err)
	}
	return nil
}

// backfillReviewStatus 兼容审核流程上线前已写入摘要的录音：
// 这些摘要此前没有确认过程，视为已通过版本，继续出现在项目时间线。
func backfillReviewStatus(db *gorm.DB) error {
	return db.Model(&model.Recording{}).
		Where("review_status = ? AND summary <> ?", constants.ReviewStatusDraft, "").
		Update("review_status", constants.ReviewStatusApproved).Error
}

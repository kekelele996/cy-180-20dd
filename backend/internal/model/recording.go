package model

import "time"

// Recording 录音片段实体，audio_key 指向 MinIO 对象，status 承载录音状态机流转，
// review_status / pending_summary / reject_reason 承载摘要审核状态机。
//
// 摘要采用「双版本」设计：
//   - Summary 为已通过版本，始终对外（项目时间线）展示；
//   - PendingSummary 为采访员提交、等待档案员处理的待审版本；
//   - 待审期间 Summary 仍保持上一版，审核通过后才切换为待审版本。
type Recording struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	ProjectID       uint       `gorm:"index;not null" json:"project_id"`
	QuestionID      uint       `gorm:"index;not null" json:"question_id"`
	AudioKey        string     `gorm:"size:255" json:"audio_key"`
	DurationSeconds int        `gorm:"not null;default:0" json:"duration_seconds"`
	Summary         string     `gorm:"size:512" json:"summary"`
	PendingSummary  string     `gorm:"size:512" json:"pending_summary"`
	Status          string     `gorm:"size:32;not null;default:recording" json:"status"`
	ReviewStatus    string     `gorm:"size:32;not null;default:draft" json:"review_status"`
	RejectReason    string     `gorm:"size:512" json:"reject_reason"`
	ReviewedBy      uint       `gorm:"default:0" json:"reviewed_by"`
	ReviewedAt      *time.Time `json:"reviewed_at"`
	CreatedBy       uint       `gorm:"not null" json:"created_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// TableName 指定表名。
func (Recording) TableName() string { return "recordings" }

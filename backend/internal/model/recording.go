package model

import "time"

// Recording 录音片段实体，audio_key 指向 MinIO 对象，status 承载录制状态机流转。
// Summary 只保存「已通过」的摘要版本（项目时间线展示它）；
// 新提交先落在 PendingSummary，由 ReviewStatus 承载 待审/已通过/已退回 三阶段，
// 退回原因记录在 RejectReason，待审期间 Summary 仍指向上一版。
type Recording struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ProjectID       uint      `gorm:"index;not null" json:"project_id"`
	QuestionID      uint      `gorm:"index;not null" json:"question_id"`
	AudioKey        string    `gorm:"size:255" json:"audio_key"`
	DurationSeconds int       `gorm:"not null;default:0" json:"duration_seconds"`
	Summary         string    `gorm:"size:512" json:"summary"`
	ReviewStatus    string    `gorm:"size:32;not null;default:unsubmitted" json:"review_status"`
	PendingSummary  string    `gorm:"size:512" json:"pending_summary"`
	RejectReason    string    `gorm:"size:512" json:"reject_reason"`
	Status          string    `gorm:"size:32;not null;default:recording" json:"status"`
	CreatedBy       uint      `gorm:"not null" json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名。
func (Recording) TableName() string { return "recordings" }

package dto

// CreateRecordingRequest 创建录音记录请求。
type CreateRecordingRequest struct {
	ProjectID       uint `json:"project_id" binding:"required"`
	QuestionID      uint `json:"question_id" binding:"required"`
	DurationSeconds int  `json:"duration_seconds" binding:"omitempty,min=0"`
}

// UpdateRecordingRequest 更新录音信息请求（只允许更新时长/录制状态；摘要必须走审核流程）。
type UpdateRecordingRequest struct {
	DurationSeconds int    `json:"duration_seconds" binding:"omitempty,min=0"`
	Status          string `json:"status" binding:"omitempty,oneof=recording processing ready failed"`
}

// SubmitSummaryRequest 采访员提交/补充后重新提交一句话摘要。
type SubmitSummaryRequest struct {
	Summary string `json:"summary" binding:"required,max=512"`
}

// RejectSummaryRequest 档案员退回摘要并填写原因。
type RejectSummaryRequest struct {
	Reason string `json:"reason" binding:"required,max=512"`
}

// RecordingResponse 录音响应。
type RecordingResponse struct {
	ID              uint   `json:"id"`
	ProjectID       uint   `json:"project_id"`
	QuestionID      uint   `json:"question_id"`
	AudioKey        string `json:"audio_key"`
	DurationSeconds int    `json:"duration_seconds"`
	Summary         string `json:"summary"`
	ReviewStatus    string `json:"review_status"`
	PendingSummary  string `json:"pending_summary"`
	RejectReason    string `json:"reject_reason"`
	Status          string `json:"status"`
	CreatedBy       uint   `json:"created_by"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

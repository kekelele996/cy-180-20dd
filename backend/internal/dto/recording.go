package dto

// CreateRecordingRequest 创建录音记录请求。
type CreateRecordingRequest struct {
	ProjectID       uint `json:"project_id" binding:"required"`
	QuestionID      uint `json:"question_id" binding:"required"`
	DurationSeconds int  `json:"duration_seconds" binding:"omitempty,min=0"`
}

// UpdateRecordingRequest 更新录音信息请求（仅录音本身的时长/状态，摘要不走此接口）。
type UpdateRecordingRequest struct {
	DurationSeconds int    `json:"duration_seconds" binding:"omitempty,min=0"`
	Status          string `json:"status" binding:"omitempty,oneof=recording processing ready failed"`
}

// SubmitSummaryRequest 采访员提交（或重新提交）摘要审核请求。
type SubmitSummaryRequest struct {
	Summary string `json:"summary" binding:"required,max=512"`
}

// ReviewSummaryRequest 档案员审核摘要请求：action=approve 批准，action=reject 退回（需填写原因）。
type ReviewSummaryRequest struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
	Reason string `json:"reason" binding:"omitempty,max=512"`
}

// RecordingResponse 录音响应。
type RecordingResponse struct {
	ID              uint   `json:"id"`
	ProjectID       uint   `json:"project_id"`
	QuestionID      uint   `json:"question_id"`
	AudioKey        string `json:"audio_key"`
	DurationSeconds int    `json:"duration_seconds"`
	Summary         string `json:"summary"`
	PendingSummary  string `json:"pending_summary"`
	Status          string `json:"status"`
	ReviewStatus    string `json:"review_status"`
	RejectReason    string `json:"reject_reason"`
	ReviewedBy      uint   `json:"reviewed_by"`
	ReviewedAt      string `json:"reviewed_at"`
	CreatedBy       uint   `json:"created_by"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

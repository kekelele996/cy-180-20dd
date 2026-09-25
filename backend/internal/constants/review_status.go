// 录音摘要审核状态机枚举。
package constants

// 摘要审核状态定义：每段录音的摘要都要经过「待审 → 已通过 / 已退回」的审核流程。
const (
	ReviewStatusDraft     = "draft"     // 未提交：采访员尚未提交摘要
	ReviewStatusPending   = "pending"   // 待审：采访员已提交，等待档案员处理
	ReviewStatusApproved  = "approved"  // 已通过：档案员批准，摘要进入项目时间线
	ReviewStatusRejected  = "rejected"  // 已退回：档案员退回并附原因，采访员修改后重新提交
)

// ValidReviewStatus 校验摘要审核状态是否合法。
func ValidReviewStatus(status string) bool {
	switch status {
	case ReviewStatusDraft, ReviewStatusPending, ReviewStatusApproved, ReviewStatusRejected:
		return true
	default:
		return false
	}
}

// CanSubmitReview 返回采访员在当前状态下是否可以提交（重新提交）摘要审核。
// 待审期间不允许重复提交；未提交、已通过（修改上一版）、已退回（补充后重提）均可提交。
func CanSubmitReview(status string) bool {
	switch status {
	case ReviewStatusDraft, ReviewStatusApproved, ReviewStatusRejected:
		return true
	default:
		return false
	}
}

// CanReview 返回档案员在当前状态下是否可以审核（批准/退回）。
func CanReview(status string) bool {
	return status == ReviewStatusPending
}

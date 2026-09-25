// 录音摘要审核状态机枚举（与录制状态 RecordingStatus 相互独立）。
package constants

// 摘要审核状态定义。
const (
	ReviewStatusUnsubmitted = "unsubmitted" // 未提交：录音刚产生，采访员还没写摘要
	ReviewStatusPending     = "pending"     // 待审：采访员已提交，等待档案员处理
	ReviewStatusApproved    = "approved"    // 已通过：摘要进入项目时间线
	ReviewStatusRejected    = "rejected"    // 已退回：档案员填写了退回原因，采访员可补充后重新提交
)

// ValidReviewStatus 校验摘要审核状态是否合法。
func ValidReviewStatus(status string) bool {
	switch status {
	case ReviewStatusUnsubmitted, ReviewStatusPending, ReviewStatusApproved, ReviewStatusRejected:
		return true
	default:
		return false
	}
}

// CanSubmitReview 判断当前审核状态下是否允许采访员提交/重新提交摘要。
// 待审期间不允许重复提交，避免覆盖档案员正在处理的版本；其余阶段均可提交
// （已通过后再次提交会以新版本进入待审，时间线继续显示上一版）。
func CanSubmitReview(current string) bool {
	return current != ReviewStatusPending
}

// CanReview 判断当前审核状态下是否允许档案员处理（批准/退回）。
func CanReview(current string) bool {
	return current == ReviewStatusPending
}

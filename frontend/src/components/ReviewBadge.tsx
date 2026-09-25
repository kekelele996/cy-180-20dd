// 摘要审核状态徽标：待审 / 已通过 / 已退回（未提交不显示）。
import { REVIEW_STATUS_TEXT, REVIEW_STATUS_UNSUBMITTED } from '../constants'

const REVIEW_STYLES: Record<string, string> = {
  pending: 'badge-review-pending',
  approved: 'badge-review-approved',
  rejected: 'badge-review-rejected',
}

export default function ReviewBadge({ status }: { status: string }) {
  if (!status || status === REVIEW_STATUS_UNSUBMITTED) return null
  return <span className={`status-badge ${REVIEW_STYLES[status] || 'badge-default'}`}>{REVIEW_STATUS_TEXT[status] || status}</span>
}

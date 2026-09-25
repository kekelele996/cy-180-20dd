// 通用状态徽标组件，跨页面复用。
import { PROJECT_STATUS_TEXT, RECORDING_STATUS_TEXT, REVIEW_STATUS_TEXT } from '../constants'

interface StatusBadgeProps {
  status: string
  type?: 'project' | 'recording' | 'review'
}

const STYLES: Record<string, string> = {
  draft: 'badge-draft',
  in_progress: 'badge-progress',
  completed: 'badge-completed',
  archived: 'badge-archived',
  recording: 'badge-recording',
  processing: 'badge-processing',
  ready: 'badge-ready',
  failed: 'badge-failed',
  // 摘要审核状态
  pending: 'badge-pending',
  approved: 'badge-approved',
  rejected: 'badge-rejected',
}

const TEXT_MAP = {
  project: PROJECT_STATUS_TEXT,
  recording: RECORDING_STATUS_TEXT,
  review: REVIEW_STATUS_TEXT,
}

export default function StatusBadge({ status, type = 'project' }: StatusBadgeProps) {
  const text = TEXT_MAP[type][status]
  return <span className={`status-badge ${STYLES[status] || 'badge-default'}`}>{text || status}</span>
}

// 录音摘要审核区块：
// - interviewer：撰写/补充摘要并提交审核，查看当前阶段、待审版本与退回原因；
// - archivist：查看待审摘要，批准或填写原因退回。
// 项目时间线始终只展示已通过版本（recording.summary），本组件负责审核过程中的双版本呈现。
import { useState } from 'react'
import StatusBadge from './StatusBadge'
import {
  REVIEW_STATUS_APPROVED,
  REVIEW_STATUS_DRAFT,
  REVIEW_STATUS_PENDING,
  REVIEW_STATUS_REJECTED,
  ROLE_ARCHIVIST,
  ROLE_INTERVIEWER,
} from '../constants'
import { useRecordingStore } from '../stores/recordingStore'
import { formatDateTime } from '../utils/format'
import type { Recording } from '../api/types'

type ViewerRole = typeof ROLE_ARCHIVIST | typeof ROLE_INTERVIEWER | 'admin'

interface Props {
  recording: Recording
  viewerRole: ViewerRole
  onChanged?: (recording: Recording) => void
  onNotice?: (message: string, isError?: boolean) => void
}

export default function SummaryReview({ recording, viewerRole, onChanged, onNotice }: Props) {
  const { submitSummary, reviewSummary } = useRecordingStore()
  const [draft, setDraft] = useState('')
  const [rejectReason, setRejectReason] = useState('')
  const [showReject, setShowReject] = useState(false)
  const [busy, setBusy] = useState(false)

  const isInterviewer = viewerRole === ROLE_INTERVIEWER || viewerRole === 'admin'
  const isArchivist = viewerRole === ROLE_ARCHIVIST || viewerRole === 'admin'

  const run = async (fn: () => Promise<Recording>, success: string) => {
    setBusy(true)
    try {
      const updated = await fn()
      setDraft('')
      setRejectReason('')
      setShowReject(false)
      onChanged?.(updated)
      onNotice?.(success, false)
    } catch (e) {
      onNotice?.(e instanceof Error ? e.message : '操作失败', true)
    } finally {
      setBusy(false)
    }
  }

  const pending = recording.review_status === REVIEW_STATUS_PENDING

  return (
    <div>
      <div className="review-row">
        <span className="muted">审核阶段：</span>
        <StatusBadge status={recording.review_status} type="review" />
        {recording.reviewed_at && (
          <span className="muted">最近处理：{formatDateTime(recording.reviewed_at)}</span>
        )}
      </div>

      {/* 已通过版本：项目时间线展示的就是这一版 */}
      <div className="review-box approved">
        <span className="review-label">已通过摘要（时间线版本）：</span>
        {recording.summary || <span className="muted">暂无已通过摘要</span>}
      </div>

      {/* 待审版本：采访员提交后、档案员处理前保存于此 */}
      {recording.pending_summary && (
        <div className="review-box pending">
          <span className="review-label">待审摘要：</span>
          {recording.pending_summary}
        </div>
      )}

      {/* 退回原因 */}
      {recording.review_status === REVIEW_STATUS_REJECTED && recording.reject_reason && (
        <div className="review-box rejected">
          <span className="review-label">退回原因：</span>
          {recording.reject_reason}
        </div>
      )}

      {/* 采访员：撰写并提交审核（待审期间不可重复提交） */}
      {isInterviewer && (
        <>
          <div className="summary-edit">
            <input
              value={draft}
              placeholder={
                recording.review_status === REVIEW_STATUS_DRAFT
                  ? '写一句话摘要，提交后进入审核'
                  : recording.review_status === REVIEW_STATUS_REJECTED
                    ? '根据退回原因补充修改后重新提交'
                    : pending
                      ? '档案员审核中，暂不能提交新版本'
                      : '基于已通过版本修订并重新提交'
              }
              onChange={(e) => setDraft(e.target.value)}
              disabled={pending}
            />
            <button
              className="btn btn-primary btn-small"
              disabled={busy || pending || !draft.trim()}
              onClick={() => run(() => submitSummary(recording.id, draft.trim()), '摘要已提交，等待档案员审核')}
            >
              {recording.review_status === REVIEW_STATUS_DRAFT ? '提交审核' : '重新提交'}
            </button>
          </div>
          {recording.review_status === REVIEW_STATUS_REJECTED && !draft && (
            <div className="muted" style={{ fontSize: 12, marginTop: 4 }}>
              提示：待审版本保留在上方，可对照退回原因修改。
            </div>
          )}
        </>
      )}

      {/* 档案员：对待审版本批准或退回 */}
      {isArchivist && pending && (
        <>
          <div className="review-actions">
            <button
              className="btn btn-primary btn-small"
              disabled={busy}
              onClick={() => run(() => reviewSummary(recording.id, { action: 'approve' }), '摘要已通过')}
            >
              ✓ 批准
            </button>
            <button
              className="btn btn-danger btn-small"
              disabled={busy}
              onClick={() => setShowReject((v) => !v)}
            >
              ✗ 退回
            </button>
          </div>
          {showReject && (
            <div className="reject-form">
              <input
                value={rejectReason}
                placeholder="填写退回原因，采访员将据此补充修改"
                onChange={(e) => setRejectReason(e.target.value)}
              />
              <button
                className="btn btn-danger btn-small"
                disabled={busy || !rejectReason.trim()}
                onClick={() =>
                  run(
                    () => reviewSummary(recording.id, { action: 'reject', reason: rejectReason.trim() }),
                    '摘要已退回',
                  )
                }
              >
                确认退回
              </button>
            </div>
          )}
        </>
      )}
      {isArchivist && !pending && recording.review_status === REVIEW_STATUS_APPROVED && (
        <div className="muted" style={{ fontSize: 12, marginTop: 6 }}>
          该摘要已通过；采访员重新提交新版本后将再次进入待审。
        </div>
      )}
    </div>
  )
}

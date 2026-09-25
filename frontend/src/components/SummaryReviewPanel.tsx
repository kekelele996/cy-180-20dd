// 采访员侧的摘要审核面板：
// 展示当前审核阶段、时间线版本、待审版本、退回原因；
// 未提交时写摘要并提交，退回后补充内容重新提交，已通过后可提交修改版本。
import { useEffect, useState } from 'react'
import ReviewBadge from './ReviewBadge'
import {
  REVIEW_STATUS_APPROVED,
  REVIEW_STATUS_PENDING,
  REVIEW_STATUS_REJECTED,
  REVIEW_STATUS_UNSUBMITTED,
} from '../constants'
import { useRecordingStore } from '../stores/recordingStore'
import type { Recording } from '../api/types'

interface Props {
  recording: Recording
  onMessage: (msg: string, isError?: boolean) => void
  // 审核状态变化后通知父组件刷新录音列表（父组件持有本地列表副本）。
  onChanged?: () => void | Promise<void>
}

export default function SummaryReviewPanel({ recording, onMessage, onChanged }: Props) {
  const submitSummary = useRecordingStore((s) => s.submitSummary)
  const [draft, setDraft] = useState('')
  const [editing, setEditing] = useState(false)
  const [saving, setSaving] = useState(false)

  // 未提交：空白输入框直接可写；已退回：以退回的待审版本为草稿继续补充。
  useEffect(() => {
    if (recording.review_status === REVIEW_STATUS_UNSUBMITTED) {
      setDraft('')
      setEditing(true)
    } else if (recording.review_status === REVIEW_STATUS_REJECTED) {
      setDraft(recording.pending_summary || '')
      setEditing(true)
    } else {
      setDraft('')
      setEditing(false)
    }
  }, [recording.id, recording.review_status, recording.pending_summary])

  const submit = async () => {
    const text = draft.trim()
    if (!text || saving) return
    setSaving(true)
    try {
      await submitSummary(recording.id, text)
      setEditing(false)
      onMessage('摘要已提交，等待档案员审核')
      await onChanged?.()
    } catch (e) {
      onMessage(e instanceof Error ? e.message : '摘要提交失败', true)
    } finally {
      setSaving(false)
    }
  }

  const startRevise = () => {
    setDraft(recording.summary || recording.pending_summary || '')
    setEditing(true)
  }

  return (
    <div className="summary-review">
      <div className="summary-review-row">
        <span className="summary-review-label">摘要审核：</span>
        <ReviewBadge status={recording.review_status} />
      </div>

      {recording.summary && (
        <div className="summary-review-row">
          <span className="summary-review-label">时间线版本：</span>
          <span>{recording.summary}</span>
        </div>
      )}

      {recording.review_status === REVIEW_STATUS_PENDING && (
        <>
          <div className="summary-review-row">
            <span className="summary-review-label">待审版本：</span>
            <span className="summary-pending-text">{recording.pending_summary}</span>
          </div>
          <div className="muted">档案员审核期间暂不能修改，处理结果出来后会在此显示。</div>
        </>
      )}

      {recording.review_status === REVIEW_STATUS_REJECTED && (
        <div className="summary-review-row">
          <span className="summary-review-label">退回原因：</span>
          <span className="summary-reject-reason">{recording.reject_reason || '（档案员未填写原因）'}</span>
        </div>
      )}

      {editing ? (
        <div className="summary-review-actions">
          <input
            value={draft}
            maxLength={512}
            placeholder="写一句话摘要，提交后进入审核"
            onChange={(e) => setDraft(e.target.value)}
          />
          <button className="btn btn-primary btn-small" disabled={!draft.trim() || saving} onClick={submit}>
            {saving ? '提交中…' : recording.review_status === REVIEW_STATUS_REJECTED ? '重新提交' : '提交审核'}
          </button>
          {recording.review_status === REVIEW_STATUS_APPROVED && (
            <button className="btn btn-plain btn-small" onClick={() => setEditing(false)}>
              取消
            </button>
          )}
        </div>
      ) : (
        recording.review_status === REVIEW_STATUS_APPROVED && (
          <div className="summary-review-actions">
            <button className="btn btn-plain btn-small" onClick={startRevise}>
              提交修改版本
            </button>
          </div>
        )
      )}

      {recording.review_status === REVIEW_STATUS_UNSUBMITTED && !recording.summary && (
        <div className="muted">摘要经档案员审核通过后才会出现在项目时间线。</div>
      )}
    </div>
  )
}

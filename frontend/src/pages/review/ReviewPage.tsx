// 档案员审核工作台：按项目/审核阶段查看录音，批准或填写原因退回摘要。
// 已通过的摘要保留在项目时间线；待审期间时间线仍显示上一版，处理后才切换。
import { useCallback, useEffect, useMemo, useState } from 'react'
import AudioPlayer from '../../components/AudioPlayer'
import EmptyState from '../../components/EmptyState'
import ReviewBadge from '../../components/ReviewBadge'
import { useAuth } from '../../hooks/useAuth'
import { listQuestions } from '../../api/question'
import { listRecordings } from '../../api/recording'
import { useProjectStore } from '../../stores/projectStore'
import { useRecordingStore } from '../../stores/recordingStore'
import { REVIEW_STATUS_OPTIONS, REVIEW_STATUS_PENDING, ROLE_ADMIN, ROLE_ARCHIVIST } from '../../constants'
import { formatDuration } from '../../utils/format'
import type { Question, Recording } from '../../api/types'

export default function ReviewPage() {
  useAuth([ROLE_ARCHIVIST, ROLE_ADMIN])
  const { projects, fetchList: fetchProjects } = useProjectStore()
  const approveSummary = useRecordingStore((s) => s.approveSummary)
  const rejectSummary = useRecordingStore((s) => s.rejectSummary)
  const [projectId, setProjectId] = useState(0)
  const [reviewStatus, setReviewStatus] = useState<Recording['review_status']>(REVIEW_STATUS_PENDING)
  const [recordings, setRecordings] = useState<Recording[]>([])
  const [questionsByProject, setQuestionsByProject] = useState<Record<number, Question[]>>({})
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')
  const [messageError, setMessageError] = useState(false)

  const notify = useCallback((msg: string, isError = false) => {
    setMessage(msg)
    setMessageError(isError)
    setTimeout(() => setMessage(''), 4000)
  }, [])

  const reload = useCallback(async () => {
    setLoading(true)
    try {
      const res = await listRecordings({
        project_id: projectId || undefined,
        review_status: reviewStatus || '',
      })
      setRecordings(res.list)
    } catch (e) {
      notify(e instanceof Error ? e.message : '录音列表加载失败', true)
    } finally {
      setLoading(false)
    }
  }, [projectId, reviewStatus])

  useEffect(() => {
    fetchProjects({ page: 1, page_size: 100 })
  }, [fetchProjects])

  useEffect(() => {
    reload()
  }, [reload])

  // 按项目拉取问题清单并本地缓存，用于展示每段录音关联的问题。
  const projectIds = useMemo(() => Array.from(new Set(recordings.map((r) => r.project_id))), [recordings])
  useEffect(() => {
    projectIds.forEach(async (pid) => {
      if (questionsByProject[pid]) return
      try {
        const res = await listQuestions(pid)
        setQuestionsByProject((prev) => ({ ...prev, [pid]: res.list }))
      } catch (e) {
        console.error('fetch questions failed', e)
      }
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [projectIds.join(',')])

  const projectOf = (id: number) => projects.find((p) => p.id === id)
  const questionOf = (r: Recording) =>
    (questionsByProject[r.project_id] || []).find((q) => q.id === r.question_id)

  const handleApprove = async (r: Recording) => {
    try {
      await approveSummary(r.id)
      notify(`录音 #${r.id} 摘要已通过`)
      reload()
    } catch (e) {
      notify(e instanceof Error ? e.message : '审核失败', true)
    }
  }

  const handleReject = async (r: Recording, reason: string) => {
    try {
      await rejectSummary(r.id, reason)
      notify(`录音 #${r.id} 摘要已退回`)
      reload()
    } catch (e) {
      notify(e instanceof Error ? e.message : '退回失败', true)
    }
  }

  return (
    <div className="page">
      <div className="page-header">
        <h2>摘要审核工作台</h2>
      </div>
      {message && <div className={`toast ${messageError ? 'error' : 'success'}`}>{message}</div>}

      <div className="filter-bar">
        <select value={projectId} onChange={(e) => setProjectId(Number(e.target.value))}>
          <option value={0}>全部项目</option>
          {projects.map((p) => (
            <option key={p.id} value={p.id}>
              {p.title}（{p.interviewee_name}）
            </option>
          ))}
        </select>
        <select value={reviewStatus} onChange={(e) => setReviewStatus(e.target.value as Recording['review_status'])}>
          {REVIEW_STATUS_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
        <button className="btn btn-plain btn-small" onClick={reload}>
          刷新
        </button>
        <span className="filter-count">共 {recordings.length} 段录音</span>
      </div>

      {loading ? (
        <div className="muted">加载中…</div>
      ) : recordings.length === 0 ? (
        <EmptyState title="没有符合条件的录音" description="切换项目或审核阶段筛选试试" />
      ) : (
        <div className="recording-list">
          {recordings.map((r) => (
            <ReviewRow
              key={r.id}
              recording={r}
              projectTitle={projectOf(r.project_id)?.title || `项目 #${r.project_id}`}
              questionContent={questionOf(r)?.content || `问题 #${r.question_id}`}
              onApprove={() => handleApprove(r)}
              onReject={(reason) => handleReject(r, reason)}
            />
          ))}
        </div>
      )}
    </div>
  )
}

function ReviewRow({
  recording,
  projectTitle,
  questionContent,
  onApprove,
  onReject,
}: {
  recording: Recording
  projectTitle: string
  questionContent: string
  onApprove: () => void
  onReject: (reason: string) => void
}) {
  const [reason, setReason] = useState(recording.reject_reason || '')
  const [showReject, setShowReject] = useState(false)
  const pending = recording.review_status === REVIEW_STATUS_PENDING

  return (
    <div className="recording-row">
      <div className="recording-meta">
        <ReviewBadge status={recording.review_status} />
        <span className="muted">{projectTitle}</span>
        <span className="muted">{formatDuration(recording.duration_seconds)}</span>
      </div>
      <div className="muted">关联问题：{questionContent}</div>
      <AudioPlayer recordingId={recording.id} durationSeconds={recording.duration_seconds} />

      {recording.summary && (
        <div className="summary-review-row">
          <span className="summary-review-label">时间线版本（上一版）：</span>
          <span>{recording.summary}</span>
        </div>
      )}
      {recording.pending_summary && (
        <div className="summary-review-row">
          <span className="summary-review-label">待审摘要：</span>
          <span className="summary-pending-text">{recording.pending_summary}</span>
        </div>
      )}
      {recording.review_status === 'rejected' && recording.reject_reason && (
        <div className="summary-review-row">
          <span className="summary-review-label">上次退回原因：</span>
          <span className="summary-reject-reason">{recording.reject_reason}</span>
        </div>
      )}

      {pending && (
        <div className="summary-review-actions">
          <button className="btn btn-primary btn-small" onClick={onApprove}>
            ✓ 批准（切换到时间线）
          </button>
          <button className="btn btn-danger btn-small" onClick={() => setShowReject((v) => !v)}>
            ✗ 退回补充
          </button>
        </div>
      )}
      {pending && showReject && (
        <div className="summary-review-actions">
          <input
            value={reason}
            maxLength={512}
            placeholder="填写退回原因，如：摘要缺少时间地点"
            onChange={(e) => setReason(e.target.value)}
          />
          <button
            className="btn btn-danger btn-small"
            disabled={!reason.trim()}
            onClick={() => {
              onReject(reason.trim())
              setShowReject(false)
            }}
          >
            确认退回
          </button>
        </div>
      )}
    </div>
  )
}

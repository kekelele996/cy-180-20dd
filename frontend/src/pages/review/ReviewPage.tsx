// 摘要审核台（档案员）：选择项目 → 查看每段录音的当前审核阶段与待审摘要 → 批准或退回。
import { useCallback, useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import AudioPlayer from '../../components/AudioPlayer'
import EmptyState from '../../components/EmptyState'
import StatusBadge from '../../components/StatusBadge'
import SummaryReview from '../../components/SummaryReview'
import { useAuth } from '../../hooks/useAuth'
import { useProjectStore } from '../../stores/projectStore'
import { useQuestionStore } from '../../stores/questionStore'
import { useRecordingStore } from '../../stores/recordingStore'
import { useAuthStore } from '../../stores/authStore'
import {
  REVIEW_STATUS_APPROVED,
  REVIEW_STATUS_DRAFT,
  REVIEW_STATUS_PENDING,
  REVIEW_STATUS_REJECTED,
  ROLE_ADMIN,
  ROLE_ARCHIVIST,
} from '../../constants'
import { formatDuration } from '../../utils/format'
import type { Recording } from '../../api/types'

const FILTERS = [
  { value: '', label: '全部' },
  { value: REVIEW_STATUS_PENDING, label: '待审' },
  { value: REVIEW_STATUS_REJECTED, label: '已退回' },
  { value: REVIEW_STATUS_APPROVED, label: '已通过' },
  { value: REVIEW_STATUS_DRAFT, label: '未提交' },
] as const

export default function ReviewPage() {
  useAuth([ROLE_ARCHIVIST, ROLE_ADMIN])
  const [params, setParams] = useSearchParams()
  const selectedProject = Number(params.get('project_id')) || 0
  const [filter, setFilter] = useState<string>(REVIEW_STATUS_PENDING)
  const [notice, setNotice] = useState<{ text: string; error: boolean }>({ text: '', error: false })

  const { projects, fetchList } = useProjectStore()
  const { questions, fetchByProject } = useQuestionStore()
  const { recordings, fetchByProject: fetchRecordings } = useRecordingStore()
  const user = useAuthStore((s) => s.user)

  useEffect(() => {
    fetchList({ page: 1, page_size: 100 })
  }, [fetchList])

  useEffect(() => {
    if (selectedProject) {
      fetchByProject(selectedProject)
      fetchRecordings(selectedProject)
    }
  }, [selectedProject, fetchByProject, fetchRecordings])

  const chooseProject = (projectId: number) => {
    const next = new URLSearchParams(params)
    if (projectId) next.set('project_id', String(projectId))
    else next.delete('project_id')
    setParams(next)
  }

  const notify = useCallback((text: string, error = false) => {
    setNotice({ text, error })
    setTimeout(() => setNotice({ text: '', error: false }), 4000)
  }, [])

  const pendingCount = useMemo(
    () => recordings.filter((r) => r.review_status === REVIEW_STATUS_PENDING).length,
    [recordings],
  )

  const visible = filter ? recordings.filter((r) => r.review_status === filter) : recordings

  return (
    <div className="page">
      <div className="page-header">
        <h2>摘要审核台</h2>
        {selectedProject > 0 && <StatusBadge status={pendingCount > 0 ? REVIEW_STATUS_PENDING : REVIEW_STATUS_APPROVED} type="review" />}
      </div>
      {notice.text && <div className={`toast ${notice.error ? 'error' : 'success'}`}>{notice.text}</div>}

      <section className="card">
        <div className="card-title">选择采访项目</div>
        <select value={selectedProject} onChange={(e) => chooseProject(Number(e.target.value))}>
          <option value={0}>请选择项目</option>
          {projects.map((p) => (
            <option key={p.id} value={p.id}>
              {p.title}（{p.interviewee_name}）
            </option>
          ))}
        </select>
      </section>

      {selectedProject === 0 ? (
        <EmptyState title="请先选择采访项目" description="选择项目后即可逐段审核采访员提交的摘要" />
      ) : (
        <section className="card">
          <div className="card-title">录音摘要列表</div>
          <div className="filter-bar">
            {FILTERS.map((f) => (
              <button
                key={f.value}
                className={`btn btn-small ${filter === f.value ? 'btn-primary' : 'btn-plain'}`}
                onClick={() => setFilter(f.value)}
              >
                {f.label}
                {f.value === REVIEW_STATUS_PENDING && pendingCount > 0 ? `（${pendingCount}）` : ''}
              </button>
            ))}
          </div>

          {visible.length === 0 ? (
            <EmptyState
              title="没有符合条件的录音"
              description={
                filter === REVIEW_STATUS_PENDING ? '当前没有待审摘要，采访员提交后会出现在这里' : '切换筛选条件查看其他录音'
              }
            />
          ) : (
            <div className="recording-list">
              {visible.map((r) => (
                <ReviewRow
                  key={r.id}
                  recording={r}
                  question={questions.find((q) => q.id === r.question_id)?.content || `问题 #${r.question_id}`}
                  viewerRole={(user?.role as 'archivist' | 'admin') || 'archivist'}
                  onChanged={() => fetchRecordings(selectedProject)}
                  onNotice={notify}
                />
              ))}
            </div>
          )}
        </section>
      )}
    </div>
  )
}

function ReviewRow({
  recording,
  question,
  viewerRole,
  onChanged,
  onNotice,
}: {
  recording: Recording
  question: string
  viewerRole: 'archivist' | 'admin'
  onChanged: () => void
  onNotice: (text: string, error?: boolean) => void
}) {
  return (
    <div className="recording-row">
      <div className="recording-meta">
        <StatusBadge status={recording.status} type="recording" />
        <span className="muted">{formatDuration(recording.duration_seconds)}</span>
        <span>问题：{question}</span>
      </div>
      <AudioPlayer recordingId={recording.id} durationSeconds={recording.duration_seconds} />
      <SummaryReview recording={recording} viewerRole={viewerRole} onChanged={onChanged} onNotice={onNotice} />
    </div>
  )
}

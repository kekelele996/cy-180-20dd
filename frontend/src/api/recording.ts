import { del, get, post, put, upload } from '../utils/request'
import type { Recording } from './types'

export function listRecordings(params: {
  project_id?: number
  question_id?: number
  review_status?: Recording['review_status'] | ''
}) {
  return get<{ list: Recording[] }>('/recordings', params)
}

export function createRecording(payload: {
  project_id: number
  question_id: number
  duration_seconds?: number
}) {
  return post<Recording>('/recordings', payload)
}

export function getRecording(id: number) {
  return get<Recording>(`/recordings/${id}`)
}

export function updateRecording(id: number, payload: { duration_seconds?: number; status?: string }) {
  return put<Recording>(`/recordings/${id}`, payload)
}

// 提交/重新提交摘要，进入待审。
export function submitRecordingSummary(id: number, summary: string) {
  return put<Recording>(`/recordings/${id}/summary/submit`, { summary })
}

// 档案员批准待审摘要。
export function approveRecordingSummary(id: number) {
  return put<Recording>(`/recordings/${id}/summary/approve`)
}

// 档案员退回待审摘要并填写原因。
export function rejectRecordingSummary(id: number, reason: string) {
  return put<Recording>(`/recordings/${id}/summary/reject`, { reason })
}

export function uploadRecordingAudio(id: number, file: Blob, durationSeconds: number, onProgress?: (p: number) => void) {
  const form = new FormData()
  form.append('file', file, `recording_${id}.webm`)
  form.append('duration_seconds', String(durationSeconds))
  return upload<Recording>(`/recordings/${id}/audio`, form, onProgress)
}

export function deleteRecording(id: number) {
  return del<null>(`/recordings/${id}`)
}

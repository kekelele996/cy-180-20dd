package service

import (
	"log/slog"
	"testing"

	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/dto"
	"github.com/oralhistory/oralhistory/internal/model"
	"github.com/oralhistory/oralhistory/internal/repository"
)

type fakeRecordingRepo struct {
	recordings map[uint]*model.Recording
	err        error
}

func (f *fakeRecordingRepo) Create(r *model.Recording) error {
	if f.err != nil {
		return f.err
	}
	r.ID = uint(len(f.recordings) + 1)
	f.recordings[r.ID] = r
	return nil
}
func (f *fakeRecordingRepo) FindByID(id uint) (*model.Recording, error) {
	if r, ok := f.recordings[id]; ok {
		return r, nil
	}
	return nil, repository.ErrNotFound
}
func (f *fakeRecordingRepo) List(projectID, questionID uint, reviewStatus string) ([]model.Recording, error) {
	out := make([]model.Recording, 0)
	for _, r := range f.recordings {
		if projectID > 0 && r.ProjectID != projectID {
			continue
		}
		if questionID > 0 && r.QuestionID != questionID {
			continue
		}
		if reviewStatus != "" && r.ReviewStatus != reviewStatus {
			continue
		}
		out = append(out, *r)
	}
	return out, nil
}
func (f *fakeRecordingRepo) ListByProject(projectID uint) ([]model.Recording, error) {
	return f.List(projectID, 0, "")
}
func (f *fakeRecordingRepo) ListByQuestion(questionID uint) ([]model.Recording, error) {
	return f.List(0, questionID, "")
}
func (f *fakeRecordingRepo) FindByIDForUpdate(id uint) (*model.Recording, error) {
	return f.FindByID(id)
}
func (f *fakeRecordingRepo) Update(r *model.Recording) error {
	if f.err != nil {
		return f.err
	}
	f.recordings[r.ID] = r
	return nil
}
func (f *fakeRecordingRepo) UpdateStatus(r *model.Recording) error { return f.Update(r) }
func (f *fakeRecordingRepo) Delete(id uint) error {
	delete(f.recordings, id)
	return nil
}
func (f *fakeRecordingRepo) CountByProject(projectID uint) (int64, error) {
	var n int64
	for _, r := range f.recordings {
		if r.ProjectID == projectID {
			n++
		}
	}
	return n, nil
}

type stubProjectRepo struct{ exists bool }

func (s stubProjectRepo) Create(*model.Project) error { return nil }
func (s stubProjectRepo) FindByID(uint) (*model.Project, error) {
	if !s.exists {
		return nil, repository.ErrNotFound
	}
	return &model.Project{ID: 1}, nil
}
func (s stubProjectRepo) FindByIDForUpdate(id uint) (*model.Project, error)     { return s.FindByID(id) }
func (s stubProjectRepo) List(int, int, string) ([]model.Project, int64, error) { return nil, 0, nil }
func (s stubProjectRepo) ListByUser(uint, int, int) ([]model.Project, int64, error) {
	return nil, 0, nil
}
func (s stubProjectRepo) Update(*model.Project) error       { return nil }
func (s stubProjectRepo) UpdateStatus(*model.Project) error { return nil }
func (s stubProjectRepo) Delete(uint) error                 { return nil }
func (s stubProjectRepo) Count() (int64, error)             { return 0, nil }

type stubQuestionRepo struct{}

func (stubQuestionRepo) Create(*model.Question) error { return nil }
func (stubQuestionRepo) FindByID(uint) (*model.Question, error) {
	return &model.Question{ID: 1, ProjectID: 1}, nil
}
func (stubQuestionRepo) ListByProject(uint) ([]model.Question, error) { return nil, nil }
func (stubQuestionRepo) Update(*model.Question) error                 { return nil }
func (stubQuestionRepo) Delete(uint) error                            { return nil }
func (stubQuestionRepo) CountByProject(uint) (int64, error)           { return 0, nil }

func newTestRecordingService() (RecordingService, *fakeRecordingRepo) {
	repo := &fakeRecordingRepo{recordings: map[uint]*model.Recording{}}
	svc := NewRecordingService(repo, stubProjectRepo{exists: true}, stubQuestionRepo{}, slog.Default())
	return svc, repo
}

func TestSummaryReviewFullLifecycle(t *testing.T) {
	svc, repo := newTestRecordingService()
	interviewer := &model.User{ID: 1, Username: "iv", Role: constants.RoleInterviewer}
	archivist := &model.User{ID: 2, Username: "ar", Role: constants.RoleArchivist}

	created, err := svc.Create(interviewer, &dto.CreateRecordingRequest{ProjectID: 1, QuestionID: 1})
	if err != nil {
		t.Fatalf("create recording: %v", err)
	}
	if created.ReviewStatus != constants.ReviewStatusUnsubmitted {
		t.Fatalf("new recording review_status = %s, want unsubmitted", created.ReviewStatus)
	}

	// 1. 采访员提交 → 待审；正式版本仍为空。
	submitted, err := svc.SubmitSummary(interviewer, created.ID, "  初版摘要  ")
	if err != nil {
		t.Fatalf("submit summary: %v", err)
	}
	if submitted.ReviewStatus != constants.ReviewStatusPending || submitted.PendingSummary != "初版摘要" || submitted.Summary != "" {
		t.Fatalf("unexpected state after submit: %+v", submitted)
	}

	// 2. 待审期间重复提交被拒。
	if _, err := svc.SubmitSummary(interviewer, created.ID, "再写一版"); err == nil {
		t.Fatal("expected error when submitting while pending")
	}

	// 3. 档案员退回并写原因；正式版本仍为空，待审版本保留。
	rejected, err := svc.RejectSummary(archivist, created.ID, "缺少时间地点")
	if err != nil {
		t.Fatalf("reject summary: %v", err)
	}
	if rejected.ReviewStatus != constants.ReviewStatusRejected || rejected.RejectReason != "缺少时间地点" {
		t.Fatalf("unexpected state after reject: %+v", rejected)
	}
	if rejected.PendingSummary != "初版摘要" || rejected.Summary != "" {
		t.Fatal("reject should keep pending draft and published summary unchanged")
	}

	// 4. 采访员补充后重新提交 → 待审，退回原因清空。
	resubmitted, err := svc.SubmitSummary(interviewer, created.ID, "补充版摘要")
	if err != nil {
		t.Fatalf("resubmit summary: %v", err)
	}
	if resubmitted.ReviewStatus != constants.ReviewStatusPending || resubmitted.PendingSummary != "补充版摘要" || resubmitted.RejectReason != "" {
		t.Fatalf("unexpected state after resubmit: %+v", resubmitted)
	}

	// 5. 档案员批准 → 已通过，待审版本转正。
	approved, err := svc.ApproveSummary(archivist, created.ID)
	if err != nil {
		t.Fatalf("approve summary: %v", err)
	}
	if approved.ReviewStatus != constants.ReviewStatusApproved || approved.Summary != "补充版摘要" || approved.PendingSummary != "" {
		t.Fatalf("unexpected state after approve: %+v", approved)
	}

	// 6. 已通过后采访员可再提交修改版本；待审期间时间线版本（Summary）保持上一版。
	revised, err := svc.SubmitSummary(interviewer, created.ID, "修订版摘要")
	if err != nil {
		t.Fatalf("submit revision after approval: %v", err)
	}
	if revised.ReviewStatus != constants.ReviewStatusPending || revised.Summary != "补充版摘要" || revised.PendingSummary != "修订版摘要" {
		t.Fatalf("timeline version must stay previous while revision pending: %+v", revised)
	}
	approved2, err := svc.ApproveSummary(archivist, created.ID)
	if err != nil {
		t.Fatalf("approve revision: %v", err)
	}
	if approved2.Summary != "修订版摘要" || approved2.ReviewStatus != constants.ReviewStatusApproved {
		t.Fatalf("revision should become the published version: %+v", approved2)
	}

	if n := len(repo.recordings); n != 1 {
		t.Fatalf("expected 1 recording, got %d", n)
	}
}

func TestSummaryReviewRoleGuard(t *testing.T) {
	svc, _ := newTestRecordingService()
	interviewer := &model.User{ID: 1, Role: constants.RoleInterviewer}
	archivist := &model.User{ID: 2, Role: constants.RoleArchivist}
	admin := &model.User{ID: 3, Role: constants.RoleAdmin}

	r, err := svc.Create(interviewer, &dto.CreateRecordingRequest{ProjectID: 1, QuestionID: 1})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// 档案员不能提交摘要。
	if _, err := svc.SubmitSummary(archivist, r.ID, "x"); err == nil {
		t.Fatal("archivist must not submit summary")
	}
	// 采访员提交后，采访员不能批准/退回。
	if _, err := svc.SubmitSummary(interviewer, r.ID, "待审内容"); err != nil {
		t.Fatalf("submit: %v", err)
	}
	if _, err := svc.ApproveSummary(interviewer, r.ID); err == nil {
		t.Fatal("interviewer must not approve summary")
	}
	if _, err := svc.RejectSummary(interviewer, r.ID, "原因"); err == nil {
		t.Fatal("interviewer must not reject summary")
	}
	// 管理员两种操作都允许。
	if _, err := svc.RejectSummary(admin, r.ID, "管理员退回"); err != nil {
		t.Fatalf("admin reject: %v", err)
	}
	if _, err := svc.SubmitSummary(interviewer, r.ID, "再次提交"); err != nil {
		t.Fatalf("resubmit: %v", err)
	}
	if _, err := svc.ApproveSummary(admin, r.ID); err != nil {
		t.Fatalf("admin approve: %v", err)
	}
}

func TestSummaryReviewInvalidTransitions(t *testing.T) {
	svc, _ := newTestRecordingService()
	interviewer := &model.User{ID: 1, Role: constants.RoleInterviewer}
	archivist := &model.User{ID: 2, Role: constants.RoleArchivist}

	r, _ := svc.Create(interviewer, &dto.CreateRecordingRequest{ProjectID: 1, QuestionID: 1})

	// 未提交状态不能批准/退回。
	if _, err := svc.ApproveSummary(archivist, r.ID); err == nil {
		t.Fatal("approve without pending must fail")
	}
	if _, err := svc.RejectSummary(archivist, r.ID, "原因"); err == nil {
		t.Fatal("reject without pending must fail")
	}
	// 空摘要、空白摘要不能提交。
	if _, err := svc.SubmitSummary(interviewer, r.ID, ""); err == nil {
		t.Fatal("empty summary must fail")
	}
	if _, err := svc.SubmitSummary(interviewer, r.ID, "   "); err == nil {
		t.Fatal("blank summary must fail")
	}
	// 空退回原因不能退回。
	if _, err := svc.SubmitSummary(interviewer, r.ID, "待审"); err != nil {
		t.Fatalf("submit: %v", err)
	}
	if _, err := svc.RejectSummary(archivist, r.ID, "  "); err == nil {
		t.Fatal("blank reject reason must fail")
	}
	// 已通过后不能再次批准（没有待审版本）。
	if _, err := svc.ApproveSummary(archivist, r.ID); err != nil {
		t.Fatalf("approve pending: %v", err)
	}
	if _, err := svc.ApproveSummary(archivist, r.ID); err == nil {
		t.Fatal("approving twice must fail")
	}
}

func TestRecordingListByReviewStatus(t *testing.T) {
	svc, _ := newTestRecordingService()
	interviewer := &model.User{ID: 1, Role: constants.RoleInterviewer}

	r1, _ := svc.Create(interviewer, &dto.CreateRecordingRequest{ProjectID: 1, QuestionID: 1})
	r2, _ := svc.Create(interviewer, &dto.CreateRecordingRequest{ProjectID: 1, QuestionID: 1})
	if _, err := svc.SubmitSummary(interviewer, r1.ID, "待审摘要"); err != nil {
		t.Fatalf("submit: %v", err)
	}

	pending, err := svc.List(0, 0, constants.ReviewStatusPending)
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(pending) != 1 || pending[0].ID != r1.ID {
		t.Fatalf("expected only recording %d pending, got %+v", r1.ID, pending)
	}

	unsubmitted, err := svc.List(1, 0, constants.ReviewStatusUnsubmitted)
	if err != nil {
		t.Fatalf("list unsubmitted: %v", err)
	}
	if len(unsubmitted) != 1 || unsubmitted[0].ID != r2.ID {
		t.Fatalf("expected only recording %d unsubmitted in project 1", r2.ID)
	}

	if _, err := svc.List(0, 0, "bogus"); err == nil {
		t.Fatal("invalid review status filter must fail")
	}
}

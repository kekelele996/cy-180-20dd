package service

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/dto"
	"github.com/oralhistory/oralhistory/internal/model"
	"github.com/oralhistory/oralhistory/internal/repository"
	"github.com/oralhistory/oralhistory/internal/util"
)

// RecordingService 录音片段业务接口。
type RecordingService interface {
	Create(actor *model.User, req *dto.CreateRecordingRequest) (*model.Recording, error)
	Get(id uint) (*model.Recording, error)
	// List 同时服务「按项目」与「按问题」两个接口，复用同一 service 方法。
	List(projectID, questionID uint) ([]model.Recording, error)
	Update(actor *model.User, id uint, req *dto.UpdateRecordingRequest) (*model.Recording, error)
	// SubmitSummary 采访员提交/重新提交摘要，进入待审；待审期间不改变时间线展示的已通过版本。
	SubmitSummary(actor *model.User, id uint, summary string) (*model.Recording, error)
	// ReviewSummary 档案员批准（切换时间线版本）或填写原因退回。
	ReviewSummary(actor *model.User, id uint, action, reason string) (*model.Recording, error)
	AttachAudio(actor *model.User, id uint, audioKey string, duration int) (*model.Recording, error)
	Delete(actor *model.User, id uint) error
	CountByProject(projectID uint) (int64, error)
}

type recordingService struct {
	recordingRepo repository.RecordingRepository
	projectRepo   repository.ProjectRepository
	questionRepo  repository.QuestionRepository
	logger        *slog.Logger
}

// NewRecordingService 构造录音服务。
func NewRecordingService(recordingRepo repository.RecordingRepository, projectRepo repository.ProjectRepository, questionRepo repository.QuestionRepository, logger *slog.Logger) RecordingService {
	return &recordingService{recordingRepo: recordingRepo, projectRepo: projectRepo, questionRepo: questionRepo, logger: logger}
}

func (s *recordingService) Create(actor *model.User, req *dto.CreateRecordingRequest) (*model.Recording, error) {
	if _, err := s.projectRepo.FindByID(req.ProjectID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, fmt.Sprintf("项目 %d 不存在", req.ProjectID), err)
		}
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询项目 %d 失败", req.ProjectID), err)
	}
	if _, err := s.questionRepo.FindByID(req.QuestionID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, fmt.Sprintf("问题 %d 不存在", req.QuestionID), err)
		}
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询问题 %d 失败", req.QuestionID), err)
	}
	recording := &model.Recording{
		ProjectID:       req.ProjectID,
		QuestionID:      req.QuestionID,
		DurationSeconds: req.DurationSeconds,
		Status:          constants.RecordingStatusRecording,
		ReviewStatus:    constants.ReviewStatusDraft,
		CreatedBy:       actor.ID,
	}
	if err := s.recordingRepo.Create(recording); err != nil {
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("创建问题 %d 的录音失败", req.QuestionID), err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogRecordingUpload, actor.Username, recording.ProjectID, recording.QuestionID, recording.DurationSeconds, recording.Status))
	return recording, nil
}

func (s *recordingService) Get(id uint) (*model.Recording, error) {
	recording, err := s.recordingRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, fmt.Sprintf("录音 %d 不存在", id), err)
		}
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询录音 %d 失败", id), err)
	}
	return recording, nil
}

func (s *recordingService) List(projectID, questionID uint) ([]model.Recording, error) {
	var (
		recordings []model.Recording
		err        error
	)
	if projectID > 0 {
		recordings, err = s.recordingRepo.ListByProject(projectID)
	} else {
		recordings, err = s.recordingRepo.ListByQuestion(questionID)
	}
	if err != nil {
		return nil, util.NewAppError(constants.CodeInternal, "录音列表查询失败", err)
	}
	return recordings, nil
}

func (s *recordingService) Update(actor *model.User, id uint, req *dto.UpdateRecordingRequest) (*model.Recording, error) {
	recording, err := s.recordingRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, fmt.Sprintf("录音 %d 不存在", id), err)
		}
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询录音 %d 失败", id), err)
	}
	if req.DurationSeconds > 0 {
		recording.DurationSeconds = req.DurationSeconds
	}
	if req.Status != "" {
		if !constants.ValidRecordingStatus(req.Status) {
			return nil, util.NewAppError(constants.CodeValidation, fmt.Sprintf("录音状态 %s 不合法", req.Status), nil)
		}
		if !constants.CanTransitionRecording(recording.Status, req.Status) {
			return nil, util.NewAppError(constants.CodeRecordingStatus,
				fmt.Sprintf("录音 %d 状态不允许从 %s 流转到 %s", id, recording.Status, req.Status), nil)
		}
		recording.Status = req.Status
	}
	if err := s.recordingRepo.Update(recording); err != nil {
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("更新录音 %d 失败", id), err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogRecordingStatus, actor.Username, recording.ID, recording.Status, recording.Status))
	return recording, nil
}

func (s *recordingService) SubmitSummary(actor *model.User, id uint, summary string) (*model.Recording, error) {
	if actor.Role != constants.RoleInterviewer && actor.Role != constants.RoleAdmin {
		return nil, util.NewAppError(constants.CodeForbidden, "仅采访员可以提交摘要审核", nil)
	}
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return nil, util.NewAppError(constants.CodeValidation, "摘要内容不能为空", nil)
	}
	recording, err := s.recordingRepo.FindByIDForUpdate(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, fmt.Sprintf("录音 %d 不存在", id), err)
		}
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询录音 %d 失败", id), err)
	}
	if !constants.CanSubmitReview(recording.ReviewStatus) {
		return nil, util.NewAppError(constants.CodeReviewStatus,
			fmt.Sprintf("录音 %d 摘要当前为待审状态，档案员处理后才能重新提交", id), nil)
	}
	// 新内容先作为待审版本保存，已通过版本 Summary 保持不变（时间线继续显示上一版）。
	recording.PendingSummary = summary
	recording.ReviewStatus = constants.ReviewStatusPending
	recording.RejectReason = ""
	recording.ReviewedBy = 0
	recording.ReviewedAt = nil
	if err := s.recordingRepo.Update(recording); err != nil {
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("提交录音 %d 摘要审核失败", id), err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogSummarySubmit, actor.Username, recording.ID, recording.ReviewStatus))
	return recording, nil
}

func (s *recordingService) ReviewSummary(actor *model.User, id uint, action, reason string) (*model.Recording, error) {
	if actor.Role != constants.RoleArchivist && actor.Role != constants.RoleAdmin {
		return nil, util.NewAppError(constants.CodeForbidden, "仅档案员可以审核摘要", nil)
	}
	reason = strings.TrimSpace(reason)
	if action == "reject" && reason == "" {
		return nil, util.NewAppError(constants.CodeValidation, "退回摘要时必须填写退回原因", nil)
	}
	recording, err := s.recordingRepo.FindByIDForUpdate(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, fmt.Sprintf("录音 %d 不存在", id), err)
		}
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询录音 %d 失败", id), err)
	}
	if !constants.CanReview(recording.ReviewStatus) {
		return nil, util.NewAppError(constants.CodeReviewStatus,
			fmt.Sprintf("录音 %d 摘要当前不是待审状态，无法审核", id), nil)
	}
	now := time.Now()
	recording.ReviewedBy = actor.ID
	recording.ReviewedAt = &now
	switch action {
	case "approve":
		// 批准后待审版本切换为正式版本，出现在项目时间线。
		recording.Summary = recording.PendingSummary
		recording.PendingSummary = ""
		recording.ReviewStatus = constants.ReviewStatusApproved
		recording.RejectReason = ""
	case "reject":
		// 退回后保留待审版本供采访员补充，已通过版本（上一版）继续在时间线展示。
		recording.ReviewStatus = constants.ReviewStatusRejected
		recording.RejectReason = reason
	}
	if err := s.recordingRepo.Update(recording); err != nil {
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("审核录音 %d 摘要失败", id), err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogSummaryReview, actor.Username, recording.ID, action, recording.ReviewStatus))
	return recording, nil
}

func (s *recordingService) AttachAudio(actor *model.User, id uint, audioKey string, duration int) (*model.Recording, error) {
	recording, err := s.recordingRepo.FindByIDForUpdate(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, fmt.Sprintf("录音 %d 不存在", id), err)
		}
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询录音 %d 失败", id), err)
	}
	recording.AudioKey = audioKey
	if duration > 0 {
		recording.DurationSeconds = duration
	}
	if recording.Status == constants.RecordingStatusRecording || recording.Status == constants.RecordingStatusProcessing {
		recording.Status = constants.RecordingStatusReady
	}
	if err := s.recordingRepo.Update(recording); err != nil {
		return nil, util.NewAppError(constants.CodeInternal, fmt.Sprintf("录音 %d 音频关联失败", id), err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogRecordingUpload, actor.Username, recording.ProjectID, recording.QuestionID, recording.DurationSeconds, recording.Status))
	return recording, nil
}

func (s *recordingService) Delete(actor *model.User, id uint) error {
	if _, err := s.recordingRepo.FindByID(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.CodeNotFound, fmt.Sprintf("录音 %d 不存在", id), err)
		}
		return util.NewAppError(constants.CodeInternal, fmt.Sprintf("查询录音 %d 失败", id), err)
	}
	if err := s.recordingRepo.Delete(id); err != nil {
		return util.NewAppError(constants.CodeInternal, fmt.Sprintf("删除录音 %d 失败", id), err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogRecordingDelete, actor.Username, id))
	return nil
}

func (s *recordingService) CountByProject(projectID uint) (int64, error) {
	return s.recordingRepo.CountByProject(projectID)
}

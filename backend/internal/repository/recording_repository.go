package repository

import (
	"errors"
	"fmt"

	"github.com/oralhistory/oralhistory/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RecordingRepository 录音片段数据访问接口。
type RecordingRepository interface {
	Create(recording *model.Recording) error
	FindByID(id uint) (*model.Recording, error)
	// List 按项目/问题（任一为 0 则不过滤该维度）并可按摘要审核状态筛选；
	// 全部过滤条件为 0 值时返回全量录音（供档案员审核工作台使用）。
	List(projectID, questionID uint, reviewStatus string) ([]model.Recording, error)
	ListByProject(projectID uint) ([]model.Recording, error)
	ListByQuestion(questionID uint) ([]model.Recording, error)
	FindByIDForUpdate(id uint) (*model.Recording, error)
	Update(recording *model.Recording) error
	UpdateStatus(recording *model.Recording) error
	Delete(id uint) error
	CountByProject(projectID uint) (int64, error)
}

type recordingRepository struct {
	db *gorm.DB
}

// NewRecordingRepository 构造录音仓储。
func NewRecordingRepository(db *gorm.DB) RecordingRepository {
	return &recordingRepository{db: db}
}

func (r *recordingRepository) Create(recording *model.Recording) error {
	if err := r.db.Create(recording).Error; err != nil {
		return fmt.Errorf("create recording of question %d: %w", recording.QuestionID, err)
	}
	return nil
}

func (r *recordingRepository) FindByID(id uint) (*model.Recording, error) {
	var recording model.Recording
	if err := r.db.First(&recording, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find recording by id %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("find recording by id: %w", err)
	}
	return &recording, nil
}

func (r *recordingRepository) List(projectID, questionID uint, reviewStatus string) ([]model.Recording, error) {
	var recordings []model.Recording
	query := r.db.Model(&model.Recording{})
	if projectID > 0 {
		query = query.Where("project_id = ?", projectID)
	}
	if questionID > 0 {
		query = query.Where("question_id = ?", questionID)
	}
	if reviewStatus != "" {
		query = query.Where("review_status = ?", reviewStatus)
	}
	if err := query.Order("id ASC").Find(&recordings).Error; err != nil {
		return nil, fmt.Errorf("list recordings project=%d question=%d review=%s: %w", projectID, questionID, reviewStatus, err)
	}
	return recordings, nil
}

func (r *recordingRepository) ListByProject(projectID uint) ([]model.Recording, error) {
	return r.List(projectID, 0, "")
}

func (r *recordingRepository) ListByQuestion(questionID uint) ([]model.Recording, error) {
	return r.List(0, questionID, "")
}

func (r *recordingRepository) FindByIDForUpdate(id uint) (*model.Recording, error) {
	var recording model.Recording
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&recording, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find recording for update by id %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("find recording for update: %w", err)
	}
	return &recording, nil
}

func (r *recordingRepository) Update(recording *model.Recording) error {
	if err := r.db.Save(recording).Error; err != nil {
		return fmt.Errorf("update recording %d: %w", recording.ID, err)
	}
	return nil
}

func (r *recordingRepository) UpdateStatus(recording *model.Recording) error {
	if err := r.db.Model(recording).Update("status", recording.Status).Error; err != nil {
		return fmt.Errorf("update recording %d status: %w", recording.ID, err)
	}
	return nil
}

func (r *recordingRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("recording_id = ?", id).Delete(&model.TimelineMarker{}).Error; err != nil {
			return fmt.Errorf("delete markers of recording %d: %w", id, err)
		}
		if err := tx.Delete(&model.Recording{}, id).Error; err != nil {
			return fmt.Errorf("delete recording %d: %w", id, err)
		}
		return nil
	})
}

func (r *recordingRepository) CountByProject(projectID uint) (int64, error) {
	var total int64
	if err := r.db.Model(&model.Recording{}).Where("project_id = ?", projectID).Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count recordings of project %d: %w", projectID, err)
	}
	return total, nil
}

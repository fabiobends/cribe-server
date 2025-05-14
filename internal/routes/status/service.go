package status

import "time"

type StatusServiceInterface interface {
	GetStatus() StatusInfo
}

type StatusService struct {
	repo           StatusRepository
	getCurrentTime func() time.Time
}

func NewStatusService(repo StatusRepository, getCurrentTime func() time.Time) *StatusService {
	return &StatusService{repo: repo, getCurrentTime: getCurrentTime}
}

func (service *StatusService) GetStatus() StatusInfo {
	now := service.getCurrentTime()
	updatedAt := now.Format(time.RFC3339)

	databaseInfo, _ := service.repo.GetDatabaseInfo()
	dependencies := Dependencies{
		Database: databaseInfo,
	}

	return StatusInfo{
		UpdatedAt:    updatedAt,
		Dependencies: dependencies,
	}
}

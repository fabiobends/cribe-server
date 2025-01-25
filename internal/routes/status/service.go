package status

import "time"

type StatusServiceInterface interface {
	GetStatus() GetStatusResponse
}

type StatusService struct {
	repo           StatusRepository
	getCurrentTime func() time.Time
}

func NewStatusService(repo StatusRepository, getCurrentTime func() time.Time) *StatusService {
	return &StatusService{repo: repo, getCurrentTime: getCurrentTime}
}

func (service *StatusService) GetStatus() GetStatusResponse {
	now := service.getCurrentTime()
	updatedAt := now.Format(time.RFC3339)

	databaseInfo, _ := service.repo.GetDatabaseInfo()
	dependencies := Dependencies{
		Database: databaseInfo,
	}

	return GetStatusResponse{
		UpdatedAt:    updatedAt,
		Dependencies: dependencies,
	}
}

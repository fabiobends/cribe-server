package status

type StatusServiceInterface interface {
	GetStatus() string
}

type StatusService struct {
	repo Repository
}

func NewStatusService(repo Repository) *StatusService {
	return &StatusService{repo: repo}
}

func (service *StatusService) GetStatus() string {
	return service.repo.GetStatusMessage()
}

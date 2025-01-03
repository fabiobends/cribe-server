package status

type Repository interface {
	GetStatusMessage() string
}

type StatusRepository struct{}

func (sr *StatusRepository) GetStatusMessage() string {
	return "ok"
}

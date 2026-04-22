package service

// healthService implements HealthService
type healthService struct{}

func NewHealthService() HealthService {
	return &healthService{}
}

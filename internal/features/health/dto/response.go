package dto

import "time"

// HealthData is the internal data transfer object from service to controller.
type HealthData struct {
	Status    string
	Timestamp string
	Uptime    string
	Version   string
	Checks    map[string]string
	CheckedAt time.Time
}

package dto

type UpdatePreferencesRequest struct {
	Notifications *bool   `json:"notifications,omitempty"`
	Language      *string `json:"language,omitempty"`
}

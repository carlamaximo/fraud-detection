package domain

type Event struct {
	UserID    string `json:"user_id" validate:"required"`
	EventType string `json:"event_type" validate:"required"`
	IP        string `json:"ip" validate:"required"`
	Country   string `json:"country" validate:"required"`
	Device    string `json:"device" validate:"required"`
}
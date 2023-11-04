package structs

type Flow struct {
	ID           int          `json:"id,omitempty"`
	ClientID     int          `json:"client_id,omitempty"`
	OutboundFlow OutboundFlow `json:"outbound_json" validate:"required"`
	Status       string       `json:"status,omitempty"`
	CreatedAt    string       `json:"created_at,omitempty"`
	StatusReport StatusReport `json:"status_report,omitempty"`
}

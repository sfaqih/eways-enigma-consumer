package structs

type Status struct {
	ID                 int64  `json:"enigma_id,omitempty"`
	ClientID           int    `json:"client_id,omitempty"`
	ContactID          string `json:"contact_id,omitempty"`
	Channel            string `json:"channel,omitempty"`
	ChannelID          string `json:"channel_id,omitempty"`
	Platform           string `json:"platform,omitempty"`
	To                 string `json:"to,omitempty"`
	MessageID          string `json:"id,omitempty"`
	Direction          string `json:"direction,omitempty"`
	From               string `json:"from,omitempty"`
	ContactStatus      string `json:"contact_status,omitempty"`
	MSISDN             string `json:"msisdn,omitempty"`
	ConversationID     string `json:"conversation_id,omitempty"`
	ConversationStatus string `json:"conversation_status,omitempty"`
	SesameMessageID    string `json:"sesame_message_id,omitempty"`
	SesameID           string `json:"sesame_id,omitempty"`
	Status             int    `json:"status,omitempty"`
	MsgStatus          string `json:"message_status,omitempty"`
	ReqBody            string `json:"requst_body,omitempty"`
	MsgJson            string `json:"message_json,omitempty"`
	ContactJson        string `json:"contact_json,omitempty"`
	ConvJson           string `json:"conversation_json,omitempty"`
	NewStatus          string `json:"new_status,omitempty"`
	NewStatusInt       int    `json:"new_status_id,omitempty"`
	ApiID              int    `json:"api_id,omitempty"`
	Type               string `json:"type,omitempty"`
}

type MBRespStatus struct {
	ID        string `json:"id,omitempty"`
	Direction string `json:"direction,omitempty"`
	Status    string `json:"status,omitempty"`
}

type StatusDeliverySesame struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"`
}

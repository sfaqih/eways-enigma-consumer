package structs

type Incoming struct {
	Contact      Contact      `json:"contact"`
	Conversation Conversation `json:"conversation"`
	Message      IncMessage   `json:"message"`
	Type         string       `json:"type"`
}

type ConversationDevilery struct {
	Conversation Incoming  `json:"conversation"`
	Webhooks     []*WebHook `json:"webhooks"`
}

type IncMessage struct {
	ID        string  `json:"id,omitempty"`
	ChannelID string  `json:"channelId,omitempty"`
	Content   Content `json:"content,omitempty"`
	Direction string  `json:"direction,omitempty"`
	Status    string  `json:"status,omitempty"`
	From      string  `json:"from,omitempty"`
	To        string  `json:"to,omitempty"`
	Platform  string  `json:"platform,omitempty"`
}

type Conversation struct {
	ID     string `json:"id"`
	Status string `json:"status,omitempty"`
}

type Contact struct {
	ID            string `json:"id"`
	MSISDN        string `json:"msisdn,omitempty"`
	ContactStatus string `json:"status,omitempty"`
}

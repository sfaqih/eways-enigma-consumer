package structs

type InboundWA struct {
	Messages []Message `json:"messages" validate:"required"`
	ClientID int       `json:"client_id" validate:"required"`
}

type Message struct {
	From      string   `json:"from" validate:"required"`
	ID        string   `json:"id" validate:"required"`
	Timestamp int64    `json:"timestamp" validate:"required"`
	Text      Text     `json:"text,omitempty"`
	Location  Location `json:"location,omitempty"`
	Image     Image    `json:"image,omitempty"`
	Document  Document `json:"document,omitempty"`
	Type      string   `json:"type"`
}

type Text struct {
	Body string `json:"body"`
}

type Location struct {
	Latitude  string `json:"latitude,omitempty"`
	Longitude string `json:"longitude,omitempty"`
}

type Image struct {
	ID       string `json:"id"`
	MimeType string `json:"mime_type"`
	Caption  string `json:"caption,omitempty"`
}

type Document struct {
	ID       string `json:"id"`
	MimeType string `json:"mime_type"`
	Caption  string `json:"caption,omitempty"`
}

type Damcorp_Inbound_Request_WA struct {
	ClientID int                `json:"client_id,omitempty"`
	ReqData  Damcorp_Inbound_WA `json:"data,omitempty"`
	Webhook  []WebHook          `json:"webhook,omitempty"`
}

type Damcorp_Inbound_WA struct {
	Contacts []Damcorp_Inbound_Contact    `json:"contacts,omitempty"`
	Messages []Damcorp_Inbound_WA_Message `json:"messages,omitempty"`
}

type Damcorp_Inbound_Contact struct {
	Profile Damcorp_Contact_Profile `json:"profile,omitempty"`
	WaId    string                  `json:"wa_id,omitempty"`
}

type Damcorp_Contact_Profile struct {
	Name string `json:"name,omitempty"`
}

type Damcorp_Inbound_WA_Message struct {
	From        string                         `json:"from,omitempty"`
	Id          string                         `json:"id,omitempty"`
	Timestamp   string                         `json:"timestamp,omitempty"`
	Type        string                         `json:"type,omitempty"`
	Text        Damcorp_Inbound_WA_Text        `json:"text,omitempty"`
	Interactive Damcorp_Inbound_WA_Interactive `json:"interactive,omitempty"`
	Context     Damcorp_Inbound_WA_Context     `json:"context,omitempty"`
}

type Damcorp_Inbound_WA_Text struct {
	Body string `json:"body,omitempty"`
}

type Damcorp_Inbound_WA_Context struct {
	From string `json:"from,omitempty"`
	Id   string `json:"id,omitempty"`
}

type Damcorp_Inbound_WA_Interactive struct {
	Type        string                          `json:"type,omitempty"`
	ListReply   Damcorp_WA_Interactive_Reply    `json:"list_reply,omitempty"`
	ButtonReply Damcorp_Inbound_WA_Button_Reply `json:"button_reply,omitempty"`
}

type Damcorp_Inbound_WA_Button_Reply struct {
	Id    string `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
}

type Damcorp_WA_Interactive_Reply struct {
	Description string `json:"description,omitempty"`
	Id          string `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
}

type Damcorp_Request_Media struct {
	ClientId int    `json:"client_id" validate:"required"`
	Channel  string `json:"channel" validate:"required,oneof=DamcorpMedia"`
	MediaId  string `json:"media_id" validate:"required"`
}

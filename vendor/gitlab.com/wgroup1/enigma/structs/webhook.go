package structs

type WebHookParamObj struct {
	ID         int    `json:"id"`
	ClientId   int    `json:"client_id,omitempty"`
	ClientCode string `json:"client_code,omitempty"`
	Page       string `json:"page" validate:"required"`
	Limit      string `json:"limit" validate:"required"`
	//, page string, limit string
}

type WebHook struct {
	ID           int     `json:"id,omitempty"`
	ClientCode   string  `json:"client_code" validate:"required"`
	ClientID     int     `json:"client_id,omitempty"`
	URL          string  `json:"url" validate:"required"`
	Method       string  `json:"method" validate:"required"`
	HeaderPrefix string  `json:"header_prefix,omitempty"`
	Token        string  `json:"token,omitempty"`
	Events       []Event `json:"events" validate:"required"`
	TrxType      string  `json:"trx_type" validate:"required"`
	HttpCode     int     `json:"expected_http_code" validate:"required"`
	Retry        int     `json:"retry" validate:"required,gt=0,lte=5"`
	Timeout      int     `json:"timeout" validate:"required,gt=0,lte=30"`
	Status       int     `json:"status,omitempty"`
	CreatedAt    string  `json:"created_at,omitempty"`
	CreatedBy    string  `json:"created_by,omitempty"`
}

type WebHookObj struct {
	ID           int     `json:"id,omitempty"`
	ClientID     int     `json:"client_id,omitempty"`
	ClientCode   string  `json:"client_code" validate:"required"`
	URL          string  `json:"url" validate:"required"`
	Method       string  `json:"method" validate:"required,oneof=GET POST PUT PATCH DELETE"`
	HeaderPrefix string  `json:"header_prefix,omitempty"`
	Token        string  `json:"token,omitempty"`
	Events       []Event `json:"events,omitempty"`
	TrxType      string  `json:"trx_type" validate:"required,oneof=Inbound Conversation SparkPostStatus InboundDto"`
	HttpCode     int     `json:"expected_http_code" validate:"required"`
	Retry        int     `json:"retry" validate:"required,gt=0,lte=5"`
	Timeout      int     `json:"timeout" validate:"required,gt=0,lte=30"`
	Status       int     `json:"status,omitempty"`
	CreatedAt    string  `json:"created_at,omitempty"`
	CreatedBy    string  `json:"created_by,omitempty"`
}

type WebHookPlain struct {
	ID           int    `json:"id"`
	ClientID     int    `json:"client_id"`
	ClientCode   string `json:"client_code"`
	URL          string `json:"url"`
	Method       string `json:"method"`
	HeaderPrefix string `json:"header_prefix,omitempty"`
	Token        string `json:"token,omitempty"`
	Events       string `json:"events,omitempty"`
	HttpCode     int    `json:"expected_http_code,omitempty"`
	Retry        int    `json:"retry,omitempty"`
	Timeout      int    `json:"timeout,omitempty"`
	Status       int    `json:"status,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	CreatedBy    string `json:"created_by,omitempty"`
}

type Event struct {
	//EventName string `json:"event_name"`
	ChannelName string `json:"channel_name,omitempty"`
	From        string `json:"from,omitempty"`
	To          string `json:"to,omitempty"`
}

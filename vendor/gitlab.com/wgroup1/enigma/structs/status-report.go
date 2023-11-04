package structs

import "time"

type StatusReport struct {
	ID          int    `json:"id,omitempty"`
	Flowid      int    `json:"flow_id,omitempty"`
	ClientID    int    `json:"client_id,omitempty"`
	Channel     string `json:"channel,omitempty"`
	Destination string `json:"destination,omitempty"`
	//MessageId       string `json:"message_id,omitempty"`
	MessageTemplate       string                `json:"message_template,omitempty"`
	Status                string                `json:"status,omitempty"`
	StatusJson            string                `json:"status_json,omitempty"`
	TransactionAt         string                `json:"transaction_at,omitempty"`
	IsProcess             int                   `json:"is_process,omitempty"`
	StatusReportMailJet   StatusReportMailJet   `json:"status_report_mailjet,omitempty"`
	StatusReportSparkPost StatusReportSparkPost `json:"status_report_sparkpost,omitempty"`

	//struct MailJet
	Event     string `json:"event,omitempty"`
	Timestamp int64  `json:"time,omitempty"`
	//MessageID      int64  `json:"MessageID,omitempty"`
	Message_GUID   string `json:"Message_GUID,omitempty"`
	Email          string `json:"email,omitempty"`
	MjCampaignId   int64  `json:"mj_campaign_id,omitempty"`
	MjContactId    int64  `json:"mj_contact_id,omitempty"`
	CustomCampaign string `json:"customcampaign,omitempty"`
	SmtpReply      string `json:"smtp_reply,omitempty"`
	CustomID       string `json:"CustomID,omitempty"`
	Payload        string `json:"Payload,omitempty"`
	Ip             string `json:"ip,omitempty"`
	Geo            string `json:"geo,omitempty"`
	Agent          string `json:"agent,omitempty"`

	//struct SparkPost
	ClickTracking         bool      `json:"click_tracking,omitempty"`
	CustomerID            string    `json:"customer_id,omitempty"`
	DelvMethod            string    `json:"delv_method,omitempty"`
	EventID               string    `json:"event_id,omitempty"`
	FriendlyFrom          string    `json:"friendly_from,omitempty"`
	InitialPixel          bool      `json:"initial_pixel,omitempty"`
	InjectionTime         time.Time `json:"injection_time,omitempty"`
	IPAddress             string    `json:"ip_address,omitempty"`
	IPPool                string    `json:"ip_pool,omitempty"`
	MailboxProvider       string    `json:"mailbox_provider,omitempty"`
	MailboxProviderRegion string    `json:"mailbox_provider_region,omitempty"`
	MessageID             string    `json:"message_id,omitempty"`
	MsgFrom               string    `json:"msg_from,omitempty"`
	MsgSize               string    `json:"msg_size,omitempty"`
	NumRetries            string    `json:"num_retries,omitempty"`
	OpenTracking          bool      `json:"open_tracking,omitempty"`
	OutboundTLS           string    `json:"outbound_tls,omitempty"`
	QueueTime             string    `json:"queue_time,omitempty"`
	////RcptMeta              struct {	} `json:"rcpt_meta,omitempty"`
	////RcptTags        []interface{} `json:"rcpt_tags,omitempty"`
	RcptTo          string `json:"rcpt_to,omitempty"`
	RecvMethod      string `json:"recv_method,omitempty"`
	RoutingDomain   string `json:"routing_domain,omitempty"`
	SendingIP       string `json:"sending_ip,omitempty"`
	Subject         string `json:"subject,omitempty"`
	TemplateID      string `json:"template_id,omitempty"`
	TemplateVersion string `json:"template_version,omitempty"`
	//Timestamp       string `json:"timestamp,omitempty"`
	TransmissionID  string `json:"transmission_id,omitempty"`
	Type            string `json:"type,omitempty"`
	RawRcptTo       string `json:"raw_rcpt_to,omitempty"`
	RecipientDomain string `json:"recipient_domain,omitempty"`

	MaxOrders int `json:"max_orders,omitempty"`
}

type StatusReportMailJet struct {
	Event          string `json:"event,omitempty"`
	Timestamp      int64  `json:"time,omitempty"`
	MessageID      int64  `json:"MessageID,omitempty"`
	Message_GUID   string `json:"Message_GUID,omitempty"`
	Email          string `json:"email,omitempty"`
	MjCampaignId   int64  `json:"mj_campaign_id,omitempty"`
	MjContactId    int64  `json:"mj_contact_id,omitempty"`
	CustomCampaign string `json:"customcampaign,omitempty"`
	SmtpReply      string `json:"smtp_reply,omitempty"`
	CustomID       string `json:"CustomID,omitempty"`
	Payload        string `json:"Payload,omitempty"`
	Ip             string `json:"ip,omitempty"`
	Geo            string `json:"geo,omitempty"`
	Agent          string `json:"agent,omitempty"`
}

// SPARK POST
/*
type StatusReportSparkPost struct {
	Msys Msys `json:"msys,omitempty"`
}

type Msys struct {
	MessageEvent MessageEvent `json:"message_event,omitempty"`
	TrackEvent   TrackEvent   `json:"track_event,omitempty"`
}

type MessageEvent struct {
	ClickTracking         bool      `json:"click_tracking,omitempty"`
	CustomerID            string    `json:"customer_id,omitempty"`
	DelvMethod            string    `json:"delv_method,omitempty"`
	EventID               string    `json:"event_id,omitempty"`
	FriendlyFrom          string    `json:"friendly_from,omitempty"`
	InitialPixel          bool      `json:"initial_pixel,omitempty"`
	InjectionTime         time.Time `json:"injection_time,omitempty"`
	IPAddress             string    `json:"ip_address,omitempty"`
	IPPool                string    `json:"ip_pool,omitempty"`
	MailboxProvider       string    `json:"mailbox_provider,omitempty"`
	MailboxProviderRegion string    `json:"mailbox_provider_region,omitempty"`
	MessageID             string    `json:"message_id,omitempty"`
	MsgFrom               string    `json:"msg_from,omitempty"`
	MsgSize               string    `json:"msg_size,omitempty"`
	NumRetries            string    `json:"num_retries,omitempty"`
	OpenTracking          bool      `json:"open_tracking,omitempty"`
	OutboundTLS           string    `json:"outbound_tls,omitempty"`
	QueueTime             string    `json:"queue_time,omitempty"`
	//RcptMeta              struct {	} `json:"rcpt_meta,omitempty"`
	//RcptTags        []interface{} `json:"rcpt_tags,omitempty"`
	RcptTo          string `json:"rcpt_to,omitempty"`
	RecvMethod      string `json:"recv_method,omitempty"`
	RoutingDomain   string `json:"routing_domain,omitempty"`
	SendingIP       string `json:"sending_ip,omitempty"`
	Subject         string `json:"subject,omitempty"`
	TemplateID      string `json:"template_id,omitempty"`
	TemplateVersion string `json:"template_version,omitempty"`
	Timestamp       string `json:"timestamp,omitempty"`
	TransmissionID  string `json:"transmission_id,omitempty"`
	Type            string `json:"type,omitempty"`
	RawRcptTo       string `json:"raw_rcpt_to,omitempty"`
	RecipientDomain string `json:"recipient_domain,omitempty"`
}

type TrackEvent struct {
	ClickTracking         bool      `json:"click_tracking,omitempty"`
	CustomerID            string    `json:"customer_id,omitempty"`
	DelvMethod            string    `json:"delv_method,omitempty"`
	EventID               string    `json:"event_id,omitempty"`
	FriendlyFrom          string    `json:"friendly_from,omitempty"`
	GeoIp                 GeoIp     `json:"geo_ip,omitempty"`
	InitialPixel          bool      `json:"initial_pixel,omitempty"`
	InjectionTime         time.Time `json:"injection_time,omitempty"`
	IPAddress             string    `json:"ip_address,omitempty"`
	IPPool                string    `json:"ip_pool,omitempty"`
	MailboxProvider       string    `json:"mailbox_provider,omitempty"`
	MailboxProviderRegion string    `json:"mailbox_provider_region,omitempty"`
	MessageID             string    `json:"message_id,omitempty"`
	MsgFrom               string    `json:"msg_from,omitempty"`
	MsgSize               string    `json:"msg_size,omitempty"`
	NumRetries            string    `json:"num_retries,omitempty"`
	OpenTracking          bool      `json:"open_tracking,omitempty"`
	OutboundTLS           string    `json:"outbound_tls,omitempty"`
	QueueTime             string    `json:"queue_time,omitempty"`
	//RcptMeta              struct {	} `json:"rcpt_meta,omitempty"`
	//RcptTags        []interface{} `json:"rcpt_tags,omitempty"`
	RcptTo          string `json:"rcpt_to,omitempty"`
	RecvMethod      string `json:"recv_method,omitempty"`
	RoutingDomain   string `json:"routing_domain,omitempty"`
	SendingIP       string `json:"sending_ip,omitempty"`
	Subject         string `json:"subject,omitempty"`
	TemplateID      string `json:"template_id,omitempty"`
	TemplateVersion string `json:"template_version,omitempty"`
	Timestamp       string `json:"timestamp,omitempty"`
	TransmissionID  string `json:"transmission_id,omitempty"`
	Type            string `json:"type,omitempty"`
	RawRcptTo       string `json:"raw_rcpt_to,omitempty"`
	RecipientDomain string `json:"recipient_domain,omitempty"`
}

type GeoIp struct {
	Latitude   string `json:"latitude,omitempty"`
	Longitude  string `json:"longitude,omitempty"`
	City       string `json:"city,omitempty"`
	Region     string `json:"region,omitempty"`
	Country    string `json:"country,omitempty"`
	Zip        string `json:"zip,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Continent  string `json:"continent,omitempty"`
}


*/

type StatusReportSparkPost struct {
	Msys Msys `json:"msys,omitempty"`
}

type Msys struct {
	MessageEvent MessageEvent `json:"message_event,omitempty"`
	TrackEvent   TrackEvent   `json:"track_event,omitempty"`
}

type MessageEvent struct {
	ClickTracking         bool      `json:"click_tracking,omitempty"`
	CustomerID            string    `json:"customer_id,omitempty"`
	DelvMethod            string    `json:"delv_method,omitempty"`
	EventID               string    `json:"event_id,omitempty"`
	FriendlyFrom          string    `json:"friendly_from,omitempty"`
	InitialPixel          bool      `json:"initial_pixel,omitempty"`
	InjectionTime         time.Time `json:"injection_time,omitempty"`
	IPAddress             string    `json:"ip_address,omitempty"`
	IPPool                string    `json:"ip_pool,omitempty"`
	MailboxProvider       string    `json:"mailbox_provider,omitempty"`
	MailboxProviderRegion string    `json:"mailbox_provider_region,omitempty"`
	MessageID             string    `json:"message_id,omitempty"`
	MsgFrom               string    `json:"msg_from,omitempty"`
	MsgSize               string    `json:"msg_size,omitempty"`
	NumRetries            string    `json:"num_retries,omitempty"`
	OpenTracking          bool      `json:"open_tracking,omitempty"`
	OutboundTLS           string    `json:"outbound_tls,omitempty"`
	QueueTime             string    `json:"queue_time,omitempty"`
	RcptTo                string    `json:"rcpt_to,omitempty"`
	RecvMethod            string    `json:"recv_method,omitempty"`
	RoutingDomain         string    `json:"routing_domain,omitempty"`
	SendingIP             string    `json:"sending_ip,omitempty"`
	Subject               string    `json:"subject,omitempty"`
	TemplateID            string    `json:"template_id,omitempty"`
	TemplateVersion       string    `json:"template_version,omitempty"`
	Timestamp             string    `json:"timestamp,omitempty"`
	TransmissionID        string    `json:"transmission_id,omitempty"`
	Type                  string    `json:"type,omitempty"`
	RawRcptTo             string    `json:"raw_rcpt_to,omitempty"`
	RecipientDomain       string    `json:"recipient_domain,omitempty"`
}

type TrackEvent struct {
	ClickTracking   bool      `json:"click_tracking,omitempty"`
	CustomerID      string    `json:"customer_id,omitempty"`
	DelvMethod      string    `json:"delv_method,omitempty"`
	EventID         string    `json:"event_id,omitempty"`
	FriendlyFrom    string    `json:"friendly_from,omitempty"`
	InjectionTime   time.Time `json:"injection_time,omitempty"`
	IPAddress       string    `json:"ip_address,omitempty"`
	IPPool          string    `json:"ip_pool,omitempty"`
	MessageID       string    `json:"message_id,omitempty"`
	MsgFrom         string    `json:"msg_from,omitempty"`
	MsgSize         string    `json:"msg_size,omitempty"`
	NumRetries      string    `json:"num_retries,omitempty"`
	OpenTracking    bool      `json:"open_tracking,omitempty"`
	OutboundTLS     string    `json:"outbound_tls,omitempty"`
	QueueTime       string    `json:"queue_time,omitempty"`
	RcptTo          string    `json:"rcpt_to,omitempty"`
	RecvMethod      string    `json:"recv_method,omitempty"`
	RoutingDomain   string    `json:"routing_domain,omitempty"`
	SendingIP       string    `json:"sending_ip,omitempty"`
	Subject         string    `json:"subject,omitempty"`
	TemplateID      string    `json:"template_id,omitempty"`
	TemplateVersion string    `json:"template_version,omitempty"`
	Timestamp       string    `json:"timestamp,omitempty"`
	TransmissionID  string    `json:"transmission_id,omitempty"`
	Type            string    `json:"type,omitempty"`
	RawRcptTo       string    `json:"raw_rcpt_to,omitempty"`
	RecipientDomain string    `json:"recipient_domain,omitempty"`
}

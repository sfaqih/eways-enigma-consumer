package structs

type ParamData interface{}

type Outbound struct {
	Messages   []OutMessage `json:"messages" validate:"required"`
	ClientID   int          `json:"client_id" validate:"required"`
	ClientCode string       `json:"client_code,omitempty"`
}

type OutboundConsumer struct {
	Messages      []OutMessage    `json:"messages" validate:"required"`
	ClientID      int             `json:"client_id" validate:"required"`
	ClientCode    string          `json:"client_code,omitempty"`
	VendorService []VendorService `json:"vendor_service,omitempty"`
}

type OutMessage struct {
	Flowid          int            `json:"flow_id,omitempty"`
	Channel         string         `json:"channel" validate:"required,oneof=Whatsapp BlastWhatsapp Email Email_Dev Instagram WhatsappTest BlastWhatsappTest Facebook BlastFacebook Telegram BlastTelegram Flow BlastWhatsappHeader Dummy DAMCORP DamcorpWhatsappBlast"`
	ChannelID       string         `json:"channel_id,omitempty"` //for MessageBird
	From            string         `json:"from,omitempty"`
	SesameID        string         `json:"sesame_id,omitempty"`
	MessageID       string         `json:"id,omitempty"`
	SesameMessageID string         `json:"sesame_message_id,omitempty"`
	Timestamp       int64          `json:"timestamp" validate:"required"`
	Title           string         `json:"title,omitempty"`
	Content         Content        `json:"content,omitempty"`
	To              string         `json:"to,omitempty"`
	Cc              string         `json:"cc,omitempty"`
	Bcc             string         `json:"bcc,omitempty"`
	WATemplateID    string         `json:"wa_template_id,omitempty"`
	ConversationID  string         `json:"conversation_id,omitempty"`
	Type            string         `json:"type" validate:"required,oneof=template hsm text audio document image video sticker flow interactive text_template button_template"`
	Hsm             Hsm            `json:"hsm,omitempty"`
	ParamDataEmail  ParamData      `json:"param_data_email,omitempty"`
	Name            string         `json:"name,omitempty"`
	Components      []Component    `json:"components,omitempty"`
	URLAttachmengts []*Attachments `json:"url_attachments,omitempty"`
	CampaignId      string         `json:"campaign_id,omitempty"`
	//CallBackUrl     string         `json:"call_back_url,omitempty"`
	CallBackAuth  CallBackAuth                  `json:"call_back_auth,omitempty"`
	TrxId         string                        `json:"trx_id,omitempty"`
	RecipientType string                        `json:"recipient_type,omitempty"`
	Interactive   *Damcorp_Outbound_Interactive `json:"interactive,omitempty"`
}

type OutboundFlow struct {
	Messages   []OutMessageFlow `json:"messages" validate:"required"`
	ClientID   int              `json:"client_id,omitempty"`
	ClientCode string           `json:"client_code,omitempty"`
}
type OutMessageFlow struct {
	Flowid   int    `json:"flow_id,omitempty"`
	Order    int    `json:"order,omitempty"`
	CustName string `json:"cust_name,omitempty"`

	Channel         string         `json:"channel,omitempty"`
	ChannelID       string         `json:"channel_id,omitempty"` //for MessageBird
	From            string         `json:"from,omitempty"`
	SesameID        string         `json:"sesame_id,omitempty"`
	MessageID       string         `json:"id,omitempty"`
	SesameMessageID string         `json:"sesame_message_id,omitempty"`
	Timestamp       int64          `json:"timestamp" validate:"required"`
	Title           string         `json:"title,omitempty"`
	Content         Content        `json:"content,omitempty"`
	To              string         `json:"to,omitempty"`
	Cc              string         `json:"cc,omitempty"`
	Bcc             string         `json:"bcc,omitempty"`
	WATemplateID    string         `json:"wa_template_id,omitempty"`
	ConversationID  string         `json:"conversation_id,omitempty"`
	Type            string         `json:"type" validate:"required,oneof=template hsm text audio document image video sticker flow"`
	Hsm             Hsm            `json:"hsm,omitempty"`
	ParamDataEmail  ParamData      `json:"param_data_email,omitempty"`
	Name            string         `json:"name,omitempty"`
	Components      []Component    `json:"components,omitempty"`
	URLAttachmengts []*Attachments `json:"url_attachments,omitempty"`
	CampaignId      string         `json:"campaign_id,omitempty"`
	//ReplyToMail     string         `json:"reply_to,omitempty"`
	CallBackAuth CallBackAuth `json:"call_back_auth,omitempty"`
	TrxId        string       `json:"trx_id,omitempty"`
}

type JTSWABlast struct {
	To         string    `json:"to" validate:"min=9,max=25,startswith=62"`
	Type       string    `json:"type"`
	Template   Tmp       `json:"template,omitempty"`
	Name       string    `json:"name,omitempty"`
	Hsm        Hsm       `json:"hsm,omitempty"`
	Components Component `json:"components,omitempty"`
}

type MSGBRD_WABlast struct {
	To        string  `json:"to" validate:"min=9,max=25,startswith=62"`
	Type      string  `json:"type"`
	ChannelID string  `json:"channelId,omitempty"`
	Content   Content `json:"content,omitempty"`
}

/*
type MSGBRD_Wa_Blast_Media_Header struct {
	To        string  `json:"to" validate:"min=9,max=25,startswith=62"`
	Type      string  `json:"type"`
	ChannelID string  `json:"channelId,omitempty"`
	Content   Content `json:"content,omitempty"`
}
*/

type MSGBRD_FBBlast struct {
	To        string  `json:"to" validate:"required"`
	Type      string  `json:"type"`
	ChannelID string  `json:"channelId,omitempty"`
	Content   Content `json:"content,omitempty"`
}

type MSGBRD_TeleBlast struct {
	To        string  `json:"to" validate:"required"`
	Type      string  `json:"type"`
	ChannelID string  `json:"channelId,omitempty"`
	Content   Content `json:"content,omitempty"`
}

type Content struct {
	Hsm         *Hsm           `json:"hsm,omitempty"`
	Text        *string        `json:"text,omitempty"`
	Audio       *OutAudio      `json:"audio,omitempty"`
	Image       *OutImage      `json:"image,omitempty"`
	Document    *OutDocument   `json:"document,omitempty"`
	Video       *OutVideo      `json:"video,omitempty"`
	Sticker     *OutSticker    `json:"sticker,omitempty"`
	TextEmail   *string        `json:"text_email,omitempty"`
	Name        string         `json:"name,omitempty"`
	Components  []Component    `json:"components,omitempty"`
	Interactive *SFInteractive `json:"interactive,omitempty"`
}

type Component struct {
	Type       string    `json:"type"` //Type for header or body
	Parameters []PrmType `json:"parameters"`
}

type PrmType struct {
	Type     string        `json:"type"` //Type for document, image, text, currency, date_time
	Text     string        `json:"text,omitempty"`
	Document MediaDoc      `json:"document,omitempty"`
	Image    MediaImage    `json:"image,omitempty"`
	Video    MediaVideo    `json:"video,omitempty"`
	Currency MediaCurrency `json:"currency,omitempty"`
	DateTime MediaDate     `json:"date_time,omitempty"`
}

type MediaDate struct {
	Fallback   string `json:"fallback_value"`
	DayOfWeek  string `json:"day_of_week"`
	DayOfMonth string `json:"day_of_month"`
	Year       int    `json:"year"`
	Month      int    `json:"month"`
	Hour       int    `json:"hour"`
	Minute     int    `json:"minute"`
	Timestamp  int64  `json:"timestamp"`
}

type MediaCurrency struct {
	Fallback string `json:"fallback_value"`
	Code     string `json:"code"`
	Amount   string `json:"amount_1000"`
}

type MediaImage struct {
	Link string `json:"link"`
	Name string `json:"name"`
	Url  string `json:"url,omitempty"`
}

type MediaVideo struct {
	Url string `json:"url,omitempty"`
}

type MediaDoc struct {
	Link     string   `json:"link"`
	Provider Provider `json:"provider"`
	Filename string   `json:"filename"`
}

type Provider struct {
	Name string `json:"name"`
}

// it will be used for WA Outbound non blast
type Tmp struct {
	Namespace string   `json:"namespace"`
	Lang      Language `json:"language"`
}

// this struct only used for template message
type Hsm struct {
	Namespace    string   `json:"namespace"`
	TemplateName string   `json:"templateName,omitempty"`
	ElementName  string   `json:"element_name,omitempty"`
	Lang         Language `json:"language"`
	Localizable  []Params `json:"localizable_params,omitempty"`
	Params       []Params `json:"params,omitempty"`
}

type Params struct {
	Default string `json:"default,omitempty"`
}

type Language struct {
	Policy string `json:"policy"`
	Code   string `json:"code"`
}

type JTSOutWA struct {
	RecipientType string      `json:"recipient_type" validate:"default=individual"`
	To            string      `json:"to" validate:"min=9,max=25,startswith=62"`
	Type          string      `json:"type" validate:"required,oneof=text audio document image video sticker"`
	Text          OutText     `json:"text,omitempty"`
	Audio         OutAudio    `json:"audio,omitempty"`
	Document      OutDocument `json:"document,omitempty"`
	Video         OutVideo    `json:"video,omitempty"`
	Sticker       OutSticker  `json:"sticker,omitempty"`
}

type MSGBRD_OutWA struct {
	Type    string  `json:"type" validate:"required,oneof=text audio document image video sticker"`
	Content Content `json:"content,omitempty"`
	Source  Source  `json:"source,omitempty"`
}

type Source struct {
	ConversationID  string `json:"conversation_id,omitempty"`
	SesameMessageID string `json:"sesame_message_id,omitempty"`
	SesameID        string `json:"sesame_id,omitempty"`
}

type OutText struct {
	Body string `json:"body"`
}

type OutSticker struct {
	Caption string `json:"caption,omitempty"`
	Link    string `json:"url,omitempty"`
}

type OutLocation struct {
	Latitude  string `json:"latitude,omitempty"`
	Longitude string `json:"longitude,omitempty"`
}

type OutAudio struct {
	Caption string `json:"caption,omitempty"`
	Link    string `json:"url,omitempty"`
}

type OutVideo struct {
	Caption string `json:"caption,omitempty"`
	Link    string `json:"url,omitempty"`
}

type OutImage struct {
	Caption string `json:"caption,omitempty"`
	Link    string `json:"url,omitempty"`
}

type OutDocument struct {
	Caption string `json:"caption,omitempty"`
	Link    string `json:"url,omitempty"`
}

type Attachments struct {
	Location *OutLocation `json:"location,omitempty"`
	Image    *OutImage    `json:"image,omitempty"`
	Document *OutDocument `json:"document,omitempty"`
	Audio    *OutAudio    `json:"audio,omitempty"`
	Video    *OutVideo    `json:"video,omitempty"`
}

//=============================================================================

type MTARGET_Email struct {
	AccessToken string    `json:"accessToken" validate:"required"`
	From        string    `json:"from" validate:"required"`
	To          []string  `json:"to" validate:"required"`
	Cc          []string  `json:"cc,omitempty"`
	Bcc         []string  `json:"bcc,omitempty"`
	Labels      []string  `json:"labels,omitempty"`
	Subject     string    `json:"subject,omitempty"`
	Content     *string   `json:"content,omitempty"`
	TemplateId  *string   `json:"templateId,omitempty"`
	Data        ParamData `json:"data,omitempty"`
}

type MAILJET_Email struct {
	Messages []MAILJET_Messages `json:"messages" validate:"required"`
}
type MAILJET_Messages struct {
	From     MAILJET_From `json:"from,omitempty"`
	To       []MAILJET_To `json:"to" validate:"required"`
	Subject  string       `json:"subject,omitempty"`
	TextPart string       `json:"TextPart,omitempty"`
	HTMLPart string       `json:"HTMLPart,omitempty"`
}

type MAILJET_From struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

type MAILJET_To struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

// =============================================================================
type SubstitutionData interface{}
type SPARKPOST_Email struct {
	CampaignId          string                `json:"campaign_id,omitempty"`
	SparkPostRecipients []SPARKPOST_Recipient `json:"recipients" validate:"required"`
	SparkPostContent    SPARKPOST_Content     `json:"content,omitempty"`
}
type SPARKPOST_Content struct {
	From       SPARKPOST_From               `json:"from,omitempty"`
	Subject    string                       `json:"subject,omitempty"`
	Html       string                       `json:"html,omitempty"`
	Text       string                       `json:"text,omitempty"`
	TemplateId string                       `json:"template_id,omitempty"`
	ReplyTo    string                       `json:"reply_to,omitempty"`
	HeadersCc  SPARKPOST_Content_Headers_Cc `json:"cc,omitempty"`
}
type SPARKPOST_From struct {
	Email     string `json:"email,omitempty"`
	FromAlias string `json:"name,omitempty"`
}
type SPARKPOST_Content_Headers_Cc struct {
	CC string `json:"cc,omitempty"`
}
type SPARKPOST_Recipient struct {
	//Address string `json:"address,omitempty"`
	//Address SPARKPOST_Address_inf `json:"address,omitempty"`
	Address          interface{}      `json:"address"`
	SubstitutionData SubstitutionData `json:"substitution_data,omitempty"`
}

type SPARKPOST_To struct {
	Email string `json:"email,omitempty"`
}
type SPARKPOST_Bcc struct {
	Email string `json:"email,omitempty"`
	Bcc   string `json:"header_to,omitempty"`
}

//=========================================================
// MSGBRD MEDIA HEADER

type MSGBRD_Wa_Blast_Media_Header struct {
	Content MSGBRD_Content_Media_Header `json:"content,omitempty"`
	From    string                      `json:"from"`
	To      string                      `json:"to" validate:"min=9,max=25,startswith=62"`
	Type    string                      `json:"type"`
}

type MSGBRD_Content_Media_Header struct {
	Hsm       *HsmMediaHeader `json:"hsm,omitempty"`
	Text      *string         `json:"text,omitempty"`
	Audio     *OutAudio       `json:"audio,omitempty"`
	Image     *OutImage       `json:"image,omitempty"`
	Document  *OutDocument    `json:"document,omitempty"`
	Video     *OutVideo       `json:"video,omitempty"`
	Sticker   *OutSticker     `json:"sticker,omitempty"`
	TextEmail *string         `json:"text_email,omitempty"`
	Name      string          `json:"name,omitempty"`
	//Components []Component     `json:"components,omitempty"`
}

type HsmMediaHeader struct {
	Components   []ComponentMediaHeader `json:"components"`
	Lang         Language               `json:"language"`
	Namespace    string                 `json:"namespace"`
	TemplateName string                 `json:"templateName"`
}

type ComponentMediaHeader struct {
	Type       string    `json:"type"` //Type for header or body
	Parameters []PrmType `json:"parameters"`
}

// type VendorResponse interface{}
type RspOutbound struct {
	Outbound       interface{} `json:"outbound,omitempty"`
	VendorResponse interface{} `json:"vendor_response,omitempty"`
}

type Damcorp_OutboundWA struct {
	To            string                        `json:"to,omitempty"`
	Type          string                        `json:"type,omitempty"`
	RecipientType string                        `json:"recipient_type,omitempty"`
	Text          Damcorp_Text_Body             `json:"text,omitempty"`
	Interactive   *Damcorp_Outbound_Interactive `json:"interactive,omitempty"`
	Image         Damcorp_Outbound_Image        `json:"image,omitempty"`
}

type Damcorp_Outbound_Image struct {
	Link string `json:"link,omitempty"`
}

type Damcorp_Text_Body struct {
	Body *string `json:"body,omitempty"`
}

type Damcorp_Outbound_Interactive struct {
	Type   *string                     `json:"type,omitempty"`
	Body   *Damcorp_Interactive_Body   `json:"body,omitempty"`
	Action *Damcorp_Interactive_Action `json:"action,omitempty"`
}

type Damcorp_Interactive_Body struct {
	Text *string `json:"text,omitempty"`
}

type Damcorp_Interactive_Action struct {
	Button   *string                  `json:"button,omitempty"`
	Sections *[]Damcorp_Section       `json:"sections,omitempty"`
	Buttons  *[]Damcorp_Action_Button `json:"buttons,omitempty"`
}

type Damcorp_Action_Button struct {
	Type  *string                      `json:"type,omitempty"`
	Reply *Damcorp_Action_Button_reply `json:"reply,omitempty"`
}

type Damcorp_Action_Button_reply struct {
	Id    *string `json:"id,omitempty"`
	Title *string `json:"title,omitempty"`
}

type Damcorp_Section struct {
	Title *string                `json:"title,omitempty"`
	Rows  *[]Damcorp_Section_Row `json:"rows,omitempty"`
}

type Damcorp_Section_Row struct {
	Id          *string `json:"id,omitempty"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
}

type Damcorp_Response struct {
	Contacts []Damcorp_Response_Contact `json:"contacts,omitempty"`
	Messages []Damcorp_Response_Message `json:"messages,omitempty"`
	Errors   []Damcorp_Response_Error   `json:"errors,omitempty"`
	Meta     Damcorp_Meta               `json:"meta,omitempty"`
}

type Damcorp_Response_Contact struct {
	Input string `json:"input,omitempty"`
	WAId  string `json:"wa_id,omitempty"`
}

type Damcorp_Response_Message struct {
	Id string `json:"id,omitempty"`
}

type Damcorp_Meta struct {
	ApiStatus string `json:"api_status,omitempty"`
	Version   string `json:"version,omitempty"`
}

type Damcorp_Close_Session struct {
	Msisdn string `json:"msisdn"`
}

type Damcorp_Close_Session_Request struct {
	ClientID int    `json:"client_id" validate:"required"`
	Msisdn   string `json:"msisdn" validate:"min=9,max=25,startswith=62"`
}

type Damcorp_Close_Session_Response struct {
	Success bool   `json:"success,omitempty"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type Damcorp_Response_Error struct {
	Code    int    `json:"code,omitempty"`
	Title   string `json:"title,omitempty"`
	Details string `json:"details,omitempty"`
}

type Damcorp_WA_Blast struct {
	TemplateName string   `json:"template_name"`
	Type         string   `json:"type"`
	Params       []string `json:"params"`
	Destination  []string `json:"destination"`
}

type OutboundBulkLog struct {
	Id            int    `json:"id,omitempty"`
	ClientId      int    `json:"client_id,omitempty"`
	RequestBody   string `json:"request_body,omitempty"`
	Status        int    `json:"status,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	ProcessAt     string `json:"process_at,omitempty"`
	LastUpdatedAt string `json:"last_updated_at,omitempty"`
}

/*
	SLEEKFLOW STRUCT
*/

type SleekFlowMessage struct {
	Channel         *string                   `json:"channel,omitempty"`
	From            *string                   `json:"from,omitempty"`
	To              *string                   `json:"to,omitempty"`
	MessageType     *string                   `json:"messageType,omitempty"`
	MessageContent  *string                   `json:"messageContent,omitempty"`
	ExtendedMessage *SleekFlowExtendedMessage `json:"extendedMessage,omitempty"`
}

type SleekFlowExtendedMessage struct {
	WhatsappCloudApiTemplateMessageObject *SFTemplateMessageObject `json:"WhatsappCloudApiTemplateMessageObject,omitempty"`
	WhatsappCloudApiInteractiveObject     *SFInteractive           `json:"whatsappCloudApiInteractiveObject,omitempty"`
}

type SFTemplateMessageObject struct {
	TemplateName *string        `json:"templateName,omitempty"`
	Language     *string        `json:"language,omitempty"`
	Components   []*SFComponent `json:"components,omitempty"`
}

type SFInteractive struct {
	Type   *string              `json:"type,omitempty"`
	Header *SFInteractiveList   `json:"header,omitempty"`
	Body   *SFInteractiveList   `json:"body,omitempty"`
	Footer *SFInteractiveList   `json:"footer,omitempty"`
	Action *SFInteractiveAction `json:"action,omitempty"`
}

type SFInteractiveList struct {
	Text *string `json:"text,omitempty"`
	Type *string `json:"type,omitempty"`
}

type SFInteractiveAction struct {
	Button   *string                 `json:"button,omitempty"`
	Sections []*SFInteractiveSection `json:"sections,omitempty"`
}

type SFInteractiveSection struct {
	Title *string                    `json:"title,omitempty"`
	Rows  []*SFInteractiveSectionRow `json:"rows,omitempty"`
}

type SFInteractiveSectionRow struct {
	Id          *string `json:"id,omitempty"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
}

type SFComponent struct {
	Type       string    `json:"type,omitempty"`
	Parameters []*SFParam `json:"parameters,omitempty"`
}

type SFParam struct {
	Type  string     `json:"type,omitempty"`
	Text  string     `json:"text,omitempty"`
	Image *ParamImage `json:"image,omitempty"`
}

type ParamImage struct {
	Link *string `json:"link,omitempty"`
}

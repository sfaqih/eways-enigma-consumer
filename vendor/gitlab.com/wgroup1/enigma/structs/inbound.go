package structs

import "time"

/*
type Inbounds struct {
	Messages []InMessage `json:"messages" validate:"required"`
	ClientID int         `json:"client_id" validate:"required"`
}


type InMessage struct {
	Channel         string         `json:"channel" validate:"required,oneof=Whatsapp BlastWhatsapp Email Instagram Facebook WhatsappTest BlastWhatsappTest"`
	ChannelID       string         `json:"channel_id,omitempty"` //for MessageBird
	From            string         `json:"from,omitempty"`
	SesameID        string         `json:"sesame_id" validate:"required"`
	MessageID       string         `json:"message_id,omitempty"`
	Timestamp       int64          `json:"timestamp" validate:"required"`
	Title           string         `json:"title,omitempty"`
	Content         Content        `json:"content,omitempty"`
	To              string         `json:"to,omitempty"`
	Cc              string         `json:"cc,omitempty"`
	Bcc             string         `json:"bcc,omitempty"`
	WATemplateID    string         `json:"wa_template_id,omitempty"`
	ConversationID  string         `json:"conversation_id,omitempty"`
	Type            string         `json:"type" validate:"required,oneof=template hsm text audio document image video sticker"`
	Hsm             Hsm            `json:"hsm,omitempty"`
	Template        Tmp            `json:"template,omitempty"`
	Name            string         `json:"name,omitempty"`
	Components      Component      `json:"components,omitempty"`
	URLAttachmengts []*Attachments `json:"url_attachments,omitempty"`
}
*/
type InMessage struct {
	Inbounds []Inbound `json:"Inbound" validate:"required"`
}

type Inbound struct {
	Contact      ContactInbound      `json:"contact"`
	Conversation ConversationInbound `json:"conversation"`
	Message      MessageInbound      `json:"message"`
	Type         string              `json:"type"`
	MessageID    string              `json:"message_id"`
	ClientID     int                 `json:"client_id"`
}

type Archived struct {
	ConversationID string `json:"conversation_id"`
	ClientID       int    `json:"client_id"`
}

type ArchStatus struct {
	Status string `json:"status"`
}

type ContactInbound struct {
	ID            string `json:"id"`
	Href          string `json:"href"`
	Msisdn        string `json:"msisdn"`
	Displayname   string `json:"displayName"`
	Firstname     string `json:"firstName"`
	Lastname      string `json:"lastName"`
	Customdetails struct {
	} `json:"customDetails"`
	Attributes struct {
	} `json:"attributes"`
	Createddatetime time.Time `json:"createdDatetime"`
	Updateddatetime time.Time `json:"updatedDatetime"`
}

type ConversationInbound struct {
	ID                   string    `json:"id"`
	Contactid            string    `json:"contactId"`
	Status               string    `json:"status"`
	Createddatetime      time.Time `json:"createdDatetime"`
	Updateddatetime      time.Time `json:"updatedDatetime"`
	Lastreceiveddatetime time.Time `json:"lastReceivedDatetime"`
	Lastusedchannelid    string    `json:"lastUsedChannelId"`
	Messages             struct {
		Totalcount int    `json:"totalCount"`
		Href       string `json:"href"`
	} `json:"messages"`
}

type MessageInbound struct {
	ID             string `json:"id"`
	Conversationid string `json:"conversationId"`
	Platform       string `json:"platform"`
	To             string `json:"to"`
	From           string `json:"from"`
	Cc             string `json:"cc"`
	Bcc            string `json:"bcc"`
	Subject        string `json:"subject"`
	Channelid      string `json:"channelId"`
	Type           string `json:"type"`
	Content        struct {
		Text  string `json:"text"`
		Image struct {
			URL     string `json:"url"`
			Caption string `json:"caption"`
		} `json:"image"`
		Video struct {
			URL string `json:"url"`
		} `json:"video"`
	} `json:"content"`
	Direction       string    `json:"direction"`
	Status          string    `json:"status"`
	Createddatetime time.Time `json:"createdDatetime"`
	Updateddatetime time.Time `json:"updatedDatetime"`
}

type InternalInboundWa struct {
	//	RecipientType string      `json:"recipient_type" validate:"default=individual"`
	Contactid string `json:"contactid"`
	//ContactJsonPlain      string `json:"contact"`
	ContactJsonPlain ContactInbound `json:"contactinbound"`
	//ConversationJsonPlain string    `json:"conversation"`
	ConversationJsonPlain ConversationInbound `json:"conversation"`
	Message               MessageInbound      `json:"messageinbound"`
	ConversationId        string              `json:"conversationid"`
	Msisdn                string              `json:"msisdn"`
	MessageId             string              `json:"messageid"`
	ChannelId             string              `json:"channelid"`
	//JsonPlainAll          string         `json:"jsonplain"`
	JsonPlainAll Inbound `json:"Inbound"`
	Platform     string  `json:"platform"`
}

type InternalInboundSesameAll struct {
	Channel        string `json:"channel,omitempty"`
	ContactID      string `json:"contact_id,omitempty"`
	ContactName    string `json:"contact_name,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
	MessageID      string `json:"message_id,omitempty"`
	To             string `json:"to,omitempty"`
	From           string `json:"from,omitempty"`
	Cc             string `json:"cc,omitempty"`
	Bcc            string `json:"bcc,omitempty"`
	Subject        string `json:"subject,omitempty"`
	Type           string `json:"type,omitempty"`
	//Content         []byte    `json:"content,omitempty"`
	Content         ContentType `json:"content,omitempty"`
	Status          string      `json:"status,omitempty"`
	Createddatetime time.Time   `json:"createdDatetime"`
	Updateddatetime time.Time   `json:"updatedDatetime"`
}
type ContentType interface{}

type ContentAll struct {
	Text  string                  `json:"text,omitempty"`
	Image ContentImageSesameChild `json:"image,omitempty"`
}

type ContentTextSesame struct {
	Text string `json:"text,omitempty"`
}

type ContentImageSesame struct {
	Text  string                  `json:"text,omitempty"`
	Image ContentImageSesameChild `json:"image,omitempty"`
}
type ContentFileSesame struct {
	Text string                 `json:"text,omitempty"`
	File ContentFileSesameChild `json:"file,omitempty"`
}
type ContentVideoSesame struct {
	Text  string                  `json:"text,omitempty"`
	Video ContentVideoSesameChild `json:"video,omitempty"`
}
type ContentAudioSesame struct {
	Text  string                  `json:"text,omitempty"`
	Audio ContentAudioSesameChild `json:"audio,omitempty"`
}

type ContentImageSesameChild struct {
	Text    string `json:"text,omitempty"`
	Url     string `json:"url,omitempty"`
	Caption string `json:"caption,omitempty"`
}
type ContentFileSesameChild struct {
	Text    string `json:"text,omitempty"`
	Url     string `json:"url,omitempty"`
	Caption string `json:"caption,omitempty"`
}
type ContentVideoSesameChild struct {
	Text    string `json:"text,omitempty"`
	Url     string `json:"url,omitempty"`
	Caption string `json:"caption,omitempty"`
}
type ContentAudioSesameChild struct {
	Text    string `json:"text,omitempty"`
	Url     string `json:"url,omitempty"`
	Caption string `json:"caption,omitempty"`
}

/*
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

type Content struct {
	Hsm      *Hsm         `json:"hsm,omitempty"`
	Text     string       `json:"text,omitempty"`
	Audio    *OutAudio    `json:"audio,omitempty"`
	Image    *OutImage    `json:"image,omitempty"`
	Document *OutDocument `json:"document,omitempty"`
	Video    *OutVideo    `json:"video,omitempty"`
	Sticker  *OutSticker  `json:"sticker,omitempty"`
}

type Component struct {
	Type       string  `json:"type"` //Type for header or body
	Parameters PrmType `json:"parameters"`
}

type PrmType struct {
	Type     string        `json:"type"` //Type for document, image, text, currency, date_time
	Text     string        `json:"text,omitempty"`
	Document MediaDoc      `json:"document,omitempty"`
	Image    MediaImage    `json:"image,omitempty"`
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
}

type MediaDoc struct {
	Link     string   `json:"link"`
	Provider Provider `json:"provider"`
	Filename string   `json:"filename"`
}

type Provider struct {
	Name string `json:"name"`
}

//it will be used for WA Outbound non blast
type Tmp struct {
	Namespace string   `json:"namespace"`
	Lang      Language `json:"language"`
}

//this struct only used for template message
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
	ConversationID string `json:"conversation_id,omitempty"`
	MessageID      string `json:"message_id,omitempty"`
	SesameID       string `json:"sesame_id,omitempty"`
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

*/

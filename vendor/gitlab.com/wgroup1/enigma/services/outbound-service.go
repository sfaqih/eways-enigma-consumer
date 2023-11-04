package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gitlab.com/wgroup1/enigma/common"
	"gitlab.com/wgroup1/enigma/interfaces"
	"gitlab.com/wgroup1/enigma/repositories/mysql"
	"gitlab.com/wgroup1/enigma/structs"
	"gopkg.in/go-playground/validator.v9"
)

type OutboundService interface {
	CreateOutbound(outm structs.OutMessage, reqStr string, respStr string, cli int) *structs.ErrorMessage
	Validate(outm *structs.OutMessage) (*structs.OutMessage, *structs.ErrorMessage)
	GetToken(clientID int) ([]structs.VendorService, *structs.ErrorMessage)
	SendOutboundSgl(outm structs.OutMessage, vs map[string]structs.VendorService, clientID int) *structs.ErrorMessage
	SendOutboundBulk(outm structs.OutMessage, vs map[string]structs.VendorService, clientID int) structs.HTTPRequest

	ValidateFlow(outmflow *structs.OutMessageFlow) (*structs.OutMessageFlow, *structs.ErrorMessage)
	SendOutboundFlowSgl(outbflow structs.OutboundFlow, vs map[string]structs.VendorService, clientID int) *structs.ErrorMessage
	GetMessageStatusService()
	CloseSessionWA(cs structs.Damcorp_Close_Session, vs map[string]structs.VendorService) *structs.ErrorMessage
	CreateOutboundBulkLog(outbound []byte, cli int) (*structs.OutboundBulkLog, error)
	InsertOutboundBulk(outm []*structs.HTTPRequest, cli int)
}

type outboundService struct{
	outboundRepo interfaces.OutboundRepository
	DB *sql.DB
}


func NewOutboundService(repository interfaces.OutboundRepository, db *sql.DB) OutboundService {
	return &outboundService{
		outboundRepo: repository,
		DB: db,
	}
}

func (oub *outboundService) GetToken(clientID int) ([]structs.VendorService, *structs.ErrorMessage) {
	return oub.outboundRepo.GetToken(clientID)
}

func (oub *outboundService) CreateOutbound(outm structs.OutMessage, reqStr string, respStr string, cli int) *structs.ErrorMessage {
	return oub.outboundRepo.CreateOutbound(outm, reqStr, respStr, cli)
}

// Will be used by read status scheduler
func (oub *outboundService) GetMessageStatusService() {
	// get all status (pending, sent, delivered)
	stats, errs := oub.outboundRepo.GetMessageStatus()

	// -- fetch results
	s := make(map[int]structs.Status)
	for j := range stats {
		s[j] = stats[j]
	}

	if stats == nil {
		return
	}
	if errs != nil {
		fmt.Println(errs)
	}

	// get token and session for WhatsappStatus
	vss, _ := oub.outboundRepo.GetTokenStatus("MSGBRD", "WhatsappStatus")

	t := make(map[string]structs.VendorService)
	for k := range vss {
		t[vss[k].Channel] = vss[k]
	}

	//get Status from vendor
	sNew, body, err := sendRequestStatus(t, s, oub.DB)
	if err != nil {
		fmt.Println(err.Error())
	}

	// START HIT STATUS TO SESAME
	// get token and session for WhatsappStatus
	vsstosesame, _ := oub.outboundRepo.GetTokenStatus("SESAME", "WhatsappStatus")

	ttosesame := make(map[string]structs.VendorService)
	for ktosesame := range vsstosesame {
		ttosesame[vsstosesame[ktosesame].Channel] = vsstosesame[ktosesame]
	}

	// END HIT STATUS TO SESAME

	for k := range sNew {
		//fmt.Println("newStat:", sNew[k].NewStatus)
		//fmt.Println("channel:", sNew[k].Channel)
		//schannel := strings.ToUpper(sNew[k].Channel)
		switch sNew[k].NewStatus {
		case "sent":
			stats[k].NewStatusInt = 1
			_, err := oub.outboundRepo.UpdateStatusSent(stats[k], body)
			if err != nil {
				fmt.Println(err.Error())
			}

		case "delivered":
			stats[k].NewStatusInt = 2
			_, err := oub.outboundRepo.UpdateStatusDelivered(stats[k], body)
			if err != nil {
				fmt.Println(err.Error())
			}
		case "read":
			stats[k].NewStatusInt = 3
			_, err := oub.outboundRepo.UpdateStatusRead(stats[k], body)
			if err != nil {
				fmt.Println(err.Error())
			}
		case "failed":
			stats[k].NewStatusInt = 99
			_, err := oub.outboundRepo.UpdateStatusFailed(stats[k], body)
			if err != nil {
				fmt.Println(err.Error())
			}
		case "pending":
			_, err := oub.outboundRepo.UpdateStatusPending(stats[k], body)
			if err != nil {
				fmt.Println(err.Error())
			}
		case "rejected":
			_, err := oub.outboundRepo.UpdateStatusRejected(stats[k], body)
			if err != nil {
				fmt.Println(err.Error())
			}
		case "received":
			_, err := oub.outboundRepo.UpdateStatusReceived(stats[k], body)
			if err != nil {
				fmt.Println(err.Error())
			}
		case "404":
			stats[k].NewStatusInt = 99
			_, err := oub.outboundRepo.UpdateStatusFailed(stats[k], body)
			if err != nil {
				fmt.Println(err.Error())
			}
			/*
				case "":
					if schannel == "FACEBOOK" {
						_, err := outboundRepo.UpdateStatusPending(stats[k], body)
						if err != nil {
							fmt.Println(err.Error())
						}
					}
			*/
		case "":
			stats[k].NewStatusInt = 99
			_, err := oub.outboundRepo.UpdateStatusFailed(stats[k], body)
			if err != nil {
				fmt.Println(err.Error())
			}
		default:
			fmt.Println("No information available for message_id " + stats[k].MessageID)
			//fmt.Println(sNew[k].NewStatus)
			//fmt.Println(stats[k].MessageID)

		}

		//PUSH TATUS TO SESAME
		_, errtosesame := sendPushStatusToSesame(ttosesame, sNew, body, oub.DB)
		if errtosesame != nil {
			fmt.Println(errtosesame.Error())
		}
	}
}

func sendRequestStatus(vs map[string]structs.VendorService, s map[int]structs.Status, DB *sql.DB) (map[int]structs.Status, []byte, error) {
	var respStat structs.MBRespStatus
	var respStatusCode *http.Response
	var uri string
	var apiID int
	var body []byte
	var err error

	for k, e := range s {
		switch e.Channel {
		case "Whatsapp":
			uri = sendWAStatus(s[k].MessageID, vs)
			_, respStatusCode, body, apiID, err = common.HitAPI(vs["WhatsappStatus"].URL+uri, "", vs["WhatsappStatus"].Method, vs["WhatsappStatus"].HeaderPrefix+" "+vs["WhatsappStatus"].Token, vs["WhatsappStatus"].Alias, time.Duration(0), DB)
			if err != nil {
				fmt.Println(err.Error())
				return nil, body, err
			}
			//_ = respStatusCode
			//fmt.Println(body)
			json.Unmarshal(body, &respStat)
			//fmt.Println(&respStat)
			if respStatusCode.StatusCode == 404 {
				//str1 := strconv.Itoa(respStatusCode.StatusCode)
				respStat.Status = strconv.Itoa(respStatusCode.StatusCode)
			}

			e.NewStatus = respStat.Status
			e.ApiID = apiID
			s[k] = e
		}
	}

	return s, body, nil
}

// Func to send WA to Jatis
func sendWAStatus(msgId string, vs map[string]structs.VendorService) string {
	var uri string
	alias := vs["WhatsappStatus"].Alias

	if alias == "MSGBRD" {
		uri = strings.Replace(vs["WhatsappStatus"].URI, "{{msg_id}}", msgId, -1)
	} else {
		uri = vs["WhatsappStatus"].URI
	}

	return uri
}

// Func to send WA status to SESAME
func sendPushStatusToSesame(vs map[string]structs.VendorService, s map[int]structs.Status, body []byte, DB *sql.DB) (bool, error) {
	var respStatusCode *http.Response
	var uri string
	var apiID int
	//var body []byte
	var err error
	var sds structs.StatusDeliverySesame

	for k, e := range s {
		sds.MessageID = s[k].MessageID
		sds.Status = s[k].NewStatus

		sdsByte, _ := json.Marshal(sds)

		//uri = sendWAStatus(s[k].MessageID, vs)
		uri = sendWAStatus("", vs)
		_, respStatusCode, _, apiID, err = common.HitAPI(vs["WhatsappStatus"].URL+uri, string(sdsByte), vs["WhatsappStatus"].Method, vs["WhatsappStatus"].HeaderPrefix+" "+vs["WhatsappStatus"].Token, vs["WhatsappStatus"].Alias, time.Duration(0), DB)
		if err != nil {
			fmt.Println(err.Error())
			return false, err
		}
		_ = respStatusCode
		_ = e
		_ = apiID
	}
	return true, nil
}

// Func to send WA to Jatis/MessageBird
func sendWAConv(outm structs.OutMessage, vs map[string]structs.VendorService) (string, string) {
	//var jsonStr string
	var jsonMsg []byte
	var uri string
	alias := vs["Whatsapp"].Alias

	if alias == "JTS" {
		var jts structs.JTSOutWA

		jts.RecipientType = "individual"
		jts.Type = "text"
		jts.To = outm.To
		//please note this, Jatis content still haven't been setup yet, please get back to this line after jatis ready with conversation API
		//jts.Text.Body = outm.Content
		//return the body string
		uri = vs["Whatsapp"].URI
		jsonMsg, _ = json.Marshal(jts)
	} else if alias == "MSGBRD" {
		var brd structs.MSGBRD_OutWA
		brd.Type = outm.Type
		brd.Content = outm.Content
		brd.Source.ConversationID = outm.ConversationID
		brd.Source.SesameMessageID = outm.SesameMessageID
		brd.Source.SesameID = outm.SesameID
		uri = strings.Replace(vs["Whatsapp"].URI, "{{conv_id}}", outm.ConversationID, -1)

		jsonMsg, _ = json.Marshal(brd)
	}

	return string(jsonMsg), uri
}

// Func to send FB to Jatis/MessageBird
func sendFBConv(outm structs.OutMessage, vs map[string]structs.VendorService) (string, string) {
	//var jsonStr string
	var jsonMsg []byte
	var uri string
	alias := vs["Facebook"].Alias

	if alias == "MSGBRD" {
		var brd structs.MSGBRD_OutWA
		brd.Type = outm.Type
		brd.Content = outm.Content
		brd.Source.ConversationID = outm.ConversationID
		brd.Source.SesameMessageID = outm.SesameMessageID
		brd.Source.SesameID = outm.SesameID
		uri = strings.Replace(vs["Facebook"].URI, "{{conv_id}}", outm.ConversationID, -1)

		jsonMsg, _ = json.Marshal(brd)
	}

	return string(jsonMsg), uri
}

// Func to send Telegram to Jatis/MessageBird
func sendTeleConv(outm structs.OutMessage, vs map[string]structs.VendorService) (string, string) {
	//var jsonStr string
	var jsonMsg []byte
	var uri string
	alias := vs["Telegram"].Alias

	if alias == "MSGBRD" {
		var brd structs.MSGBRD_OutWA
		brd.Type = outm.Type
		brd.Content = outm.Content
		brd.Source.ConversationID = outm.ConversationID
		brd.Source.SesameMessageID = outm.SesameMessageID
		brd.Source.SesameID = outm.SesameID
		uri = strings.Replace(vs["Telegram"].URI, "{{conv_id}}", outm.ConversationID, -1)

		jsonMsg, _ = json.Marshal(brd)
	}

	return string(jsonMsg), uri
}

func sendWABlast(outm structs.OutMessage, vs map[string]structs.VendorService) string {
	var prms structs.Params
	var jsonMsg []byte

	alias := vs["BlastWhatsapp"].Alias

	if alias == "JTS" {
		var jts structs.JTSWABlast
		jts.To = outm.To
		jts.Type = outm.Type
		if outm.Type == "hsm" {
			jts.Hsm.Namespace = outm.Hsm.Namespace
			jts.Hsm.ElementName = outm.WATemplateID
			jts.Hsm.Lang.Policy = outm.Hsm.Lang.Policy
			jts.Hsm.Lang.Code = outm.Hsm.Lang.Code

			if outm.Hsm.Localizable != nil {
				for j := range outm.Hsm.Localizable {
					prms.Default = outm.Hsm.Localizable[j].Default
					jts.Hsm.Localizable = append(jts.Hsm.Localizable, prms)
					fmt.Println("default: ", prms.Default)
				}
			}
		} else if outm.Type == "template" {
			jts.Name = outm.Name
			jts.Components = outm.Components[0]
		}
		jsonMsg, _ = json.Marshal(jts)
	} else if alias == "MSGBRD" {
		var brd structs.MSGBRD_WABlast

		brd.To = outm.To
		brd.Type = outm.Type
		brd.ChannelID = outm.ChannelID
		if outm.Type == "hsm" {
			brd.Content.Hsm = &outm.Hsm
			brd.Content.Hsm.TemplateName = outm.WATemplateID

			if outm.Hsm.Localizable != nil {
				for j := range outm.Hsm.Localizable {
					prms.Default = outm.Hsm.Localizable[j].Default
					brd.Content.Hsm.Params = append(brd.Content.Hsm.Params, prms)
					//fmt.Println("default: ", prms.Default)
				}
			}
			jsonMsg, _ = json.Marshal(brd)

		}

	}

	return string(jsonMsg)
}

func sendWABlastWithHeader(outm structs.OutMessage, vs map[string]structs.VendorService) string {
	//var prms structs.Params
	var jsonMsg []byte

	if outm.Type == "hsm" {

		var prmstype []structs.PrmType
		var componentMediaHeaders []structs.ComponentMediaHeader
		var componentMediaHeader structs.ComponentMediaHeader
		var hsmmediaheader structs.HsmMediaHeader
		var msgbrdcontentmediaheader structs.MSGBRD_Content_Media_Header
		var msgbrdwablastmediaheader structs.MSGBRD_Wa_Blast_Media_Header

		if outm.Components != nil {
			for j := range outm.Components {
				componenttype := outm.Components[j].Type
				componentMediaHeader.Type = componenttype
				prmstype = outm.Components[j].Parameters
				componentMediaHeader.Parameters = prmstype
				componentMediaHeaders = append(componentMediaHeaders, componentMediaHeader)
			}
		}
		hsmmediaheader.Components = componentMediaHeaders
		hsmmediaheader.Lang = outm.Hsm.Lang
		hsmmediaheader.Namespace = outm.Hsm.Namespace
		hsmmediaheader.TemplateName = outm.WATemplateID
		//fmt.Println(hsmmediaheader)

		msgbrdcontentmediaheader.Hsm = &hsmmediaheader
		msgbrdwablastmediaheader.Content = msgbrdcontentmediaheader
		msgbrdwablastmediaheader.From = outm.From
		msgbrdwablastmediaheader.To = outm.To
		msgbrdwablastmediaheader.Type = outm.Type

		jsonMsg, _ = json.Marshal(msgbrdwablastmediaheader)
	}

	return string(jsonMsg)
}

func sendFBBlast(outm structs.OutMessage, vs map[string]structs.VendorService) string {
	var prms structs.Params
	var jsonMsg []byte

	alias := vs["BlastFacebook"].Alias

	if alias == "MSGBRD" {
		var brd structs.MSGBRD_FBBlast
		brd.To = outm.To
		brd.Type = outm.Type
		brd.ChannelID = outm.ChannelID
		switch outm.Type {
		case "hsm":
			brd.Content.Hsm.Namespace = outm.Hsm.Namespace
			brd.Content.Hsm.TemplateName = outm.WATemplateID
			brd.Content.Hsm.Lang.Policy = outm.Hsm.Lang.Policy
			brd.Content.Hsm.Lang.Code = outm.Hsm.Lang.Code

			if outm.Hsm.Localizable != nil {
				for j := range outm.Hsm.Localizable {
					prms.Default = outm.Hsm.Localizable[j].Default
					brd.Content.Hsm.Params = append(brd.Content.Hsm.Params, prms)
					fmt.Println("default: ", prms.Default)
				}
			}
		case "text":
			brd.Content = outm.Content
		case "image":
			brd.Content = outm.Content
		case "video":
			brd.Content = outm.Content
		case "audio":
			brd.Content = outm.Content
		}

		/*
			if outm.Type == "hsm" {
				brd.Content.Hsm.Namespace = outm.Hsm.Namespace
				brd.Content.Hsm.TemplateName = outm.WATemplateID
				brd.Content.Hsm.Lang.Policy = outm.Hsm.Lang.Policy
				brd.Content.Hsm.Lang.Code = outm.Hsm.Lang.Code

				if outm.Hsm.Localizable != nil {
					for j := range outm.Hsm.Localizable {
						prms.Default = outm.Hsm.Localizable[j].Default
						brd.Content.Hsm.Params = append(brd.Content.Hsm.Params, prms)
						fmt.Println("default: ", prms.Default)
					}
				}
			}
		*/
		jsonMsg, _ = json.Marshal(brd)
	}

	return string(jsonMsg)
}

func sendTeleBlast(outm structs.OutMessage, vs map[string]structs.VendorService) string {
	var prms structs.Params
	var jsonMsg []byte

	alias := vs["BlastTelegram"].Alias

	if alias == "MSGBRD" {
		var brd structs.MSGBRD_TeleBlast
		brd.To = outm.To
		brd.Type = outm.Type
		brd.ChannelID = outm.ChannelID
		switch outm.Type {
		case "hsm":
			brd.Content.Hsm.Namespace = outm.Hsm.Namespace
			brd.Content.Hsm.TemplateName = outm.WATemplateID
			brd.Content.Hsm.Lang.Policy = outm.Hsm.Lang.Policy
			brd.Content.Hsm.Lang.Code = outm.Hsm.Lang.Code

			if outm.Hsm.Localizable != nil {
				for j := range outm.Hsm.Localizable {
					prms.Default = outm.Hsm.Localizable[j].Default
					brd.Content.Hsm.Params = append(brd.Content.Hsm.Params, prms)
					fmt.Println("default: ", prms.Default)
				}
			}
		case "text":
			brd.Content = outm.Content
		case "image":
			brd.Content = outm.Content
		case "video":
			brd.Content = outm.Content
		case "audio":
			brd.Content = outm.Content
		}

		jsonMsg, _ = json.Marshal(brd)
	}

	return string(jsonMsg)
}

// func send email
func sendEmail(outm structs.OutMessage, vs map[string]structs.VendorService) string {
	var jsonMsg []byte

	alias := vs["Email"].Alias
	switch alias {
	case "MTARGET":
		var mtEmail = structs.MTARGET_Email{}

		mtEmail.AccessToken = vs["Email"].Token
		mtEmail.From = outm.From
		mtEmail.To = append(mtEmail.To, outm.To)
		mtEmail.Cc = append(mtEmail.Cc, outm.Cc)
		mtEmail.Bcc = append(mtEmail.Bcc, outm.Bcc)
		mtEmail.Labels = append(mtEmail.Labels, outm.Title)
		mtEmail.Subject = outm.Title
		//Tambahin if else cek type
		if outm.Type == "text" {
			mtEmail.Content = outm.Content.Text
		} else if outm.Type == "template" {
			mtEmail.TemplateId = outm.Content.Text
			mtEmail.Data = outm.ParamDataEmail
		}

		jsonMsg, _ = json.Marshal(mtEmail)

	case "MAIL_JET":

		var mjFrom = structs.MAILJET_From{}
		mjFrom.Name = ""
		mjFrom.Email = outm.From

		var mjTo = structs.MAILJET_To{}
		mjTo.Name = ""
		mjTo.Email = outm.To

		var mjMessage = structs.MAILJET_Messages{}
		mjMessage.From = mjFrom
		mjMessage.To = append(mjMessage.To, mjTo)
		mjMessage.Subject = outm.Title
		mjMessage.TextPart = *outm.Content.Text
		mjMessage.HTMLPart = *outm.Content.Text

		var mjEmail = structs.MAILJET_Email{}
		mjEmail.Messages = append(mjEmail.Messages, mjMessage)

		jsonMsg, _ = json.Marshal(mjEmail)

	case "SPARKPOST":

		var spEmail = structs.SPARKPOST_Email{}

		var spContent = structs.SPARKPOST_Content{}

		var spContentFrom = structs.SPARKPOST_From{}

		spContentFrom.Email = outm.From
		spContentFrom.FromAlias = outm.Content.Name // bisa di di gunakan sebagai nama alias sender
		spContent.From = spContentFrom
		spContent.ReplyTo = vs["Email"].ReplyTo

		spContent.Subject = outm.Title

		var spRecipient = structs.SPARKPOST_Recipient{}
		addr2 := map[string]interface{}{}
		addr3 := map[string]interface{}{}
		if outm.Cc != "" && outm.Bcc == "" {

			addr2["email"] = outm.Cc
			addr2["header_to"] = outm.To

			spRecipient.Address = addr2
			spEmail.SparkPostRecipients = append(spEmail.SparkPostRecipients, spRecipient)

			spRecipient.Address = outm.To

		} else if outm.Cc == "" && outm.Bcc != "" {
			addr2["email"] = outm.Bcc
			addr2["header_to"] = outm.To

			spRecipient.Address = addr2
			spEmail.SparkPostRecipients = append(spEmail.SparkPostRecipients, spRecipient)

			spRecipient.Address = outm.To

		} else if outm.Cc != "" && outm.Bcc != "" {
			addr2["email"] = outm.Cc
			addr2["header_to"] = outm.To

			spRecipient.Address = addr2
			spEmail.SparkPostRecipients = append(spEmail.SparkPostRecipients, spRecipient)

			addr3["email"] = outm.Bcc
			addr3["header_to"] = outm.To

			spRecipient.Address = addr3
			spEmail.SparkPostRecipients = append(spEmail.SparkPostRecipients, spRecipient)

			spRecipient.Address = outm.To

		} else if outm.Bcc == "" && outm.Cc == "" {
			spRecipient.Address = outm.To
			//spEmail.SparkPostRecipients = append(spEmail.SparkPostRecipients, spRecipient)
		}

		if outm.Type == "text" {
			spContent.Html = *outm.Content.Text
		} else if outm.Type == "template" {
			spContent.TemplateId = outm.Hsm.Namespace
			byteParamDataEmail, _ := json.Marshal(outm.ParamDataEmail)

			json.Unmarshal(byteParamDataEmail, &spRecipient.SubstitutionData)
		}
		spEmail.CampaignId = ""
		spEmail.SparkPostContent = spContent

		spEmail.SparkPostRecipients = append(spEmail.SparkPostRecipients, spRecipient)
		jsonMsg, _ = json.Marshal(spEmail)

	default:
		fmt.Println("No Provider Found")
	}

	return string(jsonMsg)
}

func sendDamCorpWa(outm structs.OutMessage) string {
	var dcorp structs.Damcorp_OutboundWA
	var jsonMsg []byte

	dcorp.To = outm.To
	dcorp.Type = outm.Type
	dcorp.RecipientType = "individual"

	switch dcorp.Type {
	case "text":
		dcorp.Text.Body = outm.Content.Text
	case "interactive":
		dcorp.RecipientType = outm.RecipientType
		dcorp.Interactive = outm.Interactive
	case "image":
		if outm.Content.Image.Link != "" {
			dcorp.Image.Link = outm.Content.Image.Link
			if outm.Content.Image.Caption == "" {
				outm.Content.Image.Caption = outm.Content.Image.Link
			}
		}
	}
	jsonMsg, _ = json.Marshal(dcorp)

	return string(jsonMsg)
}

func sendDamcorpWaBlast(outm structs.OutMessage) string {
	var waBlast structs.Damcorp_WA_Blast
	var jsonWaBlast []byte

	for i := range outm.Components {
		for _, j := range outm.Components[i].Parameters {
			waBlast.Params = append(waBlast.Params, j.Text)
		}
	}

	waBlast.TemplateName = outm.WATemplateID
	waBlast.Type = outm.Type
	waBlast.Destination = append(waBlast.Destination, outm.To)

	// waBlast.Params = 

	jsonWaBlast, _ = json.Marshal(waBlast) 

	return string(jsonWaBlast)
}

/*

	SLEEKFLOW MAPPING STRUCT

*/

func sleekflowReplyConversation(outm structs.OutMessage, from *string) string {
	var sf structs.SleekFlowMessage
	channel := "whatsappcloudapi"
	sf.Channel = &channel
	sf.To = &outm.To
	sf.MessageType = &outm.Type
	sf.MessageContent = outm.Content.Text
	sf.From = from

	jsonBlast, _ := json.Marshal(&sf)

	return string(jsonBlast)
}

func sleekflowSendInteractiveWA(outm structs.OutMessage, from *string) string {
	var sf structs.SleekFlowMessage
	var sfInteractive structs.SFInteractive
	var sfObject structs.SleekFlowExtendedMessage

	channel := "whatsappcloudapi"
	sf.Channel = &channel
	sf.To = &outm.To
	sf.From = from
	sf.MessageType = &outm.Type

	interactiveByte, _ := json.Marshal(outm.Content.Interactive)

	json.Unmarshal(interactiveByte, &sfInteractive)


	sfObject.WhatsappCloudApiInteractiveObject = &sfInteractive
	sf.ExtendedMessage = &sfObject

	jsonBlast, _ := json.Marshal(&sf)

	return string(jsonBlast)

}

func sleekflowBlastTemplatePlan(outm structs.OutMessage, from *string) string {
	var sf structs.SleekFlowMessage
	var sfTemplate structs.SFTemplateMessageObject
	var sfObject structs.SleekFlowExtendedMessage
	// var sfTemplateParams structs.SFTemplateParams
	channel := "whatsappcloudapi"
	sf.Channel = &channel
	sf.To = &outm.To
	sf.From = from
	sf.MessageType = &outm.Type
	sfTemplate.TemplateName = &outm.WATemplateID
	sfTemplate.Language = &outm.Hsm.Lang.Code
	
	if len(outm.Components) > 0 {
		for _, component := range outm.Components {
			var tempComp structs.SFComponent
			tempComp.Type = component.Type
			
			if len(component.Parameters) > 0 {
				for _, param := range component.Parameters {
					var tempParam structs.SFParam

					tempParam.Type = param.Type

					switch param.Type {
					case "image":
						var tempImage structs.ParamImage
						tempImage.Link = &param.Image.Link
						tempParam.Image = &tempImage
					default:
						tempParam.Text = param.Text
					}
					
					
					tempComp.Parameters = append(tempComp.Parameters, &tempParam)
				}
			}

			sfTemplate.Components = append(sfTemplate.Components, &tempComp)
			
		}
	}

	// sfTemplate.Componens = append(sfTemplate.Componens, sfTemplateParams)
	sfObject.WhatsappCloudApiTemplateMessageObject = &sfTemplate
	sf.ExtendedMessage = &sfObject
	
	jsonBlast, _ := json.Marshal(&sf)

	return string(jsonBlast)
}

func (oub *outboundService) SendOutboundSgl(outm structs.OutMessage, vs map[string]structs.VendorService, clientID int) *structs.ErrorMessage {
	var errors structs.ErrorMessage
	var uri string
	var resp *http.Response
	var bodyStr string
	var body []byte
	var err error

	switch outm.Channel {
	case "Whatsapp":
		switch vs["Whatsapp"].Alias {
		case "SLEEKFLOW":
			fromSender := vs["Whatsapp"].FromSender
			bodyStr = sleekflowReplyConversation(outm, &fromSender)
			uri = strings.Replace(vs["Whatsapp"].URI, "{{api_key}}", vs["Whatsapp"].Token, -1)
			_, resp, body, _, err = common.HitAPI(vs["Whatsapp"].URL+uri, bodyStr, vs["Whatsapp"].Method, vs["Whatsapp"].HeaderPrefix+" "+vs["Whatsapp"].Token, vs["Whatsapp"].Alias, time.Duration(0), oub.DB)

		default:
			bodyStr, uri = sendWAConv(outm, vs)
			//fmt.Println(string(body))
			_, resp, body, _, err = common.HitAPI(vs["Whatsapp"].URL+uri, bodyStr, vs["Whatsapp"].Method, vs["Whatsapp"].HeaderPrefix+" "+vs["Whatsapp"].Token, vs["Whatsapp"].Alias, time.Duration(0), oub.DB)

		}

	case "BlastWhatsapp":
		switch vs["BlastWhatsapp"].Alias {
		case "SLEEKFLOW":
			fromSender := vs["BlastWhatsapp"].FromSender
			uri = strings.Replace(vs["BlastWhatsapp"].URI, "{{api_key}}", vs["BlastWhatsapp"].Token, -1)

			switch outm.Type {
			case "interactive":
				bodyStr = sleekflowSendInteractiveWA(outm, &fromSender)
			default:
				bodyStr = sleekflowBlastTemplatePlan(outm, &fromSender)
			}

			_, resp, body, _, err = common.HitAPI(vs["BlastWhatsapp"].URL+uri, bodyStr, vs["BlastWhatsapp"].Method, vs["BlastWhatsapp"].HeaderPrefix+" "+vs["BlastWhatsapp"].Token, vs["BlastWhatsapp"].Alias, time.Duration(0), oub.DB)

		default:

			bodyStr = sendWABlast(outm, vs)
			_, resp, body, _, err = common.HitAPI(vs["BlastWhatsapp"].URL+vs["BlastWhatsapp"].URI, bodyStr, vs["BlastWhatsapp"].Method, vs["BlastWhatsapp"].HeaderPrefix+" "+vs["BlastWhatsapp"].Token, vs["BlastWhatsapp"].Alias, time.Duration(0), oub.DB)

		}

	case "BlastWhatsappHeader":
		switch vs["BlastWhatsappHeader"].Alias {
		case "SLEEKFLOW":
			fromSender := vs["BlastWhatsappHeader"].FromSender
			uri = strings.Replace(vs["BlastWhatsappHeader"].URI, "{{api_key}}", vs["BlastWhatsappHeader"].Token, -1)

			bodyStr = sleekflowBlastTemplatePlan(outm, &fromSender)

			_, resp, body, _, err = common.HitAPI(vs["BlastWhatsappHeader"].URL+uri, bodyStr, vs["BlastWhatsappHeader"].Method, vs["BlastWhatsappHeader"].HeaderPrefix+" "+vs["BlastWhatsappHeader"].Token, vs["BlastWhatsappHeader"].Alias, time.Duration(0), oub.DB)

		default:
			
			bodyStr = sendWABlastWithHeader(outm, vs)
			// fmt.Println(bodyStr)
			_, resp, body, _, err = common.HitAPI(vs["BlastWhatsappHeader"].URL+vs["BlastWhatsappHeader"].URI, bodyStr, vs["BlastWhatsappHeader"].Method, vs["BlastWhatsappHeader"].HeaderPrefix+" "+vs["BlastWhatsappHeader"].Token, vs["BlastWhatsappHeader"].Alias, time.Duration(0), oub.DB)

		}

	case "Facebook":
		bodyStr, uri = sendFBConv(outm, vs)
		_, resp, body, _, err = common.HitAPI(vs["Facebook"].URL+uri, bodyStr, vs["Facebook"].Method, vs["Facebook"].HeaderPrefix+" "+vs["Facebook"].Token, vs["Facebook"].Alias, time.Duration(0), oub.DB)
	case "BlastFacebook":
		bodyStr = sendFBBlast(outm, vs)
		_, resp, body, _, err = common.HitAPI(vs["BlastFacebook"].URL+vs["BlastFacebook"].URI, bodyStr, vs["BlastFacebook"].Method, vs["BlastFacebook"].HeaderPrefix+" "+vs["BlastFacebook"].Token, vs["BlastFacebook"].Alias, time.Duration(0), oub.DB)
	case "Telegram":
		bodyStr, uri = sendTeleConv(outm, vs)
		_, resp, body, _, err = common.HitAPI(vs["Telegram"].URL+uri, bodyStr, vs["Telegram"].Method, vs["Telegram"].HeaderPrefix+" "+vs["Telegram"].Token, vs["Telegram"].Alias, time.Duration(0), oub.DB)
	case "BlastTelegram":
		bodyStr = sendTeleBlast(outm, vs)
		_, resp, body, _, err = common.HitAPI(vs["BlastTelegram"].URL+vs["BlastTelegram"].URI, bodyStr, vs["BlastTelegram"].Method, vs["BlastTelegram"].HeaderPrefix+" "+vs["BlastTelegram"].Token, vs["BlastTelegram"].Alias, time.Duration(0), oub.DB)
	case "Email":
		//Call func to hit email
		bodyStr = sendEmail(outm, vs)
		switch vs["Email"].Alias {
		case "MAIL_JET":
			strAuth := vs["Email"].LoginAuthType + " " + common.BasicAuth(vs["Email"].Username, vs["Email"].Password)
			_, resp, body, _, err = common.HitAPI(vs["Email"].URL+vs["Email"].URI, bodyStr, vs["Email"].Method, strAuth, vs["Email"].Alias, time.Duration(0), oub.DB)
		case "SPARKPOST":
			strAuth := vs["Email"].LoginAuthType + " " + vs["Email"].Token
			_, resp, body, _, err = common.HitAPI(vs["Email"].URL+vs["Email"].URI, bodyStr, vs["Email"].Method, strAuth, vs["Email"].Alias, time.Duration(0), oub.DB)
		default:
			_, resp, body, _, err = common.HitAPI(vs["Email"].URL+vs["Email"].URI, bodyStr, vs["Email"].Method, vs["Email"].HeaderPrefix, vs["Email"].Alias, time.Duration(0), oub.DB)
		}

	case "Instagram":
		//Call func to hit Instagram
	case "WhatsappTest":
		//call func to hit whatsapptest
	case "BlastWhatsappTest":
		//call func to hit whatsapptest
	case "Dummy":
		//bodyStr = sendTeleBlast(outm, vs)

		bodyStrbyte, _ := json.Marshal(outm)
		bodyStr = string(bodyStrbyte)

		//		_, resp, body, _, err = common.HitAPI(vs["Dummy"].URL+vs["Dummy"].URI, bodyStr, vs["Dummy"].Method, vs["Dummy"].HeaderPrefix+" "+vs["Dummy"].Token, vs["Dummy"].Alias, time.Duration(0))
		if outm.CallBackAuth.CallBackUrl != "" {
			_, resp, body, _, err = common.HitAPI(outm.CallBackAuth.CallBackUrl, bodyStr, "POST", "", "", time.Duration(0), oub.DB)
		}
	case "DAMCORP":
		// var damCorpResp structs.Damcorp_Response
		// damCorpResp := make(map[string]interface{})
		bodyStr = sendDamCorpWa(outm)

		strAuth := "Bearer " + vs["DAMCORP"].Token

		_, resp, body, _, err = common.HitAPI(vs["DAMCORP"].URI, bodyStr, "POST", strAuth, "", time.Duration(0), oub.DB)

		// json.Unmarshal(body, &damCorpResp)

		// errors.MessageID = damCorpResp.Messages[0].Id
	case "DamcorpWhatsappBlast":
		bodyStr = sendDamcorpWaBlast(outm)

		secret := []byte(vs["DamcorpWhatsappBlast"].Username+vs["DamcorpWhatsappBlast"].Password)


		dtoToken := mysql.ViperEnvVariable("DTO_TOKEN")

		tokenDecrypt, _ := common.DecryptAES(dtoToken, secret)

		strAuth := "Basic " + tokenDecrypt

		_, resp, body, _, err = common.HitAPI(vs["DamcorpWhatsappBlast"].URL+vs["DamcorpWhatsappBlast"].URI, bodyStr, "POST", strAuth, "", time.Duration(0), oub.DB)

	}
	reqSesame, _ := json.Marshal(outm)
	if err != nil {
		errors.ReqSesame = string(reqSesame)
		if resp != nil {
			errors.Message = resp.Status
		}

		errors.Data = err.Error()
		errors.ReqMessage = bodyStr
		errors.RespMessage = string(body)
		if resp != nil {
			errors.Code = resp.StatusCode
		}
		return &errors
	}
	json.Unmarshal(body, &errors)
	outm.ChannelID = errors.ChannelID
	outm.MessageID = errors.MessageID
	_ = oub.outboundRepo.CreateOutbound(outm, bodyStr, string(body), clientID)

	if (outm.CallBackAuth != structs.CallBackAuth{}) {
		var rspcallback structs.RspOutbound
		rspcallback.Outbound = outm
		//rspcallback.VendorResponse = string(body)

		jsonBody := make(map[string]interface{})
		json.Unmarshal(body, &jsonBody)

		rspcallback.VendorResponse = jsonBody
		rspcallbackbyte, _ := json.Marshal(rspcallback)
		_, resp, _, _, err = common.HitAPICallBack(outm.CallBackAuth, string(rspcallbackbyte), time.Duration(0), oub.DB)
		if err != nil {
			fmt.Println(err)
		}

	}

	//get response from the API and save into structs.ErrorMessage

	//json.Unmarshal(reqSesame, &errors)
	errors.ReqSesame = string(reqSesame)
	if resp == nil {
		errors.Message = "200"
		errors.Code = 200
	} else {
		errors.Message = resp.Status
		errors.Code = resp.StatusCode
	}

	errors.RespMessage = string(body)
	errors.ReqMessage = bodyStr

	return &errors
}

func (oub *outboundService) SendOutboundBulk(outm structs.OutMessage, vs map[string]structs.VendorService, clientID int) structs.HTTPRequest {
	var errors structs.ErrorMessage
	var uri string
	var resp *http.Response
	var bodyStr string
	var err error
	var hitResp structs.HTTPRequest

	switch outm.Channel {
	case "Whatsapp":
		bodyStr, uri = sendWAConv(outm, vs)
		//fmt.Println(string(body))
		hitResp, err = common.HitAPIBulk(vs["Whatsapp"].URL+uri, bodyStr, vs["Whatsapp"].Method, vs["Whatsapp"].HeaderPrefix+" "+vs["Whatsapp"].Token, vs["Whatsapp"].Alias, time.Duration(0))
	case "BlastWhatsapp":
		bodyStr = sendWABlast(outm, vs)
		hitResp, err = common.HitAPIBulk(vs["BlastWhatsapp"].URL+vs["BlastWhatsapp"].URI, bodyStr, vs["BlastWhatsapp"].Method, vs["BlastWhatsapp"].HeaderPrefix+" "+vs["BlastWhatsapp"].Token, vs["BlastWhatsapp"].Alias, time.Duration(0))
	case "BlastWhatsappHeader":
		bodyStr = sendWABlastWithHeader(outm, vs)
		// fmt.Println(bodyStr)
		hitResp, err = common.HitAPIBulk(vs["BlastWhatsappHeader"].URL+vs["BlastWhatsappHeader"].URI, bodyStr, vs["BlastWhatsappHeader"].Method, vs["BlastWhatsappHeader"].HeaderPrefix+" "+vs["BlastWhatsappHeader"].Token, vs["BlastWhatsappHeader"].Alias, time.Duration(0))

	case "Facebook":
		bodyStr, uri = sendFBConv(outm, vs)
		hitResp, err = common.HitAPIBulk(vs["Facebook"].URL+uri, bodyStr, vs["Facebook"].Method, vs["Facebook"].HeaderPrefix+" "+vs["Facebook"].Token, vs["Facebook"].Alias, time.Duration(0))
	case "BlastFacebook":
		bodyStr = sendFBBlast(outm, vs)
		hitResp, err = common.HitAPIBulk(vs["BlastFacebook"].URL+vs["BlastFacebook"].URI, bodyStr, vs["BlastFacebook"].Method, vs["BlastFacebook"].HeaderPrefix+" "+vs["BlastFacebook"].Token, vs["BlastFacebook"].Alias, time.Duration(0))
	case "Telegram":
		bodyStr, uri = sendTeleConv(outm, vs)
		hitResp, err = common.HitAPIBulk(vs["Telegram"].URL+uri, bodyStr, vs["Telegram"].Method, vs["Telegram"].HeaderPrefix+" "+vs["Telegram"].Token, vs["Telegram"].Alias, time.Duration(0))
	case "BlastTelegram":
		bodyStr = sendTeleBlast(outm, vs)
		hitResp, err = common.HitAPIBulk(vs["BlastTelegram"].URL+vs["BlastTelegram"].URI, bodyStr, vs["BlastTelegram"].Method, vs["BlastTelegram"].HeaderPrefix+" "+vs["BlastTelegram"].Token, vs["BlastTelegram"].Alias, time.Duration(0))
	case "Email":
		//Call func to hit email
		bodyStr = sendEmail(outm, vs)
		switch vs["Email"].Alias {
		case "MAIL_JET":
			strAuth := vs["Email"].LoginAuthType + " " + common.BasicAuth(vs["Email"].Username, vs["Email"].Password)
			hitResp, err = common.HitAPIBulk(vs["Email"].URL+vs["Email"].URI, bodyStr, vs["Email"].Method, strAuth, vs["Email"].Alias, time.Duration(0))
		case "SPARKPOST":
			strAuth := vs["Email"].LoginAuthType + " " + vs["Email"].Token
			hitResp, err = common.HitAPIBulk(vs["Email"].URL+vs["Email"].URI, bodyStr, vs["Email"].Method, strAuth, vs["Email"].Alias, time.Duration(0))
		default:
			hitResp, err = common.HitAPIBulk(vs["Email"].URL+vs["Email"].URI, bodyStr, vs["Email"].Method, vs["Email"].HeaderPrefix, vs["Email"].Alias, time.Duration(0))
		}

	case "Instagram":
		//Call func to hit Instagram
	case "WhatsappTest":
		//call func to hit whatsapptest
	case "BlastWhatsappTest":
		//call func to hit whatsapptest
	case "Dummy":
		//bodyStr = sendTeleBlast(outm, vs)

		bodyStrbyte, _ := json.Marshal(outm)
		bodyStr = string(bodyStrbyte)

		if outm.CallBackAuth.CallBackUrl != "" {
			hitResp, err = common.HitAPIBulk(outm.CallBackAuth.CallBackUrl, bodyStr, "POST", "", "", time.Duration(0))
		}
	case "DAMCORP":
		// var damCorpResp structs.Damcorp_Response
		// damCorpResp := make(map[string]interface{})
		bodyStr = sendDamCorpWa(outm)

		strAuth := "Bearer " + vs["DAMCORP"].Token

		hitResp, err = common.HitAPIBulk(vs["DAMCORP"].URI, bodyStr, "POST", strAuth, "", time.Duration(0))

		// json.Unmarshal(body, &damCorpResp)

		// errors.MessageID = damCorpResp.Messages[0].Id
	case "DamcorpWhatsappBlast":
		bodyStr = sendDamcorpWaBlast(outm)

		secret := []byte(vs["DamcorpWhatsappBlast"].Username+vs["DamcorpWhatsappBlast"].Password)


		dtoToken := mysql.ViperEnvVariable("DTO_TOKEN")

		tokenDecrypt, _ := common.DecryptAES(dtoToken, secret)

		strAuth := "Basic " + tokenDecrypt

		hitResp, err = common.HitAPIBulk(vs["DamcorpWhatsappBlast"].URL+vs["DamcorpWhatsappBlast"].URI, bodyStr, "POST", strAuth, "", time.Duration(0))

	}

	if err != nil {
		if resp != nil {
			errors.Message = resp.Status
		}

		errors.Data = err.Error()
		errors.ReqMessage = bodyStr
		errors.RespMessage = string(hitResp.ResponseBody)
		if resp != nil {
			errors.Code = resp.StatusCode
		}
		// return hitResp
	}

	json.Unmarshal(hitResp.ResponseBody, &errors)
	outm.ChannelID = errors.ChannelID
	outm.MessageID = errors.MessageID

	outboundByte, _ := json.Marshal(outm)
	// reqBodyByte, _ := json.Marshal(bodyStr)

	hitResp.Outbound = outboundByte
	hitResp.RequestBody = bodyStr

	return hitResp

}

func (oub *outboundService) Validate(outm *structs.OutMessage) (*structs.OutMessage, *structs.ErrorMessage) {
	var errors structs.ErrorMessage

	v := validator.New()
	err := v.Struct(outm)
	if err != nil {
		errors.Message = structs.Validate
		errors.Data = ""
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return nil, &errors
	}

	return outm, nil
}

func (oub *outboundService) ValidateFlow(outmflow *structs.OutMessageFlow) (*structs.OutMessageFlow, *structs.ErrorMessage) {
	var errors structs.ErrorMessage
	//if strings.ToUpper(outmflow.Channel) == "FLOW" {

	if outmflow.Order == 0 {
		errors.Message = "Invalid order"
		errors.Data = ""
		//errors.SysMessage = err.Error()
		//errors.Code = http.StatusInternalServerError
		return outmflow, &errors
	}
	if outmflow.CustName == "" {
		errors.Message = "Cust Name Can Not Be Empty"
		errors.Data = ""
		return outmflow, &errors
	}
	//}

	return outmflow, nil
}

func (oub *outboundService) SendOutboundFlowSgl(outbflow structs.OutboundFlow, vs map[string]structs.VendorService, clientID int) *structs.ErrorMessage {
	var errors structs.ErrorMessage
	//var uri string
	var resp *http.Response
	var bodyStr string
	var body []byte
	var err error

	bodyStr, _ = saveFlowLog(clientID, outbflow)

	strAuth := vs["Flow"].HeaderPrefix + " " + common.BasicAuth(vs["Flow"].Username, vs["Flow"].Password)
	_, resp, body, _, err = common.HitAPI(vs["Flow"].URL+vs["Flow"].URI, bodyStr, vs["Flow"].Method, strAuth, vs["Flow"].Alias, time.Duration(0), oub.DB)
	reqSesame, _ := json.Marshal(outbflow)
	if err != nil {
		errors.ReqSesame = string(reqSesame)
		if resp != nil {
			errors.Message = resp.Status
		}

		errors.Data = err.Error()
		errors.ReqMessage = bodyStr
		errors.RespMessage = string(body)
		if resp != nil {
			errors.Code = resp.StatusCode
		}
		return &errors
	}
	errors.ReqSesame = string(reqSesame)
	errors.Message = resp.Status
	errors.RespMessage = string(body)
	errors.ReqMessage = bodyStr
	errors.Code = resp.StatusCode
	return &errors
}

func saveFlowLog(clientID int, outbflow structs.OutboundFlow) (string, string) {

	//outbconsumerbyte, _ := json.Marshal(outbconsumer)

	var fl structs.Flow

	fl.ClientID = clientID

	fl.OutboundFlow = outbflow
	rtn := flowRepo.CreateFlow(fl)
	flowid, _ := strconv.Atoi(rtn.Data)
	fl.ID = flowid
	bodyStr, _ := json.Marshal(fl)
	return string(bodyStr), string(rtn.Data)
}

func (oub *outboundService) CloseSessionWA(cs structs.Damcorp_Close_Session, vs map[string]structs.VendorService) *structs.ErrorMessage {
	var (
		err     error
		errors  structs.ErrorMessage
		resp    *http.Response
		bodyStr string
		body    []byte
		csrsp   structs.Damcorp_Close_Session_Response
	)

	jsonByte, _ := json.Marshal(cs)
	bodyStr = string(jsonByte)
	strAuth := "Basic " + vs["CloseSession"].Token

	_, resp, body, _, err = common.HitAPI(vs["CloseSession"].URL+vs["CloseSession"].URI, bodyStr, "POST", strAuth, "", time.Duration(0), oub.DB)

	if err != nil {
		if resp != nil {
			errors.Message = resp.Status
		}

		errors.Data = ""
		errors.ReqMessage = bodyStr
		errors.RespMessage = string(body)
		if resp != nil {
			errors.Code = resp.StatusCode
		}
		return &errors
	}

	json.Unmarshal(body, &csrsp)

	errors.RespMessage = string(body)
	errors.Message = resp.Status
	errors.Code = resp.StatusCode

	return &errors
}

func (oub *outboundService) CreateOutboundBulkLog(outbound []byte, cli int) (*structs.OutboundBulkLog, error) {
	var outblk structs.OutboundBulkLog
	outblk.RequestBody = string(outbound)
	outblk.ClientId = cli
	outblk.Status = 1
	return oub.outboundRepo.CreateOutboundBulkLog(outblk)
}

func (oub *outboundService) InsertOutboundBulk(outm []*structs.HTTPRequest, cli int) {
	response := oub.outboundRepo.CreateOutboundBulk(outm, cli)

	if response.Code != 200 {
		log.Println("error insert outbound bulk : ", response.Message)
		log.Println("error code : ", response.Code)
	}else{
		log.Println("success insert outbound bulk : ", response.Message)
	}

}

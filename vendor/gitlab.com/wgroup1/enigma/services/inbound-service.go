package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gitlab.com/wgroup1/enigma/common"
	"gitlab.com/wgroup1/enigma/interfaces"
	"gitlab.com/wgroup1/enigma/structs"
	"gopkg.in/go-playground/validator.v9"
)

type InboundService interface {
	//CreateInbound(inm structs.Inbound, iiw structs.InternalInboundWa, respStr string, cli int) *structs.ErrorMessage
	Validate(inm *structs.Inbound) (*structs.Inbound, *structs.ErrorMessage)
	//GetDynamicStructByChannel(inboundChannel string, fromsender string, tosender string) ([]structs.VendorService, *structs.ErrorMessage)
	GetWebhook(inboundChannel string, fromsender string, tosender string, trx_type string) ([]structs.WebHook, *structs.ErrorMessage)
	SendInboundSgl(inm structs.Inbound, vs []structs.WebHook, clientID int) *structs.ErrorMessage
	GetTokenStatus(vendor string, channel string) []structs.VendorService
}

type inboundService struct{
	inboundRepo interfaces.InboundRepository
	outboundRepo interfaces.OutboundRepository
	DB *sql.DB
}


func NewInboundService(repository interfaces.InboundRepository, outrepo interfaces.OutboundRepository, db *sql.DB) InboundService {
	return &inboundService{
		inboundRepo: repository,
		outboundRepo: outrepo,
		DB: db,
	}
}

func (service *inboundService) GetTokenStatus(vendor string, channel string) []structs.VendorService {
	vss, _ := service.outboundRepo.GetTokenStatus(vendor, channel)
	return vss
}

func (service *inboundService) GetWebhook(inboundChannel string, fromsender string, tosender string, trx_type string) ([]structs.WebHook, *structs.ErrorMessage) {
	return service.inboundRepo.GetWebhook(inboundChannel, fromsender, tosender, trx_type)
}

func (service *inboundService) Validate(inm *structs.Inbound) (*structs.Inbound, *structs.ErrorMessage) {
	var errors structs.ErrorMessage
	v := validator.New()
	err := v.Struct(inm)
	if err != nil {
		errors.Message = structs.Validate
		errors.Data = ""
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return nil, &errors
	}
	return inm, nil
}

func (service *inboundService) SendInboundSgl(inm structs.Inbound, vss []structs.WebHook, clientID int) *structs.ErrorMessage {
	var errors structs.ErrorMessage
	var uri string
	var body []byte
	var err error
	var iiw structs.InternalInboundWa
	// var jsonIiwMsg []byte
	//var iiwsesame structs.InternalInboundSesame
	var iiwsesame []byte
	//var jsonIiwSesameMsg []byte

	vs := structs.WebHook{}
	ctx := context.Background()

	for iLoop := range vss {
		sCase := strings.ToUpper(inm.Message.Platform)
		//vs.ClientID = vss[iLoop].ClientID
		chkws := false
		switch sCase {
		case "WHATSAPP":
			iiw, iiwsesame, uri = AssignInboundWebhook(inm, vs)
			//_, _, body, _, err = common.HitAPI(vss[iLoop].URL+uri, string(iiwsesame), vss[iLoop].Method, vss[iLoop].Token, "", time.Duration(0))
		case "TELEGRAM":
			iiw, iiwsesame, uri = AssignInboundWebhook(inm, vs)
			//_, _, body, _, err = common.HitAPI(vss[iLoop].URL+uri, string(iiwsesame), vss[iLoop].Method, vss[iLoop].Token, "", time.Duration(0))
		case "FACEBOOK":
			iiw, iiwsesame, uri = AssignInboundWebhook(inm, vs)
			//jsonIiwSesameMsg, _ = json.Marshal(iiwsesame)
			//_, _, body, _, err = common.HitAPI(vss[iLoop].URL+uri, string(iiwsesame), vss[iLoop].Method, vss[iLoop].Token, "", time.Duration(0))
		case "EMAIL":
			//Call func to hit email
			iiw, iiwsesame, uri = AssignInboundWebhook(inm, vs)
			//_, _, body, _, err = common.HitAPI(vss[iLoop].URL+uri, string(iiwsesame), vss[iLoop].Method, vss[iLoop].Token, "", time.Duration(0))
		case "Instagram":
			//Call func to hit Instagram
		case "Facebook":
			//Call func to hit Facebook
		case "WhatsappTest":
			//call func to hit whatsapptest
		case "BlastWhatsappTest":
			//call func to hit whatsapptest

		}

		startAll := time.Now()
		if iiwsesame != nil {
			_, _, body, _, err = common.HitAPI(vss[iLoop].URL+uri, string(iiwsesame), vss[iLoop].Method, vss[iLoop].Token, "", time.Duration(0), service.DB)
			if err != nil {
				for wh := 0; wh < vss[iLoop].Retry; wh++ {
					if !chkws {
						_, _, body, _, err = common.HitAPI(vss[iLoop].URL+uri, string(iiwsesame), vss[iLoop].Method, vss[iLoop].Token, "", time.Duration(0), service.DB)
						if err == nil {
							chkws = true
							wh = vss[iLoop].Retry
						} else {
							fmt.Println("retry at:", time.Now())
						}
					}
				}
			}
			endTime := time.Since(startAll)
			fmt.Println("hit webhook time : ", endTime)
			if err != nil {
				errors.Message = structs.ErrNotFound
				errors.SysMessage = err.Error()
				errors.Code = http.StatusInternalServerError
				fmt.Println(err.Error())
				return &errors
			}
			//get response from the API and save into structs.ErrorMessage
			// jsonIiwMsg, _ = json.Marshal(iiw)
		}

	}
	_ = service.inboundRepo.CreateInbound(inm, iiw, string(body), clientID, ctx)

	errors.Message = structs.Success
	errors.Data = string(body)
	// errors.SysMessage = string(jsonIiwMsg)
	errors.Code = http.StatusOK
	return &errors
}

func AssignInboundWebhook(inm structs.Inbound, vs structs.WebHook) (iiw structs.InternalInboundWa, iiwsesame []byte, sUri string) {

	var uri string

	//NOTE : Untuk disimpan di log Bridge
	iiw.Contactid = inm.Contact.ID
	iiw.ContactJsonPlain = inm.Contact
	iiw.ConversationJsonPlain = inm.Conversation
	iiw.Message = inm.Message
	iiw.ConversationId = inm.Message.Conversationid
	iiw.Msisdn = inm.Contact.Msisdn
	iiw.MessageId = inm.Message.ID
	iiw.ChannelId = inm.Message.Channelid

	iiw.JsonPlainAll = inm
	iiw.ChannelId = inm.Message.Platform

	//NOTE : Untuk Hit ke API EXTERNAL

	ssm := structs.InternalInboundSesameAll{}
	ssm.Channel = inm.Message.Platform
	ssm.ContactID = inm.Contact.ID
	ssm.ContactName = inm.Contact.Displayname
	ssm.ConversationID = inm.Message.Conversationid
	ssm.MessageID = inm.Message.ID
	ssm.To = inm.Message.To
	ssm.From = inm.Message.From
	ssm.Cc = inm.Message.Cc
	ssm.Bcc = inm.Message.Bcc
	ssm.Subject = inm.Message.Subject
	ssm.Type = inm.Message.Type

	ssm.Status = inm.Message.Status
	ssm.Createddatetime = inm.Message.Createddatetime
	ssm.Updateddatetime = inm.Message.Updateddatetime

	sCaseType := strings.ToLower(inm.Message.Type)
	switch sCaseType {
	case "text":
		ctntxtssm := structs.ContentTextSesame{}
		ctntxtssm.Text = inm.Message.Content.Text
		ctn, _ := json.Marshal(ctntxtssm)
		json.Unmarshal(ctn, &ssm.Content)
		/*
			ctnimgssm := structs.ContentImageSesame{}
			ctnimgssm.Text = inm.Message.Content.Text
			ctnimgssmchild := structs.ContentImageSesameChild{}
			ctnimgssmchild.Url = inm.Message.Content.Image.URL
			ctnimgssmchild.Caption = inm.Message.Content.Image.Caption
			ctnimgssm.Image = ctnimgssmchild
			ctn, _ := json.Marshal(ctnimgssm)
			json.Unmarshal(ctn, &ssm.Content)
		*/
		var err1 error
		iiwsesame, err1 = json.Marshal(ssm)
		_ = err1
	case "image":
		ctnssm := structs.ContentImageSesame{}
		ctnssm.Text = inm.Message.Content.Text
		ctnssmchild := structs.ContentImageSesameChild{}
		ctnssmchild.Url = inm.Message.Content.Image.URL
		ctnssmchild.Caption = inm.Message.Content.Image.Caption
		ctnssm.Image = ctnssmchild
		ctn, _ := json.Marshal(ctnssm)
		json.Unmarshal(ctn, &ssm.Content)

		var err1 error
		iiwsesame, err1 = json.Marshal(ssm)
		_ = err1

	case "video":
		ctnssm := structs.ContentVideoSesame{}
		ctnssm.Text = inm.Message.Content.Text
		ctnssmchild := structs.ContentVideoSesameChild{}
		ctnssmchild.Url = inm.Message.Content.Video.URL
		//ctnssmchild.Caption = inm.Message.Content.Image.Caption
		ctnssm.Video = ctnssmchild
		ctn, _ := json.Marshal(ctnssm)
		json.Unmarshal(ctn, &ssm.Content)

		var err1 error
		iiwsesame, err1 = json.Marshal(ssm)
		_ = err1
	}
	uri = vs.URL

	return iiw, iiwsesame, uri
}

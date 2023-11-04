package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"gitlab.com/wgroup1/enigma/common"
	service "gitlab.com/wgroup1/enigma/services"
	"gitlab.com/wgroup1/enigma/structs"
)

type inboundLogic struct{
	inboundService service.InboundService
	DB *sql.DB
}

type InboundLogic interface {
	CreateInbound(w http.ResponseWriter, r *http.Request)
	ReceiveMessageID(w http.ResponseWriter, r *http.Request)
	ArchivedConversation(w http.ResponseWriter, r *http.Request)
}

func NewInboundLogic(service service.InboundService, db *sql.DB) InboundLogic {
	return &inboundLogic{
		inboundService: service,
		DB: db,
	}
}

//was built for MessageBird flowbuilder
func (ib *inboundLogic) ReceiveMessageID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var inb structs.Inbound
	var msg structs.MessageInbound
	var errs []structs.ErrorMessage
	var errstr structs.ErrorMessage

	_ = json.NewDecoder(r.Body).Decode(&inb)

	vss := ib.inboundService.GetTokenStatus("MSGBRD", "WhatsappStatus")

	t := make(map[string]structs.VendorService)
	for k := range vss {
		if inb.ClientID == vss[k].ClientID {
			t[vss[k].Channel] = vss[k]
		}
	}

	uri := strings.Replace(t["WhatsappStatus"].URI, "{{msg_id}}", inb.MessageID, -1)

	_, resp, body, _, err := common.HitAPI(t["WhatsappStatus"].URL+uri, "", t["WhatsappStatus"].Method, t["WhatsappStatus"].HeaderPrefix+" "+t["WhatsappStatus"].Token, t["WhatsappStatus"].Alias, time.Duration(0), ib.DB)

	if err != nil {
		fmt.Println(err.Error())
		errstr.Code = 500
		errstr.Message = "Internal Server Error"
		errstr.SysMessage = string(body)
		common.JSONErr(w, &errstr)
		return
	}

	if resp.StatusCode != 200 {
		errstr.Code = resp.StatusCode
		errstr.Message = resp.Status
		errstr.SysMessage = string(body)
		common.JSONErr(w, &errstr)
		return
	}

	json.Unmarshal(body, &msg)

	inb.Message = msg

	vs, _ := ib.inboundService.GetWebhook(inb.Message.Platform, "", inb.Message.To, "Inbound")

	if vs != nil {
		errStr := ib.inboundService.SendInboundSgl(inb, vs, 1)
		errs = append(errs, *errStr)
		common.JSONErrs(w, &errs)
	} else {

		var errors structs.ErrorMessage
		errors.Message = structs.ErrNotFound
		errors.Code = http.StatusNoContent
		errs = append(errs, errors)
		common.JSONErrs(w, &errs)

	}

}

//was built for MessageBird flowbuilder
func (ib *inboundLogic) ArchivedConversation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var arch structs.Archived
	var errstr structs.ErrorMessage

	_ = json.NewDecoder(r.Body).Decode(&arch)

	vss := ib.inboundService.GetTokenStatus("MSGBRD", "Archived")

	t := make(map[string]structs.VendorService)
	for k := range vss {
		if arch.ClientID == vss[k].ClientID {
			t[vss[k].Channel] = vss[k]
		}
	}

	stat := structs.ArchStatus{}
	stat.Status = "archived"
	req, _ := json.Marshal(stat)
	uri := strings.Replace(t["Archived"].URI, "{{conv_id}}", arch.ConversationID, -1)

	_, resp, body, _, err := common.HitAPI(t["Archived"].URL+uri, string(req), t["Archived"].Method, t["Archived"].HeaderPrefix+" "+t["Archived"].Token, t["Archived"].Alias, time.Duration(0), ib.DB)

	if err != nil {
		fmt.Println(err.Error())
		errstr.Code = 500
		errstr.Message = "Internal Server Error"
		errstr.SysMessage = string(body)
		common.JSONErr(w, &errstr)
		return
	}

	if resp.StatusCode != 200 {
		errstr.Code = 500
		errstr.Message = "Internal Server Error"
		errstr.SysMessage = string(body)
		common.JSONErr(w, &errstr)
		return
	}

	errstr.Code = 200
	errstr.Message = "Conversation ID:" + arch.ConversationID + " has been archived"
	errstr.SysMessage = string(body)
	common.JSONErr(w, &errstr)
}

func (ib *inboundLogic) CreateInbound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var inb structs.Inbound
	var errs []structs.ErrorMessage

	//var wg sync.WaitGroup

	_ = json.NewDecoder(r.Body).Decode(&inb)

	if inb.Message.Direction == "received" {
		//fmt.Println("request inbound:", inb)
		vs, _ := ib.inboundService.GetWebhook(inb.Message.Platform, "", inb.Message.To, "Inbound")
		//fmt.Println("vs:", vs)
		if vs != nil {
			////convert struct.authvendor to maps
			//t := make(map[string]structs.VendorService)
			////t := make(map[string]structs.Inbound)
			//for k := range vs {
			//t[vs[k].Alias] = vs[k]
			//}
	
			errStr := ib.inboundService.SendInboundSgl(inb, vs, 1)
			errs = append(errs, *errStr)
			log.Printf("inbound from: %s", inb.Message.From)
			common.JSONErrs(w, &errs)
		} else {
	
			var errors structs.ErrorMessage
			errors.Message = structs.ErrNotFound
			errors.Code = http.StatusNoContent
			errs = append(errs, errors)
			common.JSONErrs(w, &errs)
	
		}

	} else {
		var errors structs.ErrorMessage
		errors.Message = structs.Success
		errors.Code = http.StatusOK
		errs = append(errs, errors)
		common.JSONErrs(w, &errs)
	}

}

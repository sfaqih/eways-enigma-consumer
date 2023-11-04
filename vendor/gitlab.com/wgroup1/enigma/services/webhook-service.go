package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gitlab.com/wgroup1/enigma/interfaces"
	"gitlab.com/wgroup1/enigma/structs"
	"gopkg.in/go-playground/validator.v9"
)

type WebhookService interface {
	GetURL(clientID int) []structs.WebHook

	GetWebHooks(clientId int, clientCode string, page string, limit string) []structs.WebHookObj
	GetWebHook(clientId int, clientCode string, webhookId int) structs.WebHookObj

	ValidateSave(wh structs.WebHookObj) (structs.WebHookObj, *structs.ErrorMessage)
	SaveWebHook(wh structs.WebHookObj) *structs.ErrorMessage
}

type webhookService struct{}

var (
	webhookRepo interfaces.WebHookRepository
)

func NewWebHookService(repository interfaces.WebHookRepository) WebhookService {
	webhookRepo = repository
	return &webhookService{}
}

func (*webhookService) GetURL(clientID int) []structs.WebHook {
	return webhookRepo.GetURL(clientID)
}

func (*webhookService) Validate(web *structs.WebHook) (*structs.WebHook, *structs.ErrorMessage) {
	var errors structs.ErrorMessage
	v := validator.New()
	err := v.Struct(web)
	if err != nil {
		errors.Message = structs.Validate
		errors.Data = ""
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return nil, &errors
	}
	return web, nil
}

func (*webhookService) GetWebHooks(clientId int, clientCode string, page string, limit string) []structs.WebHookObj {

	var webHooksObj []structs.WebHookObj
	var webHookObj structs.WebHookObj
	//var eventsObj []structs.Event
	//clientID := 0 //clientRepo.GetClientID(clientCode)

	webHooksPlainRtn := webhookRepo.GetWebHooks(clientId, page, limit)
	for j := range webHooksPlainRtn {
		webHookObj.ID = webHooksPlainRtn[j].ID
		webHookObj.ClientID = webHooksPlainRtn[j].ClientID
		webHookObj.ClientCode = webHooksPlainRtn[j].ClientCode
		webHookObj.URL = webHooksPlainRtn[j].URL
		webHookObj.Method = webHooksPlainRtn[j].Method
		webHookObj.HeaderPrefix = webHooksPlainRtn[j].HeaderPrefix
		webHookObj.Token = webHooksPlainRtn[j].Token

		err1 := json.Unmarshal([]byte(webHooksPlainRtn[j].Events), &webHookObj.Events)
		if err1 != nil {
			fmt.Println(err1.Error())
		}
		//eventsJson, _ := json.Marshal([]byte(webHooksPlainRtn[j].Events))
		//webHookObj.Events = eventsJson

		//webHookObj.Events = webHooksPlainRtn[j].Events

		webHookObj.HttpCode = webHooksPlainRtn[j].HttpCode
		webHookObj.Retry = webHooksPlainRtn[j].Retry
		webHookObj.Timeout = webHooksPlainRtn[j].Timeout
		webHookObj.Status = webHooksPlainRtn[j].Status
		webHookObj.CreatedAt = webHooksPlainRtn[j].CreatedAt
		webHookObj.CreatedBy = webHooksPlainRtn[j].CreatedBy

		webHooksObj = append(webHooksObj, webHookObj)
	}
	return webHooksObj

}

func (*webhookService) GetWebHook(clientId int, clientCode string, webhookid int) structs.WebHookObj {
	var webHookObj structs.WebHookObj
	//var webHookEventsObj []structs.Event
	//clientID := 0 //clientRepo.GetClientID(clientCode)

	webHooksPlainRtn := webhookRepo.GetWebHook(clientId, webhookid)
	webHookObj.ID = webHooksPlainRtn.ID
	webHookObj.ClientID = webHooksPlainRtn.ClientID
	webHookObj.ClientCode = webHooksPlainRtn.ClientCode
	webHookObj.URL = webHooksPlainRtn.URL
	webHookObj.Method = webHooksPlainRtn.Method
	webHookObj.HeaderPrefix = webHooksPlainRtn.HeaderPrefix

	err1 := json.Unmarshal([]byte(webHooksPlainRtn.Events), &webHookObj.Events)
	if err1 != nil {
		fmt.Println(err1.Error())
	}
	//webHookObj.Events = webHookEventsObj

	//eventsJson, _ := json.Marshal([]byte(webHooksPlainRtn.Events))
	//webHookObj.Events = eventsJson
	//webHookObj.Events = webHooksPlainRtn.Events

	webHookObj.HttpCode = webHooksPlainRtn.HttpCode
	webHookObj.Retry = webHooksPlainRtn.Retry
	webHookObj.Timeout = webHooksPlainRtn.Timeout
	webHookObj.Status = webHooksPlainRtn.Status
	webHookObj.CreatedAt = webHooksPlainRtn.CreatedAt
	webHookObj.CreatedBy = webHooksPlainRtn.CreatedBy

	return webHookObj
}

func (*webhookService) ValidateSave(wh structs.WebHookObj) (structs.WebHookObj, *structs.ErrorMessage) {
	var errors structs.ErrorMessage
	v := validator.New()
	err := v.Struct(wh)
	if err != nil {
		errDetail := err.(validator.ValidationErrors)[0]
		errors.Message = structs.Validate
		errors.Data = ""
		switch errDetail.Tag() {
		case "oneof":
			errors.SysMessage = fmt.Sprintf("validation error with key : %s allowed value %s : %s", errDetail.Namespace(), errDetail.Tag() ,errDetail.Param())
		default:
			errors.SysMessage = fmt.Sprintf("validation error with key : %s %s", errDetail.Namespace(), errDetail.Tag())
		}
		errors.Code = http.StatusInternalServerError
		return wh, &errors
	}
	if wh.ClientCode == "" {
		errors.Message = structs.ClientErr
		errors.Data = ""
		errors.SysMessage = structs.ClientErr
		errors.Code = http.StatusInternalServerError
		return wh, &errors
	}
	return wh, nil
}

func (*webhookService) SaveWebHook(wh structs.WebHookObj) *structs.ErrorMessage {
	var errors structs.ErrorMessage

	isExist := true
	if wh.ID == 0 {
		isExist = false
	}
	/*
		_, vss, errorstr := rewardRepo.GetRewards(getrwds)
		if errorstr != nil {
			errors.Message = errorstr.Message
			errors.SysMessage = errorstr.SysMessage
			errors.Code = http.StatusInternalServerError

			return &errors
		}
		if vss == nil {
			isRewardCodeExists = false
		}
	*/

	var body []byte
	sCase := isExist
	switch sCase {
	case false:
		errorstr := webhookRepo.AddWebHook(wh, string(body))
		if errorstr != nil {
			if errorstr.Message == "success" {
				errors.Data = errorstr.Data
				errors.Message = structs.SuccessInsert
			} else {
				errors.Message = errorstr.Message
				errors.SysMessage = errorstr.SysMessage
				errors.Code = http.StatusInternalServerError
	
				return &errors
			}
		}
		errors.Message = structs.SuccessInsert
	default:
		errorstr := webhookRepo.UpdateWebHook(wh, string(body))
		if errorstr != nil {
			if errorstr.Message == "success" {
				errors.Message = structs.SuccessUpdate
			} else {
				errors.Message = errorstr.Message
				errors.SysMessage = errorstr.SysMessage
				errors.Code = http.StatusInternalServerError
	
				return &errors
			}

		}
	}

	errors.SysMessage = string("")
	errors.Code = http.StatusOK
	return &errors
}

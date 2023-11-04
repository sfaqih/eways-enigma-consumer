package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"gitlab.com/wgroup1/enigma/common"
	service "gitlab.com/wgroup1/enigma/services"
	"gitlab.com/wgroup1/enigma/structs"
	// "gopkg.in/go-playground/validator.v9"
)

type webHookLogic struct{}

var (
	webhookService service.WebhookService
)

type WebHookLogic interface {
	GetWebHooks(w http.ResponseWriter, r *http.Request)
	GetWebHook(w http.ResponseWriter, r *http.Request)
	SaveWebHook(w http.ResponseWriter, r *http.Request)
}

func NewWebHookLogic(service service.WebhookService) WebHookLogic {
	webhookService = service
	return &webHookLogic{}
}

func (*webHookLogic) GetWebHooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var webHookParamObj structs.WebHookParamObj
	var errMsg structs.ErrorMessage
	var payload structs.RequestEncrypt
	var webhookBytes []byte
	var response structs.ResponseEncrypt
	hashKey := r.Header.Get("X-Hash-Key")
	encryptMode, _ := strconv.ParseBool(r.Header.Get("X-Encrypt-Mode")) 

	switch encryptMode {
	case true:
		_ = json.NewDecoder(r.Body).Decode(&payload)

		if payload.Payload == "" {
			errMsg.Message = "Payload encryption can't be null"
			errMsg.Code = http.StatusBadRequest
			errMsg.SysMessage = "Encryption mode is true, but payload encryption is empty"
			w.WriteHeader(errMsg.Code)
			json.NewEncoder(w).Encode(errMsg)
			return
		}

		request, _ :=  common.DecryptAES(payload.Payload, []byte(hashKey))
		json.Unmarshal([]byte(request), &webHookParamObj)
		webHooksPlain := webhookService.GetWebHooks(webHookParamObj.ClientId, webHookParamObj.ClientCode, webHookParamObj.Page, webHookParamObj.Limit)
		webhookBytes, _ = json.Marshal(webHooksPlain)
		webhooks, _ := common.EncryptAES(string(webhookBytes), []byte(hashKey))

		response.Data = webhooks
		response.Status = "OK"
		response.Code = http.StatusOK
		json.NewEncoder(w).Encode(response)
		return
	default:
		_ = json.NewDecoder(r.Body).Decode(&webHookParamObj)
		// validate := common.ValidateStruct(webHookParamObj)
		if err := common.ValidateStruct(webHookParamObj); err != nil {
			errMsg.Message = structs.Validate
			errMsg.Data = ""
			errMsg.SysMessage = err.Error()
			errMsg.Code = http.StatusBadRequest
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errMsg)
			return
		}
		// v := validator.New()
		// err := v.Struct(webHookParamObj)
		// if err != nil {
		// 	errMsg.Message = structs.Validate
		// 	errMsg.Data = ""
		// 	errMsg.SysMessage = err.Error()
		// 	errMsg.Code = http.StatusInternalServerError
		// 	json.NewEncoder(w).Encode(errMsg)
		// 	return
		// }
		webHooksPlain := webhookService.GetWebHooks(webHookParamObj.ClientId, webHookParamObj.ClientCode, webHookParamObj.Page, webHookParamObj.Limit)
		json.NewEncoder(w).Encode(webHooksPlain)
		return
	}

}
func (*webHookLogic) GetWebHook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var webHookParamObj structs.WebHookParamObj
	var payload structs.RequestEncrypt
	var errMsg structs.ErrorMessage
	var response structs.ResponseEncrypt
	var webhookBytes []byte
	hashKey := r.Header.Get("X-Hash-Key")
	encryptMode, _ := strconv.ParseBool(r.Header.Get("X-Encrypt-Mode"))

	switch encryptMode {
	case true:
		_ = json.NewDecoder(r.Body).Decode(&payload)
		request, _ :=  common.DecryptAES(payload.Payload, []byte(hashKey))
		json.Unmarshal([]byte(request), &webHookParamObj)

		if payload.Payload == "" {
			errMsg.Message = "Payload encryption can't be null"
			errMsg.Code = http.StatusBadRequest
			errMsg.SysMessage = "Encryption mode is true, but payload encryption is empty"
			w.WriteHeader(errMsg.Code)
			json.NewEncoder(w).Encode(errMsg)
			return
		}

		webHooksPlain := webhookService.GetWebHook(webHookParamObj.ClientId, webHookParamObj.ClientCode, webHookParamObj.ID)
		webhookBytes, _ = json.Marshal(webHooksPlain)
		webhooks, _ := common.EncryptAES(string(webhookBytes), []byte(hashKey))
		response.Data = webhooks
		json.NewEncoder(w).Encode(response)
		return
	default:
		_ = json.NewDecoder(r.Body).Decode(&webHookParamObj)
		//iClientId, _ := strconv.Atoi(getbanks.ClientID)
		webHooksPlain := webhookService.GetWebHook(webHookParamObj.ClientId, webHookParamObj.ClientCode, webHookParamObj.ID)
		json.NewEncoder(w).Encode(webHooksPlain)
		return
	}


}

func (*webHookLogic) SaveWebHook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var wh structs.WebHookObj
	var errs []structs.ErrorMessage
	var payload structs.RequestEncrypt
	var errMsg structs.ErrorMessage
	var webhookBytes []byte
	var response structs.ResponseEncrypt
	hashKey := r.Header.Get("X-Hash-Key")
	encryptMode, _ := strconv.ParseBool(r.Header.Get("X-Encrypt-Mode"))

	switch encryptMode {
	case true:
		_ = json.NewDecoder(r.Body).Decode(&payload)
		request, _ :=  common.DecryptAES(payload.Payload, []byte(hashKey))
		json.Unmarshal([]byte(request), &wh)

		if payload.Payload == "" {
			errMsg.Message = "Payload encryption can't be null"
			errMsg.Code = http.StatusBadRequest
			errMsg.SysMessage = "Encryption mode is true, but payload encryption is empty"
			w.WriteHeader(errMsg.Code)
			json.NewEncoder(w).Encode(errMsg)
			return
		}

		_, err1 := webhookService.ValidateSave(wh)
		if err1 != nil {
			errs = append(errs, *err1)
			common.JSONErrs(w, &errs)
			return
		}
		err2 := webhookService.SaveWebHook(wh)
		errs = append(errs, *err2)
		webhookBytes, _ = json.Marshal(errs)
		webhooks, _ := common.EncryptAES(string(webhookBytes), []byte(hashKey))
		if err2.Code != 200 {
			w.WriteHeader(err2.Code)
			json.NewEncoder(w).Encode(errs)
			return
		}
		response.Data = webhooks
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(response)

	default:
		_ = json.NewDecoder(r.Body).Decode(&wh)

		_, err1 := webhookService.ValidateSave(wh)
		if err1 != nil {
			errs = append(errs, *err1)
			common.JSONErrs(w, &errs)
			return
		}
	
		err2 := webhookService.SaveWebHook(wh)
		errs = append(errs, *err2)
		common.JSONErrs(w, &errs)
	}

}

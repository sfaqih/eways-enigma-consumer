package api

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"sync"
	"time"

	"gitlab.com/wgroup1/enigma/common"
	"gitlab.com/wgroup1/enigma/repositories/redis"
	service "gitlab.com/wgroup1/enigma/services"
	"gitlab.com/wgroup1/enigma/structs"
	"gopkg.in/go-playground/validator.v9"
)

type inbWALogic struct{
	inbWAService service.InboundWAService
	DB *sql.DB
}

// var (
	
// 	//inbWARepo    interfaces.InboundWARepository = repositories.NewInboundWARepository()
// )

type InboundWALogic interface {
	CreateInboundWA(w http.ResponseWriter, r *http.Request)
	DamcorpInboundWA(w http.ResponseWriter, r *http.Request)
	DamcorpGetMedia(w http.ResponseWriter, r *http.Request)
}

func NewInboundWALogic(service service.InboundWAService, db *sql.DB) InboundWALogic {
	// inbWAService = service
	return &inbWALogic{
		inbWAService: service,
		DB: db,
	}
}

func (ibwa *inbWALogic) CreateInboundWA(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var inbs structs.InboundWA
	var errs []structs.ErrorMessage
	var wg sync.WaitGroup

	_ = json.NewDecoder(r.Body).Decode(&inbs)
	cliID := inbs.ClientID

	for j := range inbs.Messages {
		_, errStr := ibwa.inbWAService.Validate(&inbs.Messages[j])
		jsonMsg, _ := json.Marshal(inbs.Messages[j])
		if errStr != nil {
			errs = append(errs, *errStr)
			continue
		}

		wg.Add(1)
		go func(count int64) {
			defer wg.Done()
			errStr = ibwa.inbWAService.SendInboundWA(inbs.Messages[j], cliID)
			if errStr.Code != http.StatusOK {
				errs = append(errs, *errStr)
				//continue
			} else {
				errs = append(errs, structs.ErrorMessage{Data: string(jsonMsg), Message: structs.Success, SysMessage: "", Code: http.StatusOK})
			}
			//fmt.Println("worker count:", count)
		}(int64(1))

		errStr = ibwa.inbWAService.CreateInboundWA(inbs.Messages[j], cliID)
		fmt.Printf(errStr.SysMessage)
		wg.Wait()
	}
	common.JSONErrs(w, &errs)
	//return
}

func (ibwa *inbWALogic) DamcorpInboundWA(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//var reqBody structs.Damcorp_Inbound_Request_WA
	var errs structs.ErrorMessage
	var body_plain interface{}
	//var body_plain []byte
	//_ = json.NewDecoder(r.Body).Decode(&inbs)
	json.NewDecoder(r.Body).Decode(&body_plain)

	// reqBody.ClientID = 41

	body_plainmarsh, err := json.Marshal(body_plain)

	if err != nil {
		log.Println("Inbound Dto row 84 " + err.Error())
		errs.Message = err.Error()
		errs.Code = 500

		w.WriteHeader(500)
		json.NewEncoder(w).Encode(errs)
		return
	}

	// vs, _ := inboundService.GetWebhook("WhatsAppDto", "InboundDto", "SesameDto", "InboundDto")
	// reqBody.Webhook = vs

	// CATATAN.... idupin redis...
	redis.RPush(common.DAMCORP_INBOUND_WA, body_plainmarsh, common.REDIS_DB)

	// _ = json.NewDecoder(r.Body).Decode(&reqBody)

	// inbWAService.SendInboundWADamcorp(reqBody)

	log.Println("Inbound From DTO")

	errs.Message = structs.Success
	errs.Code = 200

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(errs)

}

func (ibwa *inbWALogic) DamcorpGetMedia(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var errs structs.ErrorMessage

	var reqBody structs.Damcorp_Request_Media

	var (
		resp *http.Response
		err error
		body []byte
	)


	json.NewDecoder(r.Body).Decode(&reqBody)

	v := validator.New()
	err2 := v.Struct(reqBody)
	if err2 != nil {
		errs.Message = structs.Validate
		errs.Data = ""
		errs.SysMessage = err.Error()
		errs.Code = http.StatusBadRequest
		w.WriteHeader(errs.Code)
		json.NewEncoder(w).Encode(errs)
		return
	}

	vs, err1 := outboundService.GetToken(reqBody.ClientId)

	if err1.Code != 200 {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(err1)
		return
	}

	t := make(map[string]structs.VendorService)
	for k := range vs {
		t[vs[k].Channel] = vs[k]
	}

	switch reqBody.Channel {
	case "DamcorpMedia":
		url :=  strings.Replace(t["DamcorpMedia"].URI, "{{media_id}}", reqBody.MediaId, -1)
		token := "Bearer " + t["DamcorpMedia"].Token
		_, resp, body, _, err = common.HitAPI(url, "", t["DamcorpMedia"].Method, token, t["DamcorpMedia"].Channel, time.Duration(0), ibwa.DB)
	default:
		errs.Message = "Channel not allow to access the data"
		errs.Data = ""
		errs.SysMessage = err.Error()
		errs.Code = http.StatusBadRequest
		w.WriteHeader(errs.Code)
		json.NewEncoder(w).Encode(errs)
		return
	}

	if err != nil {
		if resp != nil {
			errs.Message = resp.Status
			errs.Code = resp.StatusCode
		}else{
			errs.Message = err.Error()
			errs.Code = resp.StatusCode
		}

		w.WriteHeader(errs.Code)
		json.NewEncoder(w).Encode(errs)

		return
	}

	// Encode the byte slice as a base64 string
	encoded := base64.StdEncoding.EncodeToString(body)

	
	errs.Data = encoded

	if resp == nil {
		errs.Message = "200"
		errs.Code = 200
	} else {
		errs.Message = resp.Status
		errs.Code = resp.StatusCode
	}

	w.WriteHeader(errs.Code)
	json.NewEncoder(w).Encode(errs)
}

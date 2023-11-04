package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"gitlab.com/wgroup1/enigma/common"
	"gitlab.com/wgroup1/enigma/repositories/redis"
	service "gitlab.com/wgroup1/enigma/services"
	"gitlab.com/wgroup1/enigma/structs"
)

type outboundLogic struct{}

var (
	outboundService service.OutboundService
	clientService   service.ClientService
)

type OutboundLogic interface {
	//CreateOutbound(w http.ResponseWriter, r *http.Request)
	CreateOutboundConsumer(w http.ResponseWriter, r *http.Request)
	CreateOutboundSynch(w http.ResponseWriter, r *http.Request)
	CreateFlowSynch(w http.ResponseWriter, r *http.Request)
	CreateOutboundConsumerBulk(w http.ResponseWriter, r *http.Request)
	CloseSessionWA(w http.ResponseWriter, r *http.Request)
}

func NewOutboundLogic(service service.OutboundService, flservice service.FlowService, clService service.ClientService) OutboundLogic {
	outboundService = service
	flowService = flservice
	clientService = clService
	return &outboundLogic{}
}

func (*outboundLogic) CreateOutboundSynch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var outb structs.Outbound
	var errs []structs.ErrorMessage
	var wg sync.WaitGroup

	var cliID = 0
	// masuk ke channel outbound biasa
	_ = json.NewDecoder(r.Body).Decode(&outb)
	cliID = outb.ClientID
	//Assumption: 1 vendor only have 1 token format (it will not differ based on channel)
	//but, 1 client could have more than 1 vendor, but only 1 vendor for 1 channel
	vs, errgettoken := outboundService.GetToken(cliID)
	if errgettoken != nil {
		if errgettoken.Code == 200 {
			//convert struct.authvendor to maps
			t := make(map[string]structs.VendorService)
			for k := range vs {
				t[vs[k].Channel] = vs[k]
			}

			for j := range outb.Messages {
				_, errStr := outboundService.Validate(&outb.Messages[j])
				if errStr != nil {
					errs = append(errs, *errStr)
					continue
				}

				outb.Messages[j].TrxId = common.RandomString("", 32, "", 0, 1)

				//// Ini sebelum di ubah menggunakan enigma consumer
				////send API to Vendor
				wg.Add(1)
				go func(count int64) {
					defer wg.Done()
				errStr = outboundService.SendOutboundSgl(outb.Messages[j], t, cliID)
				errs = append(errs, *errStr)
				}(int64(1))

				wg.Wait()
			}
		} else {
			errs = append(errs, *errgettoken)
		}
	}
	//fmt.Println(errs)
	common.JSONErrs(w, &errs)
}

func (*outboundLogic) CreateFlowSynch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var outbflow structs.OutboundFlow
	var errs []structs.ErrorMessage

	_ = json.NewDecoder(r.Body).Decode(&outbflow)

	// masuk ke channel outbound FLOW
	cliID := clientService.GetClientID(outbflow.ClientCode)
	//Assumption: 1 vendor only have 1 token format (it will not differ based on channel)
	//but, 1 client could have more than 1 vendor, but only 1 vendor for 1 channel
	vs, errgettoken := outboundService.GetToken(cliID)
	if errgettoken != nil {
		if errgettoken.Code == 200 {
			//convert struct.authvendor to maps
			t := make(map[string]structs.VendorService)
			for k := range vs {
				t[vs[k].Channel] = vs[k]
			}
			outbflow.ClientID = cliID

			errStr := outboundService.SendOutboundFlowSgl(outbflow, t, cliID)
			errs = append(errs, *errStr)
		} else {
			errs = append(errs, *errgettoken)
		}
	}
	common.JSONErrs(w, &errs)
}

func (*outboundLogic) CreateOutboundConsumer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var outb structs.Outbound
	var outbconsumer structs.OutboundConsumer

	var errs []structs.ErrorMessage

	_ = json.NewDecoder(r.Body).Decode(&outbconsumer)

	cliID := outbconsumer.ClientID

	//Assumption: 1 vendor only have 1 token format (it will not differ based on channel)
	//but, 1 client could have more than 1 vendor, but only 1 vendor for 1 channel
	vs, errgettoken := outboundService.GetToken(cliID)
	if errgettoken != nil {
		if errgettoken.Code == 200 {
			//convert struct.authvendor to maps
			t := make(map[string]structs.VendorService)
			for k := range vs {
				t[vs[k].Channel] = vs[k]
			}

			//Outbound biasa
			iLenMessage := len(outbconsumer.Messages)
			iCtr := 1
			iCtr2 := 1
			iMaxHit := 200
			var errpush structs.ErrorMessage
			var outbconsumernew structs.OutboundConsumer

			for j := range outbconsumer.Messages {
				_, errStr := outboundService.Validate(&outbconsumer.Messages[j])
				if errStr != nil {
					errs = append(errs, *errStr)
					continue
				}

				outbconsumernew.ClientID = cliID
				//outbconsumernew.Messages[0].TrxId = common.RandomString("", 32, "", 0, 1)
				outbconsumer.Messages[j].TrxId = common.RandomString("", 32, "", 0, 1)
				outbconsumernew.Messages = append(outbconsumernew.Messages, outbconsumer.Messages[j])

				outbconsumernew.VendorService = vs
				outbconsumerByte, _ := json.Marshal(outbconsumernew)
				json.Unmarshal(outbconsumerByte, &outb)

				clientReq := outbconsumernew
				clientReq.VendorService = []structs.VendorService{}

				clientReqJson, _ := json.Marshal(clientReq)

				//outbconsumer.Messages[j].TrxId = common.RandomString("", 32, "", 0, 1)
				//outbconsumer.VendorService = vs
				//outbconsumerByte, _ := json.Marshal(outbconsumer)
				//json.Unmarshal(outbconsumerByte, &outb)

				////outboundService.SendOutboundSgl(outbconsumer.Messages[j], t, cliID)
				////errs = append(errs, *errStr)

				if iCtr == iMaxHit || iCtr2 == iLenMessage {
					redis.RPush(common.OUTBOUND_QUEUE, outbconsumerByte, common.REDIS_DB)
					iCtr = 1
					outbconsumernew = structs.OutboundConsumer{}

				} else {
					iCtr += 1
				}
				iCtr2 += 1



				errpush.ReqSesame = string(clientReqJson)
				errpush.Message = "success"
				// errpush.RespMessage = string(clientReqJson)
				// errpush.ReqMessage = string(clientReqJson)
				errpush.Code = 200

				errs = append(errs, errpush)
			}

		} else {
			errs = append(errs, *errgettoken)
		}
	}
	common.JSONErrs(w, &errs)

}

func (*outboundLogic) CreateOutboundConsumerBulk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	startAll := time.Now()
	var outbconsumer structs.OutboundConsumer

	var errs []structs.ErrorMessage

	_ = json.NewDecoder(r.Body).Decode(&outbconsumer)

	cliID := outbconsumer.ClientID


	//Assumption: 1 vendor only have 1 token format (it will not differ based on channel)
	//but, 1 client could have more than 1 vendor, but only 1 vendor for 1 channel
	vs, errgettoken := outboundService.GetToken(cliID)
	if errgettoken != nil && len(vs) > 0 {
		if errgettoken.Code == 200 {
			

			iLenMessage := len(outbconsumer.Messages)
			var errpush structs.ErrorMessage
			var outbconsumernew structs.OutboundConsumer

			for _, msg := range outbconsumer.Messages {
				_, errStr := outboundService.Validate(&msg)
				if errStr != nil {
					errs = append(errs, *errStr)
					continue
				}

				msg.TrxId = common.RandomString("", 32, "", 0, 1)
				outbconsumernew.Messages = append(outbconsumernew.Messages, msg)

			}

			outbconsumernew.ClientID = cliID
			outbconsumernew.VendorService = vs
			outbconsumerByte, _ := json.Marshal(outbconsumernew)
			redis.RPush(common.OUTBOUND_QUEUE_BULK, outbconsumerByte, common.REDIS_DB)

			errpush.Code = http.StatusOK
			errpush.Message = "success"
			totalData := strconv.Itoa(iLenMessage)
			outbconsumernew.VendorService = []structs.VendorService{}
			requestJsonByte, _ := json.Marshal(outbconsumernew)
			errpush.ReqSesame = string(requestJsonByte)
			errpush.SysMessage = "Total Data : " + totalData + " and Data blast is currently in the queueing process."

			errs = append(errs, errpush)


		} else {
			errs = append(errs, *errgettoken)
		}
	}else{
		var errMsg = structs.ErrorMessage{
			Code: http.StatusBadRequest,
			Message: "something wrong with request format",
			SysMessage: "BAD REQUEST",
		}
		errs = append(errs, errMsg)
	}
	elapsedAll := time.Since(startAll)
	log.Println("outbound bulk time : ", elapsedAll)
	common.JSONErrs(w, &errs)

}

func (*outboundLogic) CloseSessionWA(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var (
		csrqs structs.Damcorp_Close_Session_Request
		cs    structs.Damcorp_Close_Session
		errs  *structs.ErrorMessage
	)

	_ = json.NewDecoder(r.Body).Decode(&csrqs)

	cliID := csrqs.ClientID

	vs, errgettoken := outboundService.GetToken(cliID)

	if errgettoken.Code == 200 {
		t := make(map[string]structs.VendorService)
		for k := range vs {
			t[vs[k].Channel] = vs[k]
		}

		cs.Msisdn = csrqs.Msisdn

		errs = outboundService.CloseSessionWA(cs, t)

	} else {
		errs = errgettoken
	}

	reqBody, _ := json.Marshal(cs)
	errs.ReqMessage = string(reqBody)

	w.WriteHeader(errs.Code)
	json.NewEncoder(w).Encode(errs)
}

package api

import (
	"encoding/json"
	"net/http"
	"sync"

	"gitlab.com/wgroup1/enigma/common"
	service "gitlab.com/wgroup1/enigma/services"
	"gitlab.com/wgroup1/enigma/structs"
)

type flowLogic struct{}

var (
	flowService service.FlowService
)

type FlowLogic interface {
	CreateFlow(w http.ResponseWriter, r *http.Request)
	//	CreateOutboundConsumer(w http.ResponseWriter, r *http.Request)
}

func NewFlowLogic(service service.FlowService) FlowLogic {
	flowService = service
	return &flowLogic{}
}

func (*flowLogic) CreateFlow(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var outb structs.OutboundConsumer
	var errs []structs.ErrorMessage
	var wg sync.WaitGroup

	_ = json.NewDecoder(r.Body).Decode(&outb)
	cliID := outb.ClientID

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
	common.JSONErrs(w, &errs)
}

/*
func (*outboundLogic) CreateOutboundConsumer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var outb structs.Outbound
	var outbflowconsumer structs.OutboundFlowConsumer
	var errs []structs.ErrorMessage

	_ = json.NewDecoder(r.Body).Decode(&outb)
	cliID := outb.ClientID

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

				var errpush structs.ErrorMessage
				outbByte, _ := json.Marshal(outb)
				json.Unmarshal(outbByte, &outbflowconsumer)
				outbflowconsumer.VendorService = vs
				//outbflowconsumerbyte, _ := json.Marshal(outbflowconsumer)

				outboundService.SendOutboundSgl(outbflowconsumer.Messages[j], t, cliID)
				//errs = append(errs, *errStr)
				//redis.RPush(common.OUTBOUND_QUEUE, outbflowconsumerbyte, common.REDIS_DB)

				errpush.ReqSesame = string(outbByte)
				errpush.Message = "success"
				errpush.RespMessage = string(outbByte)
				errpush.ReqMessage = string(outbByte)
				errpush.Code = 200

				errs = append(errs, errpush)
			}
		} else {
			errs = append(errs, *errgettoken)
		}
	}
	common.JSONErrs(w, &errs)

}
*/

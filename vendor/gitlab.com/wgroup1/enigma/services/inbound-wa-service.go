package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"gitlab.com/wgroup1/enigma/common"
	"gitlab.com/wgroup1/enigma/interfaces"
	"gitlab.com/wgroup1/enigma/repositories/mysql"
	"gitlab.com/wgroup1/enigma/services/pb"
	"gitlab.com/wgroup1/enigma/structs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/go-playground/validator.v9"
)

type InboundWAService interface {
	CreateInboundWA(inb structs.Message, cli int) *structs.ErrorMessage
	Validate(inb *structs.Message) (*structs.Message, *structs.ErrorMessage)
	SendInboundWA(inb structs.Message, cli int) *structs.ErrorMessage
	SendInboundWADamcorp(iw []byte, vss []structs.WebHook) *structs.ErrorMessage
}

type inbWAService struct{
	inbWARepo interfaces.InboundWARepository
	DB *sql.DB
}


func NewInboundWAService(repository interfaces.InboundWARepository, db *sql.DB) InboundWAService {
	return &inbWAService{
		inbWARepo: repository,
		DB: db,
	}
}

func (ibwa *inbWAService) CreateInboundWA(inb structs.Message, cli int) *structs.ErrorMessage {
	return ibwa.inbWARepo.CreateInboundWA(inb, cli)
}

func (*inbWAService) Validate(inb *structs.Message) (*structs.Message, *structs.ErrorMessage) {
	var errors structs.ErrorMessage
	v := validator.New()
	err := v.Struct(inb)
	if err != nil {
		errors.Message = structs.Validate
		errors.Data = ""
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return nil, &errors
	}
	return inb, nil
}

func (ibwa *inbWAService) SendInboundWA(inb structs.Message, cli int) *structs.ErrorMessage {
	var errors structs.ErrorMessage
	//	conn, err := grpc.Dial(mysql.ViperEnvVariable("SESAME_GRPC_URL"), grpc.WithInsecure())
	conn, err := grpc.Dial(mysql.ViperEnvVariable("SESAME_GRPC_URL"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	//WithTransportCredentials and insecure.NewCredentials()
	//conn, err := grpc.Dial("localhost:50005", grpc.WithInsecure())

	if err != nil {
		fmt.Println("cannot connect with server:", err)
	}

	client := pb.NewStreamServiceClient(conn)

	jsonreq, _ := json.Marshal(inb)

	in := &pb.Request{Jsonreq: string(jsonreq)}

	stream, err := client.FetchResponse(context.Background(), in)
	if err != nil {
		errors.Message = string(jsonreq)
		errors.Data = ""
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return &errors
	}

	done := make(chan bool)

	//received response
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				done <- true //means stream is finished
				return
			}
			if err != nil {
				fmt.Println("cannot receive:", err)
			}
			log.Printf("Resp received: %s", resp.Jsonresponse)

		}
	}()
	<-done

	errors.Message = structs.Success
	errors.Code = http.StatusOK

	return &errors
}

func (ibwa *inbWAService) SendInboundWADamcorp(iw []byte, vss []structs.WebHook) *structs.ErrorMessage {
	err := ibwa.inbWARepo.CreateInboundWADamcorp(iw, vss)

	//vss := iw.Webhook
	if err != nil {
		if err.Code == 200 {
			chkws := false
			_ = chkws
			//_, resp, _, _, err = common.HitAPI(outm.CallBackAuth, string(rspcallbackbyte), time.Duration(0))
			for iLoop, wh := range vss {
				for iwh := 0; iwh < wh.Retry; iwh++ {

					//iwbyte, _ := json.Marshal(iw.ReqData)
					_, _, _, _, errhit := common.HitAPI(wh.URL, string(iw), wh.Method, wh.Token, "", time.Duration(0), ibwa.DB)
					//common.HitAPI(wh.URL+uri, string(iiwsesame), vss[iLoop].Method, vss[iLoop].Token, "", time.Duration(0))
					if errhit == nil {
						chkws = true
						iwh = vss[iLoop].Retry
					} else {
						fmt.Println("retry at:", time.Now())
					}
					//}
				}
			}
		}
	}

	//common.HitAPI(outm.CallBackAuth, string(rspcallbackbyte), time.Duration(0))
	return err
}

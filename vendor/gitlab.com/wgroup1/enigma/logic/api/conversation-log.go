package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	// "sync"

	"gitlab.com/wgroup1/enigma/common"
	"gitlab.com/wgroup1/enigma/repositories/redis"
	service "gitlab.com/wgroup1/enigma/services"
	"gitlab.com/wgroup1/enigma/structs"
)

type conversationLogic struct{}

var (
	conversationService service.ConversationService
)

type ConversationLogic interface {
	ReceiveConversation(w http.ResponseWriter, r *http.Request)
}

func NewConversationLogic(service service.ConversationService) ConversationLogic {
	conversationService = service
	return &conversationLogic{}
}

func (*conversationLogic) ReceiveConversation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var conv structs.Incoming
	var errs []structs.ErrorMessage
	var convConsumer structs.ConversationDevilery
	// var wg sync.WaitGroup

	_ = json.NewDecoder(r.Body).Decode(&conv)
	
	// fmt.Println("request body:", string(bdy))
	fmt.Printf("[Conv] ChannelID:%s \n", conv.Message.ChannelID)
	fmt.Printf("[Conv] Platform:%s \n", conv.Message.Platform)
	//send API to Vendor
	// wg.Add(1)
	// go func(count int64) {
		// defer wg.Done()
	errStr, whs := conversationService.ReceiveConversation(conv)
	errs = append(errs, *errStr)

	convConsumer.Conversation = conv
	convConsumer.Webhooks = whs
	convconsumerByte, _ := json.Marshal(&convConsumer)

	if errStr.Code != 200 {
		log.Printf("error insert conversation log : %s", errStr.SysMessage)
	}else {
		
		redis.RPush(common.CONVERSATION_QUEUE, convconsumerByte, common.REDIS_DB)
		log.Printf("success insert data to conversation log")
	}


	// }(int64(1))

	// wg.Wait()
	// conversationService.ConversationDevlivery(convconsumerByte)


	common.JSONErrs(w, &errs)
}

func (*conversationLogic) ReceiveConversationConsumer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var conv structs.Incoming
	var errs []structs.ErrorMessage
	//var wg sync.WaitGroup

	_ = json.NewDecoder(r.Body).Decode(&conv)
	bdy, _ := json.Marshal(conv)
	convconsumerByte, _ := json.Marshal(conv)
	fmt.Println("request body:", string(bdy))
	fmt.Printf("[Conv] ChannelID:%s \n", conv.Message.ChannelID)
	fmt.Printf("[Conv] Platform:%s \n", conv.Message.Platform)
	//send API to Consumer
	redis.RPush(common.CONVERSATION_QUEUE, convconsumerByte, common.REDIS_DB)

	/*

		wg.Add(1)
		go func(count int64) {
			defer wg.Done()
			errStr := conversationService.ReceiveConversation(conv)
			errs = append(errs, *errStr)
		}(int64(1))

		wg.Wait()
	*/

	common.JSONErrs(w, &errs)
}

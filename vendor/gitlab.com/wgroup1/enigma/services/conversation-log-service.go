package service

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"gitlab.com/wgroup1/enigma/common"
	"gitlab.com/wgroup1/enigma/interfaces"
	"gitlab.com/wgroup1/enigma/structs"
)

//this is conversationService func
type ConversationService interface {
	ReceiveConversation(conv structs.Incoming) (*structs.ErrorMessage, []*structs.WebHook)
	ConversationDevlivery(conv []byte)
}

type conversationService struct{
	conversationRepo interfaces.ConversationRepository
	DB *sql.DB
}


func NewConversationService(repository interfaces.ConversationRepository, db *sql.DB) ConversationService {
	return &conversationService{
		conversationRepo: repository,
		DB: db,
	}
}

func (repo *conversationService) ReceiveConversation(conv structs.Incoming) (*structs.ErrorMessage, []*structs.WebHook) {
	errs, whs := repo.conversationRepo.ReceiveConversation(conv)
	return errs, whs
}

func(repo *conversationService) ConversationDevlivery(conv []byte) {
	// var incoming structs.Incoming
	var convReceive structs.ConversationDevilery
	json.Unmarshal(conv, &convReceive)
	convByte, _ := json.Marshal(convReceive.Conversation)
	// incoming = convReceive.Conversation 
	// whs := repo.conversationRepo.GetWebhookConversation(&incoming)

	reqJson := string(convByte)

	for _, wh := range convReceive.Webhooks {

		iRetryStart := 0
		chkws := false
		for iwh := 0; iwh < wh.Retry; iwh++ {
			if !chkws {
				_, resp, _, _, err := common.HitAPI(wh.URL, reqJson, wh.Method, wh.Token, "", time.Duration(wh.Timeout), repo.DB)
				if err == nil {
					if resp.StatusCode == wh.HttpCode {
						chkws = true
					} else {
						if iRetryStart >= wh.Retry {
							chkws = true
						}
					}
				} else {
					if iRetryStart >= wh.Retry {
						chkws = true
					}
					log.Println("retry at:", time.Now())
				}
			}

		}
	}

	log.Println("whatsapp report status delivered to client")
}

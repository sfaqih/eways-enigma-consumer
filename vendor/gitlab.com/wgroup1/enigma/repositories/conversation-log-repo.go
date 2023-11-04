package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	// "fmt"
	"log"
	"net/http"
	"strings"
	// "time"

	// "gitlab.com/wgroup1/enigma/common"
	"gitlab.com/wgroup1/enigma/database"
	"gitlab.com/wgroup1/enigma/interfaces"
	"gitlab.com/wgroup1/enigma/structs"
)

type conversationRepo struct{
	DB *sql.DB
}

func NewConversationRepository(db *sql.DB) interfaces.ConversationRepository {
	return &conversationRepo{
		DB: db,
	}
}

func getClientID(channelId string, DB *sql.DB) int {
	var clientID int

	db := DB
	ctx := context.Background()

	cntQry := "select distinct client_id from client_vendors cv where channel_id = ?"
	err := db.QueryRowContext(ctx, cntQry, channelId).Scan(&clientID)

	if err != nil {
		log.Println(err.Error())
		return 0
	}

	return clientID
}

func (repo *conversationRepo) ReceiveConversation(conv structs.Incoming) (*structs.ErrorMessage, []*structs.WebHook) {
	var errors structs.ErrorMessage
	var msgJson []byte
	var contactJson []byte
	var cnvJson []byte
	var reqJson []byte
	var channel string = ""
	var whs []*structs.WebHook

	ctx := context.Background()
	tx, err := repo.DB.Begin()


	if err != nil {
		errors.Message = structs.QueryErr
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return &errors, whs
	}

	clientID := getClientID(conv.Message.ChannelID, repo.DB)

	sqlQuery := "INSERT INTO conversation_logs (client_id, contact_id, message_id, conversation_id, conversation_status, channel_id, platform, channel, direction, message_status, `from`, `to`, msisdn, contact_status, `type`, request_body, message_json, contact_json, conversation_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

	reqJson, _ = json.Marshal(conv)
	msgJson, _ = json.Marshal(conv.Message)
	contactJson, _ = json.Marshal(conv.Contact)
	cnvJson, _ = json.Marshal(conv.Conversation)

	plat := strings.ToLower(conv.Message.Platform)
	switch plat {
	case "whatsapp_sandbox":
		channel = "Whatsapp"
	case "whatsapp":
		//do for whatsapp on Production
		channel = "Whatsapp"
	case "facebook":
		channel = "Facebook"
	case "telegram":
		channel = "Telegram"
	}

	res, err := tx.ExecContext(ctx, sqlQuery, &clientID, &conv.Contact.ID, &conv.Message.ID, &conv.Conversation.ID, &conv.Conversation.Status, &conv.Message.ChannelID, &conv.Message.Platform, channel, &conv.Message.Direction, &conv.Message.Status, &conv.Message.From, &conv.Message.To, &conv.Contact.MSISDN, &conv.Contact.ContactStatus, &conv.Type, reqJson, msgJson, contactJson, cnvJson)
	lastID, _ := res.LastInsertId()

	if err != nil {
		log.Println(err.Error())
		tx.Rollback()
		errors.Message = structs.QueryErr
		errors.Data = string(reqJson)
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return &errors, whs
	}
	if err != nil {
		tx.Rollback()
		errors.Message = structs.LastIDErr
		errors.Data = string(reqJson)
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		log.Println(lastID)
		return &errors, whs
	}

	//start sisipan insert ke status_log

	//COMMENT DULU BIANG KEROK LOCK
	/*
		sts := structs.StatusReport{}
		sts.MessageID = conv.Message.ID
		sts.Event = conv.Message.Status
		sts.Message_GUID = conv.Conversation.ID


		if strings.ToLower(channel) == "whatsapp" && (conv.Message.Content.Hsm != nil) {
			var errors structs.ErrorMessage

			sts.Flowid = 0
			sts.ClientID = 0
			sts.Channel = "BlastWhatsapp"
			sts.Destination = strings.Replace(conv.Message.To, "+", "", 1)

			sts.MessageTemplate = conv.Message.Content.Hsm.TemplateName
			sts.Status = conv.Message.Status

			err := NewStatusReportRepository().AddStatusLogReport(sts, int(lastID))

			if err != nil {
				errors.Message = err.Message
				errors.RespMessage = err.RespMessage
				errors.ReqMessage = err.Data
				errors.Code = err.Code
				fmt.Println(errors)

			}

		}
	*/

	//end sisipan insert ke status_log

	// switch conv.Message.Status {
	// case "received":
	// 	tx.Commit()
	// case "pending":
	// 	statusQuery := "INSERT INTO status_pending (contact_id, message_id, conversation_id, conversation_status, channel_id, platform, channel, direction, message_status, `from`, `to`, msisdn, contact_status, `type`, request_body, message_json, contact_json, conversation_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

	// 	resStatus, err2 := tx.ExecContext(ctx, statusQuery, &conv.Contact.ID, &conv.Message.ID, &conv.Conversation.ID, &conv.Conversation.Status, &conv.Message.ChannelID, &conv.Message.Platform, channel, &conv.Message.Direction, &conv.Message.Status, &conv.Message.From, &conv.Message.To, &conv.Contact.MSISDN, &conv.Contact.ContactStatus, &conv.Type, reqJson, msgJson, contactJson, cnvJson)

	// 	if err2 != nil {
	// 		log.Println(err2.Error())
	// 		tx.Rollback()
	// 		errors.Message = structs.QueryErr
	// 		errors.Data = string(reqJson)
	// 		errors.SysMessage = err2.Error()
	// 		errors.Code = http.StatusInternalServerError
	// 		return &errors
	// 	}

	// 	lastID, err3 := resStatus.LastInsertId()
	// 	if err3 != nil {
	// 		tx.Rollback()
	// 		errors.Message = structs.LastIDErr
	// 		errors.Data = string(reqJson)
	// 		errors.SysMessage = err3.Error()
	// 		errors.Code = http.StatusInternalServerError
	// 		fmt.Println(lastID)
	// 		return &errors
	// 	}

	// 	tx.Commit()
	// case "sent":
	// 	statusQuery := "INSERT INTO status_sent (contact_id, message_id, conversation_id, conversation_status, channel_id, platform, channel, direction, message_status, `from`, `to`, msisdn, contact_status, `type`, request_body, message_json, contact_json, conversation_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

	// 	resStatus, err2 := tx.ExecContext(ctx, statusQuery, &conv.Contact.ID, &conv.Message.ID, &conv.Conversation.ID, &conv.Conversation.Status, &conv.Message.ChannelID, &conv.Message.Platform, channel, &conv.Message.Direction, &conv.Message.Status, &conv.Message.From, &conv.Message.To, &conv.Contact.MSISDN, &conv.Contact.ContactStatus, &conv.Type, reqJson, msgJson, contactJson, cnvJson)

	// 	if err2 != nil {
	// 		log.Println(err2.Error())
	// 		tx.Rollback()
	// 		errors.Message = structs.QueryErr
	// 		errors.Data = string(reqJson)
	// 		errors.SysMessage = err2.Error()
	// 		errors.Code = http.StatusInternalServerError
	// 		return &errors
	// 	}

	// 	lastID, err3 := resStatus.LastInsertId()
	// 	if err3 != nil {
	// 		tx.Rollback()
	// 		errors.Message = structs.LastIDErr
	// 		errors.Data = string(reqJson)
	// 		errors.SysMessage = err3.Error()
	// 		errors.Code = http.StatusInternalServerError
	// 		log.Println(lastID)
	// 		return &errors
	// 	}

	// 	tx.Commit()
	// case "delivered":
	// 	statusQuery := "INSERT INTO status_delivered (contact_id, message_id, conversation_id, conversation_status, channel_id, platform, channel, direction, message_status, `from`, `to`, msisdn, contact_status, `type`, request_body, message_json, contact_json, conversation_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

	// 	resStatus, err2 := tx.ExecContext(ctx, statusQuery, &conv.Contact.ID, &conv.Message.ID, &conv.Conversation.ID, &conv.Conversation.Status, &conv.Message.ChannelID, &conv.Message.Platform, channel, &conv.Message.Direction, &conv.Message.Status, &conv.Message.From, &conv.Message.To, &conv.Contact.MSISDN, &conv.Contact.ContactStatus, &conv.Type, reqJson, msgJson, contactJson, cnvJson)

	// 	if err2 != nil {
	// 		fmt.Println(err2.Error())
	// 		tx.Rollback()
	// 		errors.Message = structs.QueryErr
	// 		errors.Data = string(reqJson)
	// 		errors.SysMessage = err2.Error()
	// 		errors.Code = http.StatusInternalServerError
	// 		return &errors
	// 	}

	// 	lastID, err3 := resStatus.LastInsertId()
	// 	if err3 != nil {
	// 		tx.Rollback()
	// 		errors.Message = structs.LastIDErr
	// 		errors.Data = string(reqJson)
	// 		errors.SysMessage = err3.Error()
	// 		errors.Code = http.StatusInternalServerError
	// 		log.Println(lastID)
	// 		return &errors
	// 	}

	// 	tx.Commit()
	// case "read":
	// 	statusQuery := "INSERT INTO status_read (contact_id, message_id, conversation_id, conversation_status, channel_id, platform, channel, direction, message_status, `from`, `to`, msisdn, contact_status, `type`, request_body, message_json, contact_json, conversation_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

	// 	resStatus, err2 := tx.ExecContext(ctx, statusQuery, &conv.Contact.ID, &conv.Message.ID, &conv.Conversation.ID, &conv.Conversation.Status, &conv.Message.ChannelID, &conv.Message.Platform, channel, &conv.Message.Direction, &conv.Message.Status, &conv.Message.From, &conv.Message.To, &conv.Contact.MSISDN, &conv.Contact.ContactStatus, &conv.Type, reqJson, msgJson, contactJson, cnvJson)

	// 	if err2 != nil {
	// 		log.Println(err2.Error())
	// 		tx.Rollback()
	// 		errors.Message = structs.QueryErr
	// 		errors.Data = string(reqJson)
	// 		errors.SysMessage = err2.Error()
	// 		errors.Code = http.StatusInternalServerError
	// 		return &errors
	// 	}

	// 	lastID, err3 := resStatus.LastInsertId()
	// 	if err3 != nil {
	// 		tx.Rollback()
	// 		errors.Message = structs.LastIDErr
	// 		errors.Data = string(reqJson)
	// 		errors.SysMessage = err3.Error()
	// 		errors.Code = http.StatusInternalServerError
	// 		log.Println(lastID)
	// 		return &errors
	// 	}

	// 	tx.Commit()
	// case "failed":
	// 	statusQuery := "INSERT INTO status_failed (contact_id, message_id, conversation_id, conversation_status, channel_id, platform, channel, direction, message_status, `from`, `to`, msisdn, contact_status, `type`, request_body, message_json, contact_json, conversation_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

	// 	resStatus, err2 := tx.ExecContext(ctx, statusQuery, &conv.Contact.ID, &conv.Message.ID, &conv.Conversation.ID, &conv.Conversation.Status, &conv.Message.ChannelID, &conv.Message.Platform, channel, &conv.Message.Direction, &conv.Message.Status, &conv.Message.From, &conv.Message.To, &conv.Contact.MSISDN, &conv.Contact.ContactStatus, &conv.Type, reqJson, msgJson, contactJson, cnvJson)

	// 	if err2 != nil {
	// 		log.Println(err2.Error())
	// 		tx.Rollback()
	// 		errors.Message = structs.QueryErr
	// 		errors.Data = string(reqJson)
	// 		errors.SysMessage = err2.Error()
	// 		errors.Code = http.StatusInternalServerError
	// 		return &errors
	// 	}

	// 	lastID, err3 := resStatus.LastInsertId()
	// 	if err3 != nil {
	// 		tx.Rollback()
	// 		errors.Message = structs.LastIDErr
	// 		errors.Data = string(reqJson)
	// 		errors.SysMessage = err3.Error()
	// 		errors.Code = http.StatusInternalServerError
	// 		log.Println(lastID)
	// 		return &errors
	// 	}
	// 	tx.Commit()
	// }

	whs = getWebhookConversation(&conv, clientID, repo.DB)

	tx.Commit()

	// sCaseChannel := strings.ToLower(channel)
	// if sCaseChannel == "whatsapp_sandbox" {
	// 	sCaseChannel = "whatsapp"
	// }

	// queryWh := "select id, client_id, client_code, url, method, header_prefix, token, events, expected_http_code, retry, timeout "
	// queryWh += "from ( "
	// queryWh += "select * "
	// queryWh += "FROM webhooks, json_table(events, '$[*]' columns (idx FOR ORDINALITY, channel_name text path '$.channel_name', frm text path '$.from', `to` text path '$.to')) asd "
	// queryWh += "where status = 1 "
	// queryWh += "and client_id  = ? "
	// queryWh += "and trx_type = 'Conversation' "
	// queryWh += ") dsa "
	// queryWh += "where (frm = ? OR `to` = ?) "
	// queryWh += "and lower(channel_name) = ?"

	// resWh, err4 := repo.DB.QueryContext(ctx, queryWh, clientID, conv.Message.From, conv.Message.From, sCaseChannel)

	// defer database.CloseRows(resWh)

	// if err4 != nil {
	// 	log.Println(err4.Error())
	// 	errors.Message = structs.QueryErr
	// 	errors.Data = string(reqJson)
	// 	errors.SysMessage = err4.Error()
	// 	errors.Code = http.StatusInternalServerError
	// 	return &errors
	// }

	// for resWh.Next() {
	// 	wh := structs.WebHook{}
	// 	var eventsByte []byte
	// 	resWh.Scan(&wh.ID, &wh.ClientID, &wh.ClientCode, &wh.URL, &wh.Method, &wh.HeaderPrefix, &wh.Token, &eventsByte, &wh.HttpCode, &wh.Retry, &wh.Timeout)

	// 	iRetryStart := 0
	// 	chkws := false
	// 	if !chkws {
	// 		_, resp, _, _, err := common.HitAPI(wh.URL, string(reqJson), wh.Method, wh.Token, "", time.Duration(wh.Timeout), repo.DB)
	// 		if err == nil {
	// 			if resp.StatusCode == wh.HttpCode {
	// 				chkws = true
	// 			} else {
	// 				if iRetryStart >= wh.Retry {
	// 					chkws = true
	// 				}
	// 			}
	// 		} else {
	// 			if iRetryStart >= wh.Retry {
	// 				chkws = true
	// 			}
	// 			log.Println("retry at:", time.Now())
	// 		}
	// 	}
	// }

	errors.Message = structs.Success
	errors.Code = http.StatusOK

	return &errors, whs
}

func getWebhookConversation(conv *structs.Incoming, ClientId int, DB *sql.DB) []*structs.WebHook {
	var whs []*structs.WebHook
	var channel string = ""

	db := DB
	ctx := context.Background()

	clientID := ClientId

	plat := strings.ToLower(conv.Message.Platform)
	switch plat {
	case "whatsapp_sandbox":
		channel = "Whatsapp"
	case "whatsapp":
		//do for whatsapp on Production
		channel = "Whatsapp"
	case "facebook":
		channel = "Facebook"
	case "telegram":
		channel = "Telegram"
	}

	queryWh := "select id, client_id, client_code, url, method, header_prefix, token, events, expected_http_code, retry, timeout "
	queryWh += "from ( "
	queryWh += "select * "
	queryWh += "FROM webhooks, json_table(events, '$[*]' columns (idx FOR ORDINALITY, channel_name text path '$.channel_name', frm text path '$.from', `to` text path '$.to')) asd "
	queryWh += "where status = 1 "
	queryWh += "and client_id  = ? "
	queryWh += "and trx_type = 'Conversation' "
	queryWh += ") dsa "
	queryWh += "where (frm = ? OR `to` = ?) "
	queryWh += "and lower(channel_name) = ?"


	resp, err := db.QueryContext(ctx, queryWh, clientID, conv.Message.From, conv.Message.From, channel)

	defer database.CloseRows(resp)

	if err != nil {
		log.Println(err.Error())
		return whs
	}

	for resp.Next() {
		wh := structs.WebHook{}
		var eventsByte []byte
		resp.Scan(&wh.ID, &wh.ClientID, &wh.ClientCode, &wh.URL, &wh.Method, &wh.HeaderPrefix, &wh.Token, &eventsByte, &wh.HttpCode, &wh.Retry, &wh.Timeout)

		json.Unmarshal(eventsByte, &wh.Events)

		whs = append(whs, &wh)
	}

	return whs
}  

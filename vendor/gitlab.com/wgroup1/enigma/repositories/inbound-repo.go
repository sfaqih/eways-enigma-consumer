package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gitlab.com/wgroup1/enigma/interfaces"
	"gitlab.com/wgroup1/enigma/repositories/mysql"
	"gitlab.com/wgroup1/enigma/structs"
)

type inboundRepo struct{
	DB *sql.DB
}

func NewInboundRepository(db *sql.DB) interfaces.InboundRepository {
	return &inboundRepo{
		DB: db,
	}
}

func (inb *inboundRepo) CreateInbound(inm structs.Inbound, iiw structs.InternalInboundWa, respStr string, cli int, ctx context.Context) *structs.ErrorMessage {
	var errors structs.ErrorMessage
	if string(inm.Message.Direction) == "received" {

		tx, err := inb.DB.Begin()

		//start checking insert
		if err != nil {
			fmt.Println(err.Error())
			errors.Message = structs.QueryErr
			errors.SysMessage = err.Error()
			errors.Code = http.StatusInternalServerError
			return &errors
		}

		sqlQuery := "INSERT INTO inbound (`contactid`, `contact`, `conversation`, `message`, `conversationid`, `msisdn`, `messageid`, `channelId`, `jsonplain`, `createddate`) values(?,?,?,?,?,?,?,?,?,?)"

		//switch inm.Type {
		//case "hsm":
		//	typeJson, _ = json.Marshal(inm.Hsm)
		//case "text":
		//	//marshalling the json
		//case "image":
		//	//marshalling the json
		//}
		//typeJson, _ = json.Marshal(inm.Message)
		jsonMsg, _ := json.Marshal(inm)

		//cobaprint := &inm.Contact.Msisdn
		//fmt.Println(cobaprint)

		jsonDataContactJsonPlain, err1 := json.Marshal(inm.Contact)
		if err1 != nil {
			jsonDataContactJsonPlain = []byte(err1.Error())
		}

		jsonDataConversationJsonPlain, err2 := json.Marshal(inm.Conversation)
		if err2 != nil {
			jsonDataConversationJsonPlain = []byte(err2.Error())
		}

		jsonDataMessageJsonPlain, err3 := json.Marshal(inm.Message)
		if err3 != nil {
			jsonDataMessageJsonPlain = []byte(err3.Error())
		}

		jsonDataAllJsonPlain, err4 := json.Marshal(inm)
		if err4 != nil {
			jsonDataAllJsonPlain = []byte(err4.Error())
		}

		//res, err := tx.Exec(sqlQuery, &inm.Contact.ID, &inm.Contact, &inm.Conversation, &inm.Message, &inm.Conversation.ID, &inm.Contact.Msisdn, &inm.Message.ID, &inm.Message.Channelid, &inm, time.Now())
		res, err := tx.ExecContext(ctx, sqlQuery, &inm.Contact.ID, string(jsonDataContactJsonPlain), string(jsonDataConversationJsonPlain), string(jsonDataMessageJsonPlain), &inm.Conversation.ID, inm.Message.To, &inm.Message.ID, &inm.Message.Channelid, jsonDataAllJsonPlain, time.Now())

		if err != nil {
			tx.Rollback()
			errors.Message = structs.QueryErr
			errors.Data = string(jsonMsg)
			errors.SysMessage = err.Error()
			errors.Code = http.StatusInternalServerError
			fmt.Println(err.Error())
			return &errors
		}

		lastID, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			errors.Message = structs.LastIDErr
			errors.Data = string(jsonMsg)
			errors.SysMessage = err.Error()
			errors.Code = http.StatusInternalServerError
			fmt.Println(lastID)
			return &errors
		}

		errors.Message = structs.Success
		errors.Code = http.StatusOK

		tx.Commit()
	} else {
		errors.Message = structs.Success
		errors.Code = http.StatusOK
	}

	return &errors
}
func (inb *inboundRepo) GetWebhook(inboundChannel string, fromsender string, tosender string, trx_type string) ([]structs.WebHook, *structs.ErrorMessage) {

	var errors structs.ErrorMessage

	//should get existing token in sessions table
	//Get existing session
	ctx := context.Background()
	check, serv, err := GetWebhookByChannel(inboundChannel, fromsender, tosender, trx_type, inb.DB, ctx)
	if err != nil {
		errors.Message = structs.ErrNotFound
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return nil, &errors
	}
	if !check {
		errors.Message = structs.Success
		errors.Data = "success"
		errors.Code = http.StatusOK
		return serv, &errors
	} else {
		//True! send back the token
		errors.Message = structs.Success
		errors.Data = "success"
		errors.Code = http.StatusOK
		return serv, &errors
	}
}

func GetWebhookByChannel(inboundChannel string, frm string, to string, trx_type string, DB *sql.DB, ctx context.Context) (bool, []structs.WebHook, error) {
	var whs []structs.WebHook

	sCaseChannel := strings.ToLower(inboundChannel)
	if sCaseChannel == "whatsapp_sandbox" {
		sCaseChannel = "whatsapp"
	}

	var check bool

	db := DB
	queryWh := "select id, client_id, client_code, url, method, header_prefix, token, events, expected_http_code, retry, timeout "
	queryWh += "from ( "
	queryWh += "select * "
	queryWh += "FROM webhooks, json_table(events, '$[*]' columns (idx FOR ORDINALITY, channel_name text path '$.channel_name', frm text path '$.from', `to` text path '$.to')) asd "
	queryWh += "where status = 1 "
	queryWh += "and trx_type = ? "
	queryWh += ") dsa "
	queryWh += "where `to` = ? and lower(channel_name) = ?"

	//this query makes 2 times token access (each channel), in the future: please change this query into distinct vendor

	check = true

	res, err := db.QueryContext(ctx,queryWh, trx_type, to, sCaseChannel)
	defer mysql.CloseRows(res)

	if err != nil {
		fmt.Println(err.Error())
		return false, nil, err
	}

	for res.Next() {
		wh := structs.WebHook{}
		var eventsByte []byte
		res.Scan(&wh.ID, &wh.ClientID, &wh.ClientCode, &wh.URL, &wh.Method, &wh.HeaderPrefix, &wh.Token, &eventsByte, &wh.HttpCode, &wh.Retry, &wh.Timeout)
		json.Unmarshal(eventsByte, &wh.Events)
		whs = append(whs, wh)
	}

	return check, whs, nil

}

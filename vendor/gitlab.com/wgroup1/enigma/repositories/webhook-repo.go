package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"gitlab.com/wgroup1/enigma/common"
	"gitlab.com/wgroup1/enigma/database"
	"gitlab.com/wgroup1/enigma/interfaces"
	"gitlab.com/wgroup1/enigma/structs"
)

type webhookRepo struct{
	DB *sql.DB
}

func NewWebHookRepository(db *sql.DB) interfaces.WebHookRepository {
	return &webhookRepo{
		DB: db,
	}
}

func (whk *webhookRepo) GetURL(clientID int) []structs.WebHook {
	var webs []structs.WebHook
	var strJson []byte

	db := whk.DB
	ctx := context.Background()

	sqlQuery := "select id, client_id, client_code, url, method, header_prefix, token, events, expected_http_code, retry, timeout, status from webhooks where client_id = ? and status = 1"
	res, err := db.QueryContext(ctx, sqlQuery, clientID)
	defer database.CloseRows(res)

	if err != nil {
		log.Println(err.Error())
		return nil
	}

	web := structs.WebHook{}
	for res.Next() {
		res.Scan(&web.ID, &web.ClientID, &web.ClientCode, &web.URL, &web.Method, &web.HeaderPrefix, &web.Token, &strJson, &web.HttpCode, &web.Retry, &web.Timeout, &web.Status)
		json.Unmarshal(strJson, &web.Events)
		webs = append(webs, web)
	}
	return webs
}

func (whk *webhookRepo) GetWebHooks(ClientID int, page string, limit string) []structs.WebHookPlain {
	var webHooks []structs.WebHookPlain

	db := whk.DB
	ctx := context.Background()

	//DATE_FORMAT(start_date, '%Y-%m-%d %h:%m:%s') As start_date
	//SELECT id,client_id,client_code,url,method,header_prefix,token,events,expected_http_code,retry,timeout,status,created_at,created_by FROM webhooks
	//	sqlQuery := "select id, client_id, name, start_date, end_date, description, reward_codes, balance_ids, status, created_at,created_by, IFNULL(updated_at,'2001-01-01') As 'updated_at', IFNULL(updated_by,'') As 'updated_by' FROM redemption_config where client_id = ?"
	sqlQuery := "select id,client_id,client_code,url,method,header_prefix,token,events,expected_http_code,retry,timeout,status,created_at,created_by FROM webhooks where client_id = ? "
	sqlQuery += common.SetPageLimit(page, limit)
	//fmt.Println(sqlQuery)
	res, err := db.QueryContext(ctx, sqlQuery, ClientID)
	// res, err := db.Query(sqlQuery, ClientID)
	defer database.CloseRows(res)

	if err != nil {
		log.Println(err.Error())
		return nil
	}

	//var wh structs.WebHookPlain
	wh := structs.WebHookPlain{}
	for res.Next() {
		res.Scan(&wh.ID, &wh.ClientID, &wh.ClientCode, &wh.URL, &wh.Method, &wh.HeaderPrefix, &wh.Token, &wh.Events, &wh.HttpCode, &wh.Retry, &wh.Timeout, &wh.Status, &wh.CreatedAt, &wh.CreatedBy)
		webHooks = append(webHooks, wh)
	}
	return webHooks
}

func (whk *webhookRepo) GetWebHook(ClientID int, webHookId int) structs.WebHookPlain {
	var webHook structs.WebHookPlain
	db := whk.DB
	ctx := context.Background()

	sqlQuery := "select id,client_id,client_code,url,method,header_prefix,token,events,expected_http_code,retry,timeout,status,created_at,created_by FROM webhooks where client_id = ? and id = ?"
	err := db.QueryRowContext(ctx, sqlQuery, ClientID, webHookId).Scan(&webHook.ID, &webHook.ClientID, &webHook.ClientCode, &webHook.URL, &webHook.Method, &webHook.HeaderPrefix, &webHook.Token, &webHook.Events, &webHook.HttpCode, &webHook.Retry, &webHook.Timeout, &webHook.Status, &webHook.CreatedAt, &webHook.CreatedBy)


	if err != nil {
		log.Println(err.Error())
		return webHook
	}

	return webHook
}

func (whk *webhookRepo) AddWebHook(wh structs.WebHookObj, respStr string) *structs.ErrorMessage {
	var errors structs.ErrorMessage

	db := whk.DB
	ctx := context.Background()
	tx, err := db.Begin()

	//start checking insert
	if err != nil {
		//fmt.Println(err.Error())
		errors.Message = structs.QueryErr
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return &errors
	}
	eventsPlain, err1 := json.Marshal(wh.Events)
	if err1 != nil {
		eventsPlain = []byte(err1.Error())
	}

	var sqlQuery = ""
	sqlQuery += "INSERT INTO webhooks "
	sqlQuery += "( client_id,client_code,trx_type,url,method,header_prefix,token,events,expected_http_code,retry,timeout,status,created_at,created_by) "
	sqlQuery += "values(?,?,?,?,?,?,?,?,?,?,?,?,?,?); "

	res, err := tx.ExecContext(ctx, sqlQuery, &wh.ClientID, &wh.ClientCode, &wh.TrxType ,&wh.URL, &wh.Method, &wh.HeaderPrefix, &wh.Token, eventsPlain, &wh.HttpCode, &wh.Retry, &wh.Timeout, &wh.Status, time.Now(), &wh.CreatedBy)
	//, time.Now(), "")

	jsonDataWebHookJsonPlain, err2 := json.Marshal(wh)
	if err2 != nil {
		jsonDataWebHookJsonPlain = []byte(err2.Error())
	}

	if err != nil {
		tx.Rollback()
		errors.Message = structs.QueryErr
		errors.Data = string(jsonDataWebHookJsonPlain)
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		log.Println(err.Error())
		return &errors
	}

	if err != nil {
		tx.Rollback()
		errors.Message = structs.LastIDErr
		errors.Data = string(jsonDataWebHookJsonPlain)
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		//fmt.Println(lastID)
		return &errors
	}

	lastID, _ := res.LastInsertId()
	lastInsID := strconv.Itoa(int(lastID))

	errors.Message = structs.Success
	errors.Data = lastInsID
	errors.Code = http.StatusOK

	tx.Commit()
	return &errors
}

func (whk *webhookRepo) UpdateWebHook(wh structs.WebHookObj, respStr string) *structs.ErrorMessage {
	var errors structs.ErrorMessage

	db := whk.DB
	ctx := context.Background()
	tx, err := db.Begin()

	//start checking insert
	if err != nil {
		log.Println(err.Error())
		errors.Message = structs.QueryErr
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return &errors
	}
	jsonDataEventsJsonPlain, err1 := json.Marshal(wh.Events)
	if err1 != nil {
		jsonDataEventsJsonPlain = []byte(err1.Error())
	}
	var sqlQuery = ""
	//UPDATE products SET id=0, name='', `point`=0, status=1, created_at=CURRENT_TIMESTAMP, created_by='_utf8mb3\\''Admin\\''' WHERE product_code='' AND client_id=0;

	sqlQuery += "UPDATE webhooks "
	sqlQuery += "SET url = ? , "
	sqlQuery += "method = ? , "
	sqlQuery += "header_prefix = ? , "
	sqlQuery += "token = ? , "
	sqlQuery += "events = ? , "
	sqlQuery += "expected_http_code = ? , "
	sqlQuery += "retry = ? , "
	sqlQuery += "timeout = ? , "
	sqlQuery += "status = ? "
	sqlQuery += "WHERE id = ? "
	sqlQuery += "AND client_id = ? "

	//fmt.Println(sqlQuery)
	res, err := tx.ExecContext(ctx, sqlQuery, &wh.URL, &wh.Method, &wh.HeaderPrefix, &wh.Token, jsonDataEventsJsonPlain, &wh.HttpCode, &wh.Retry, &wh.Timeout, &wh.Status, &wh.ID, &wh.ClientID)

	jsonDataWebHookJsonPlain, err2 := json.Marshal(wh)
	if err2 != nil {
		jsonDataWebHookJsonPlain = []byte(err2.Error())
	}

	if err != nil {
		tx.Rollback()
		errors.Message = structs.QueryErr
		errors.Data = string(jsonDataWebHookJsonPlain)
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		log.Println(err.Error())
		return &errors
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		// defer db.Close()
		errors.Message = structs.LastIDErr
		errors.Data = string(jsonDataWebHookJsonPlain)
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		log.Println(lastID)
		return &errors
	}

	errors.Message = structs.Success
	errors.Code = http.StatusOK

	tx.Commit()
	return &errors
}

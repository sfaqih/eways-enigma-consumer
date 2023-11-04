package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"gitlab.com/wgroup1/enigma/interfaces"
	"gitlab.com/wgroup1/enigma/structs"
)

type inbWARepo struct{
	DB *sql.DB
}

func NewInboundWARepository(db *sql.DB) interfaces.InboundWARepository {
	return &inbWARepo{
		DB: db,
	}
}

func (ibwa *inbWARepo) CreateInboundWA(inb structs.Message, cli int) *structs.ErrorMessage {
	var errors structs.ErrorMessage
	var jsonLoc, jsonText, jsonImg, jsonDoc []byte
	db := ibwa.DB
	ctx := context.Background()
	tx, err := db.Begin()

	//start checking insert
	if err != nil {
		errors.Message = structs.QueryErr
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return &errors
	}

	sqlQuery := "INSERT INTO inbound_wa (client_id, type, timestamp_unix, timestamp_dt, text_json, location_json, image_json, document_json, messages) VALUES(?, ?, ?, FROM_UNIXTIME(?), ?, ?, ?, ?, ?)"

	if inb.Text != (structs.Text{}) {
		jsonText, _ = json.Marshal(inb.Text)
	}
	if inb.Location != (structs.Location{}) {
		jsonLoc, _ = json.Marshal(inb.Location)
	}
	if inb.Image != (structs.Image{}) {
		jsonImg, _ = json.Marshal(inb.Image)
	}
	if inb.Document != (structs.Document{}) {
		jsonDoc, _ = json.Marshal(inb.Document)
	}

	jsonMsg, _ := json.Marshal(inb)

	res, err := tx.ExecContext(ctx, sqlQuery, cli, &inb.Type, &inb.Timestamp, &inb.Timestamp, &jsonText, &jsonLoc, &jsonImg, &jsonDoc, &jsonMsg)

	if err != nil {
		log.Println(err.Error())
		tx.Rollback()
		errors.Message = structs.QueryErr
		errors.Data = string(jsonMsg)
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return &errors
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		errors.Message = structs.LastIDErr
		errors.Data = string(jsonMsg)
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		log.Println(lastID)
		return &errors
	}

	errors.Message = structs.Success
	errors.Code = http.StatusOK

	tx.Commit()
	// defer db.Close()
	return &errors
}

func (ibwa *inbWARepo) CreateInboundWADamcorp(iw []byte, vss []structs.WebHook) *structs.ErrorMessage {
	var errors structs.ErrorMessage

	//var planReq []byte

	db := ibwa.DB
	ctx := context.Background()

	tx, err := db.Begin()

	if err != nil {
		errors.Message = structs.QueryErr
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return &errors
	}

	//planReq, _ =  json.Marshal(iw)

	sqlQuery := "INSERT INTO inbound_dto (`from`, `to`, selected_code_menu, next_code_menu, json_plain) "
	sqlQuery += "VALUES(?, ?, ?, ?, ?);"

	//res, err1 := tx.Exec(sqlQuery, iw.ReqData.Contacts[0].WaId, "", iw.ReqData.Messages[0].Interactive.ListReply.Id, "", planReq)
	res, err1 := tx.ExecContext(ctx, sqlQuery, "", "", "", "", iw)

	if err1 != nil {
		log.Println(err1.Error())
		tx.Rollback()
		errors.Message = structs.QueryErr
		errors.Data = string(iw)
		errors.SysMessage = err1.Error()
		errors.Code = http.StatusInternalServerError
		return &errors
	}

	lastID, err2 := res.LastInsertId()
	if err2 != nil {
		tx.Rollback()
		errors.Message = structs.LastIDErr
		errors.Data = string(iw)
		errors.SysMessage = err2.Error()
		errors.Code = http.StatusInternalServerError
		log.Println(lastID)
		return &errors
	}

	errors.Message = structs.Success
	errors.Code = http.StatusOK

	tx.Commit()
	return &errors

}

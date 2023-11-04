package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"gitlab.com/wgroup1/enigma/interfaces"
	"gitlab.com/wgroup1/enigma/structs"
)

type flowRepo struct{
	DB *sql.DB
}

func NewFlowRepository(db *sql.DB) interfaces.FlowRepository {
	return &flowRepo{
		DB: db,
	}
}

func (flow *flowRepo) CreateFlow(fl structs.Flow) *structs.ErrorMessage {
	var errors structs.ErrorMessage
	//var typeJson []byte
	//var contentJson []byte

	db := flow.DB
	ctx := context.Background()
	tx, err := db.Begin()

	//start checking insert
	if err != nil {
		fmt.Println(err.Error())
		errors.Message = structs.QueryErr
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return &errors
	}

	sqlQuery := "INSERT INTO flow_log (client_id,outbound_json,status,created_at) values(?,?,?,?); "

	outboundbyte, _ := json.Marshal(fl.OutboundFlow)

	res, err := tx.ExecContext(ctx, sqlQuery, &fl.ClientID, outboundbyte, 1, time.Now())

	if err != nil {
		log.Println(err.Error())
		tx.Rollback()
		errors.Message = structs.QueryErr
		errors.Data = string(outboundbyte)
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return &errors
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		errors.Message = structs.LastIDErr
		errors.Data = string(outboundbyte)
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		//fmt.Println(lastID)
		return &errors
	}

	errors.Data = strconv.FormatInt(lastID, 10)
	errors.Message = structs.Success
	errors.Code = http.StatusOK

	tx.Commit()
	return &errors
}

func (*flowRepo) GetRunningFlow(sts structs.StatusReport) (structs.Flow, *structs.ErrorMessage) {
	var errors structs.ErrorMessage
	var fl structs.Flow

	return fl, &errors
}

func (*flowRepo) GetRunningFlowN8N(sts structs.StatusReport) (structs.Flow, *structs.ErrorMessage) {
	var errors structs.ErrorMessage
	var fl structs.Flow

	return fl, &errors
}

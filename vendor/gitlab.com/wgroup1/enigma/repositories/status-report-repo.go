package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"gitlab.com/wgroup1/enigma/database"
	"gitlab.com/wgroup1/enigma/interfaces"
	"gitlab.com/wgroup1/enigma/structs"
)

type statusreportRepo struct{
	DB *sql.DB
}

func NewStatusReportRepository(db *sql.DB) interfaces.StatusReportRepository {
	return &statusreportRepo{
		DB: db,
	}
}

func (srp *statusreportRepo) GetMessageByMessageId(sts structs.StatusReport) structs.StatusReport {
	var stsNew structs.StatusReport
	db := srp.DB
	ctx := context.Background()

	stsNew = sts
	strMessageID := sts.MessageID
	if sts.MjContactId != 0 {
		//strMessageID = strconv.FormatInt(sts.MessageID, 10)
		////sts.MessageId = strMessageID
		////sts.Status = sts.Event

		stsNew.MessageID = sts.MessageID
		//stsNew.MessageID = sts.MessageID

	}

	sqlQuery := "select id,flow_id, client_id,destination, channel, status, is_process FROM status_log where message_id = ? ORDER BY id DESC LIMIT 1 ;"

	err := db.QueryRowContext(ctx, sqlQuery, strMessageID).Scan(&stsNew.ID, &stsNew.Flowid, &stsNew.ClientID, &stsNew.Destination, &stsNew.Channel, &stsNew.Status, &stsNew.IsProcess)
	//	stsNew.Status = sts.Type


	if err != nil {
		log.Println(err.Error())
		return stsNew
	}
	return stsNew
}

func (srp *statusreportRepo) AddStatusLogReport(sts structs.StatusReport, conv_id int) *structs.ErrorMessage {
	var errors structs.ErrorMessage
	var stsNew structs.StatusReport

	db := srp.DB
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

	if sts.Channel == "Email" {
		sts.Destination = sts.Email
	}

	var sqlQueryGetFlowIdandMax = ""
	sqlQueryGetFlowIdandMax += "SELECT sl.flow_id,sl.client_id ,mx.max_orders ,sl.channel "
	sqlQueryGetFlowIdandMax += "FROM status_log sl "
	sqlQueryGetFlowIdandMax += "INNER JOIN  ( "
	sqlQueryGetFlowIdandMax += "SELECT od.id, MAX(od.orders) AS 'max_orders' "
	sqlQueryGetFlowIdandMax += "FROM (           SELECT id,mx1.orders "
	sqlQueryGetFlowIdandMax += "FROM flow_log, json_table(outbound_json, '$.messages[*]' columns (orders text path '$.order', idx FOR ORDINALITY)) mx1 "
	sqlQueryGetFlowIdandMax += ") od "
	sqlQueryGetFlowIdandMax += "GROUP BY od.id "
	sqlQueryGetFlowIdandMax += ") mx ON mx.id = sl.flow_id "
	sqlQueryGetFlowIdandMax += "where sl.status = 'start' AND  sl.destination = ? AND  sl.message_template = ? "
	sqlQueryGetFlowIdandMax += "ORDER BY sl.id DESC LIMIT 1 ; "

	res, err := db.QueryContext(ctx, sqlQueryGetFlowIdandMax, &sts.Destination, &sts.MessageTemplate)
	defer database.CloseRows(res)

	if err != nil {
		log.Println(err.Error())
		return nil
	}

	for res.Next() {
		res.Scan(&stsNew.Flowid, &stsNew.ClientID, &stsNew.MaxOrders, &stsNew.Channel)
		stsNew.Destination = sts.Destination
	}

	var sqlUpdate = ""
	sqlUpdate += "UPDATE status_log SET "
	sqlUpdate += "message_id = ? "
	sqlUpdate += ",max_orders = ? "
	sqlUpdate += ",last_updated_at = NOW() "
	sqlUpdate += ",is_process = 1 "
	sqlUpdate += "WHERE status = 'start' "
	//sqlUpdate += "AND ifNull(message_id,'') = '' "
	sqlUpdate += "AND channel = ? "
	sqlUpdate += "AND destination = ? "
	sqlUpdate += "ORDER BY id DESC LIMIT 1 ; "

	//fmt.Println(sqlUpdate)
	if stsNew.Channel != "" {
		_, errupdate := tx.ExecContext(ctx, sqlUpdate, &sts.MessageID, &stsNew.MaxOrders, &stsNew.Channel, &stsNew.Destination)
		if errupdate != nil {
			log.Println(errupdate.Error())
			tx.Rollback()
			errors.Message = structs.QueryErr
			errors.SysMessage = errupdate.Error()
			errors.Code = http.StatusInternalServerError
			log.Println(errupdate.Error())
			//return &errors
		}
	} else {
		_, errupdate := tx.ExecContext(ctx, sqlUpdate, &sts.MessageID, &stsNew.MaxOrders, &sts.Channel, &sts.Destination)
		if errupdate != nil {
			tx.Rollback()
			errors.Message = structs.QueryErr
			errors.SysMessage = errupdate.Error()
			errors.Code = http.StatusInternalServerError
			log.Println(errupdate.Error())
			//return &errors
		}
	}

	var sqlQuery = ""
	sqlQuery += "INSERT INTO status_log "
	sqlQuery += "(flow_id,client_id,channel,destination,message_id,message_template,status,status_json,transaction_at,is_process,max_orders) "
	sqlQuery += "VALUES (?,?,?,?,?,?,?,?,NOW(),?,?) "
	sqlQuery += "ON DUPLICATE KEY UPDATE "
	sqlQuery += "last_updated_at = NOW() ; "

	//var res sql.Result
	statusjsonbyte, _ := json.Marshal(sts)
	/*
		if sts.MjContactId != 0 {
			//strMessageID := strconv.FormatInt(sts.MessageID, 10)
			//sts.MessageId = strMessageID

			sts.Status = sts.Event
		}
	*/

	if stsNew.Flowid != 0 {
		_, err = tx.ExecContext(ctx, sqlQuery, &stsNew.Flowid, &stsNew.ClientID, &stsNew.Channel, &stsNew.Destination, &sts.MessageID, &sts.MessageTemplate, &sts.Status, statusjsonbyte, 0, &stsNew.MaxOrders)

	} else {
		_, err = tx.ExecContext(ctx, sqlQuery, &stsNew.Flowid, &stsNew.ClientID, &sts.Channel, &sts.Destination, &sts.MessageID, &sts.MessageTemplate, &sts.Status, statusjsonbyte, 0, &stsNew.MaxOrders)
	}

	if err != nil {
		log.Println(err.Error())
		tx.Rollback()
		errors.Message = structs.QueryErr
		errors.Data = string(statusjsonbyte)
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		log.Println(err.Error())
		return &errors
	}
	//}
	////iLastId,_ := sqlrslt.LastInsertId()

	tx.Commit()
	return nil
}

func (srp *statusreportRepo) AddStatusLogReportPlain(sts structs.StatusReport, conv_id int) *structs.ErrorMessage {
	var errors structs.ErrorMessage
	var stsNew structs.StatusReport

	db := srp.DB
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

	if sts.Channel == "Email" {
		sts.Destination = sts.Email
	}

	var sqlQuery = ""
	sqlQuery += "INSERT INTO status_log "
	sqlQuery += "(flow_id,client_id,channel,destination,message_id,message_template,status,status_json,transaction_at,is_process,max_orders) "
	sqlQuery += "VALUES (?,?,?,?,?,?,?,?,NOW(),?,?) "
	sqlQuery += "ON DUPLICATE KEY UPDATE "
	sqlQuery += "last_updated_at = NOW() ; "

	statusjsonbyte, _ := json.Marshal(sts)

	_, err = tx.ExecContext(ctx, sqlQuery, &stsNew.Flowid, &stsNew.ClientID, &sts.Channel, &sts.Destination, &sts.MessageID, &sts.MessageTemplate, &sts.Status, statusjsonbyte, 0, &stsNew.MaxOrders)

	if err != nil {
		tx.Rollback()
		errors.Message = structs.QueryErr
		errors.Data = string(statusjsonbyte)
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		log.Println(err.Error())
		return &errors
	}

	tx.Commit()
	return nil
}

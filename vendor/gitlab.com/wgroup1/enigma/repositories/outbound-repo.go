package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"


	"net/http"
	"strings"
	"time"

	"gitlab.com/wgroup1/enigma/common"
	"gitlab.com/wgroup1/enigma/database"
	"gitlab.com/wgroup1/enigma/interfaces"
	"gitlab.com/wgroup1/enigma/repositories/mysql"
	"gitlab.com/wgroup1/enigma/structs"
)

type outboundRepo struct{
	DB *sql.DB
}

func NewOutboundRepository(db *sql.DB) interfaces.OutboundRepository {
	return &outboundRepo{
		DB: db,
	}
}
func (oub *outboundRepo) UpdateStatusReceived(stats structs.Status, body []byte) (bool, error) {
	ctx := context.Background()
	_, errStatPending := deleteStatusPending(stats.ID, oub.DB, ctx)
	if errStatPending != nil {
		return false, errStatPending
	}
	return true, nil
}

func (oub *outboundRepo) GetMessageStatus() ([]structs.Status, *structs.ErrorMessage) {
	var stats []structs.Status
	var errors structs.ErrorMessage

	db := oub.DB
	ctx := context.Background()
	sqlQuery := "select id, client_id, sesame_message_id, sesame_id, contact_id, message_id, conversation_id, conversation_status, channel_id, platform, channel, direction, message_status, asd.from, asd.to, msisdn, contact_status, asd.type, request_body, message_json, contact_json, conversation_json, status FROM conv_status asd WHERE last_updated_at <= NOW()  order by last_updated_at asc LIMIT 100"

	res, err := db.QueryContext(ctx, sqlQuery)
	//fmt.Println(sqlQuery)
	defer database.CloseRows(res)

	if err != nil {
		errors.Message = structs.ErrNotFound
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		log.Println(err.Error())
		return nil, &errors
	}

	for res.Next() {
		stat := structs.Status{}
		res.Scan(&stat.ID, &stat.ClientID, &stat.SesameMessageID, &stat.SesameID, &stat.ContactID, &stat.MessageID, &stat.ConversationID, &stat.ConversationStatus, &stat.ChannelID, &stat.Platform, &stat.Channel, &stat.Direction, &stat.MsgStatus, &stat.From, &stat.To, &stat.MSISDN, &stat.ContactStatus, &stat.Type, &stat.ReqBody, &stat.MsgJson, &stat.ContactJson, &stat.ConvJson, &stat.Status)
		stats = append(stats, stat)
	}

	if len(stats) != 0 {
		return stats, nil
	} else {
		errors.Message = structs.ErrNotFound
		errors.SysMessage = ""
		errors.Code = http.StatusInternalServerError
		return nil, &errors
	}
}

func (oub *outboundRepo) CreateOutbound(outm structs.OutMessage, reqStr string, respStr string, cli int) *structs.ErrorMessage {
	var errors structs.ErrorMessage
	var typeJson []byte
	var contentJson []byte

	tx, err := oub.DB.Begin()
	ctx := context.Background()

	//start checking insert
	if err != nil {
		fmt.Println(err.Error())
		errors.Message = structs.QueryErr
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return &errors
	}

	sqlQuery := "INSERT INTO outbound (client_id, channel, channel_id, `from`, `to`, cc, bcc, wa_template_id, sesame_id, conversation_id, timestamp_unix, timestamp_dt, sesame_message_id, message_id, title, content, `type`, type_json, request_json, response_json,message_plain,trx_id) values(?,?,?,?,?,?,?,?,?,?,?,FROM_UNIXTIME(?),?,?,?,?,?,?,?,?,?,?)"

	switch outm.Type {
	case "hsm":
		typeJson, _ = json.Marshal(outm.Hsm)
	case "text":
		contentJson, _ = json.Marshal(outm.Content)
	case "image":
		contentJson, _ = json.Marshal(outm.Content)
		outm.Title = outm.Content.Image.Caption
	case "video":
		contentJson, _ = json.Marshal(outm.Content)
		outm.Title = outm.Content.Video.Caption
	case "audio":
		contentJson, _ = json.Marshal(outm.Content)
		outm.Title = outm.Content.Audio.Caption
	}

	jsonMsg, _ := json.Marshal(outm)

	res, err := tx.ExecContext(ctx, sqlQuery, cli, &outm.Channel, &outm.ChannelID, &outm.From, &outm.To, &outm.Cc, &outm.Bcc, &outm.WATemplateID, &outm.SesameID, &outm.ConversationID, &outm.Timestamp, &outm.Timestamp, &outm.SesameMessageID, &outm.MessageID, &outm.Title, &contentJson, &outm.Type, &typeJson, &reqStr, &respStr, jsonMsg, &outm.TrxId)

	if err != nil {
		fmt.Println(err.Error())
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
		fmt.Println(lastID)
		return &errors
	}

	errors.Message = structs.Success
	errors.Code = http.StatusOK

	tx.Commit()
	return &errors
}

func (oub *outboundRepo) CreateOutboundBulk(outm []*structs.HTTPRequest, cli int) *structs.ErrorMessage {
	var errors structs.ErrorMessage
	var valueStrings []string

	vals := []interface{}{}

	db := oub.DB
	tx, err := db.Begin()
	ctx := context.Background()

	//start checking insert
	if err != nil {
		log.Println(err.Error())
		errors.Message = structs.QueryErr
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return &errors
	}

	sqlQuery := "INSERT INTO outbound (client_id, channel, channel_id, `from`, `to`, cc, bcc, wa_template_id, sesame_id, conversation_id, timestamp_unix, timestamp_dt, sesame_message_id, message_id, title, content, `type`, type_json, request_json, response_json,message_plain,trx_id) values "

	iMaxInsert := 200
	iCnt := 0

	for j, msg := range outm {
		var typeJson []byte
		var contentJson []byte
		var outMessage structs.OutMessage
		valueStrings = append(valueStrings, "(?,?,?,?,?,?,?,?,?,?,?,FROM_UNIXTIME(?),?,?,?,?,?,?,?,?,?,?)")
		json.Unmarshal(msg.Outbound, &outMessage)

		switch outMessage.Type {
		case "hsm":
			typeJson, _ = json.Marshal(outMessage.Hsm)
		case "text":
			contentJson, _ = json.Marshal(outMessage.Content)
		case "image":
			contentJson, _ = json.Marshal(outMessage.Content)
			outMessage.Title = outMessage.Content.Image.Caption
		case "video":
			contentJson, _ = json.Marshal(outMessage.Content)
			outMessage.Title = outMessage.Content.Video.Caption
		case "audio":
			contentJson, _ = json.Marshal(outMessage.Content)
			outMessage.Title = outMessage.Content.Audio.Caption
		}

		jsonMsg, _ := json.Marshal(outMessage)

		vals = append(vals, cli, &outMessage.Channel, &outMessage.ChannelID, &outMessage.From, &outMessage.To, &outMessage.Cc, &outMessage.Bcc, &outMessage.WATemplateID, &outMessage.SesameID, &outMessage.ConversationID, &outMessage.Timestamp, &outMessage.Timestamp, &outMessage.SesameMessageID, &outMessage.MessageID, &outMessage.Title, &contentJson, &outMessage.Type, &typeJson, &msg.RequestBody, &msg.ResponseBody, jsonMsg, &outMessage.TrxId)
		iCnt++

		if iMaxInsert == iCnt || j + 1 == len(outm){
			queryVals := strings.Join(valueStrings, ",")  
			sqlQueryBulk := sqlQuery + queryVals
	
			stmt, errstmt := tx.PrepareContext(ctx, sqlQueryBulk)
			if errstmt != nil {
				log.Println(err.Error())
				continue
			}
	
			_, err = stmt.ExecContext(ctx, vals...)
			if err != nil {
				//tx.Rollback()
				log.Println(err.Error())
				continue
			}
			vals = []interface{}{}
			iCnt = 1

		}

	}

	errors.Message = structs.Success
	errors.Code = http.StatusOK
	tx.Commit()
	return &errors
}

func (oub *outboundRepo) UpdateStatusFailed(stats structs.Status, body []byte) (bool, error) {
	//var cnt int
	db := oub.DB
	ctx := context.Background()
	//Count the null token based on existing services that client used

	/*
		cntQry := "select count(*) from status_failed ss2 where message_id = ?"
		err := db.QueryRow(cntQry, stats.MessageID).Scan(&cnt)
		if err != nil {
			fmt.Println(err.Error())
			return false, err
		}


		if cnt > 0 {
			rspCleansing, errStatPendingCleansing := deleteStatusPendingByMessageId(stats.MessageID)
			if errStatPendingCleansing != nil {
				return false, errStatPendingCleansing
			}

			if rspCleansing == 0 {
				rssCleansing, errStatSentCleansing := deleteStatusSentByMessageId(stats.MessageID)
				if errStatSentCleansing != nil {
					return false, errStatSentCleansing
				}

				if rssCleansing == 0 {
					rsrCleansing, errStatDeliveredCleansing := deleteStatusDeliveredByMessageId(stats.MessageID)
					if errStatDeliveredCleansing != nil {
						return false, errStatDeliveredCleansing
					}
					if rsrCleansing == 0 {
						_, errStatReadCleansing := deleteStatusReadByMessageId(stats.MessageID)
						if errStatReadCleansing != nil {
							return false, errStatReadCleansing
						}
					}
				}
			}
			return true, nil
		}
	*/

	//insert
	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err.Error())
		return false, err
	}

	sqlQuery := "INSERT INTO status_failed (api_log_id, response_body, contact_id, message_id, conversation_id, conversation_status, channel_id, platform, channel, direction, message_status, `from`, `to`, msisdn, contact_status, `type`, request_body, message_json, contact_json, conversation_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	res, err := tx.ExecContext(ctx, sqlQuery, &stats.ApiID, string(body), &stats.ContactID, &stats.MessageID, &stats.ConversationID, &stats.ConversationStatus, &stats.ChannelID, &stats.Platform, &stats.Channel, &stats.Direction, &stats.MsgStatus, &stats.From, &stats.To, &stats.MSISDN, &stats.ContactStatus, &stats.Type, &stats.ReqBody, &stats.MsgJson, &stats.ContactJson, &stats.ConvJson)
	if err != nil {
		tx.Rollback()
		return false, err
	}

	lastID, err := res.LastInsertId()
	lastInsID := int(lastID)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	log.Println("Success insert, Last ID in status_failed:", lastInsID)
	tx.Commit()

	_, _ = deleteStatusPendingByMessageId(stats.MessageID, oub.DB, ctx)
	_, _ = deleteStatusSentByMessageId(stats.MessageID, oub.DB, ctx)
	_, _ = deleteStatusDeliveredByMessageId(stats.MessageID, oub.DB, ctx)
	return true, nil
}

func (oub *outboundRepo) UpdateStatusRejected(stats structs.Status, body []byte) (bool, error) {
	//var cnt int
	db := oub.DB
	ctx := context.Background()

	//Count the null token based on existing services that client used
	/*
		cntQry := "select count(*) from status_rejected ss where message_id = ?"
		err := db.QueryRow(cntQry, stats.MessageID).Scan(&cnt)
		if err != nil {
			fmt.Println(err.Error())
			return false, err
		}

		if cnt > 0 {
			rspCleansing, errStatPendingCleansing := deleteStatusPendingByMessageId(stats.MessageID)
			if errStatPendingCleansing != nil {
				return false, errStatPendingCleansing
			}

			if rspCleansing == 0 {
				rssCleansing, errStatSentCleansing := deleteStatusSentByMessageId(stats.MessageID)
				if errStatSentCleansing != nil {
					return false, errStatSentCleansing
				}

				if rssCleansing == 0 {
					rsrCleansing, errStatDeliveredCleansing := deleteStatusDeliveredByMessageId(stats.MessageID)
					if errStatDeliveredCleansing != nil {
						return false, errStatDeliveredCleansing
					}
					if rsrCleansing == 0 {
						_, errStatReadCleansing := deleteStatusReadByMessageId(stats.MessageID)
						if errStatReadCleansing != nil {
							return false, errStatReadCleansing
						}
					}
				}
			}
			return true, nil
		}
	*/

	//insert
	tx, err := db.Begin()
	if err != nil {
		log.Println(err.Error())
		return false, err
	}

	sqlQuery := "INSERT INTO status_rejected (api_log_id, response_body, contact_id, message_id, conversation_id, conversation_status, channel_id, platform, channel, direction, message_status, `from`, `to`, msisdn, contact_status, `type`, request_body, message_json, contact_json, conversation_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	res, err := tx.ExecContext(ctx, sqlQuery, &stats.ApiID, string(body), &stats.ContactID, &stats.MessageID, &stats.ConversationID, &stats.ConversationStatus, &stats.ChannelID, &stats.Platform, &stats.Channel, &stats.Direction, &stats.MsgStatus, &stats.From, &stats.To, &stats.MSISDN, &stats.ContactStatus, &stats.Type, &stats.ReqBody, &stats.MsgJson, &stats.ContactJson, &stats.ConvJson)
	if err != nil {
		tx.Rollback()
		log.Println(err.Error())
		// defer db.Close()
		return false, err
	}

	lastID, err := res.LastInsertId()
	lastInsID := int(lastID)
	if err != nil {
		tx.Rollback()
		// defer db.Close()
		return false, err
	}
	log.Println("Success insert, Last ID in status_rejected:", lastInsID)
	tx.Commit()

	_, _ = deleteStatusPendingByMessageId(stats.MessageID, oub.DB, ctx)
	_, _ = deleteStatusSentByMessageId(stats.MessageID, oub.DB, ctx)
	_, _ = deleteStatusDeliveredByMessageId(stats.MessageID, oub.DB, ctx)
	// defer db.Close()
	return true, nil

}

func (oub *outboundRepo) UpdateStatusPending(stats structs.Status, body []byte) (bool, error) {
	var cnt int
	db := oub.DB
	ctx := context.Background()
	//Count the null token based on existing services that client used
	cntQry := "select count(*) from status_pending ss where message_id = ?"
	err := db.QueryRowContext(ctx, cntQry, stats.MessageID).Scan(&cnt)
	// defer db.Close()

	if err != nil {
		log.Println(err.Error())
		return false, err
	}

	if cnt > 0 {
		//update last_updated_at supaya ada jeda 10 menit sebelum di process lagi
		tx, err := db.Begin()
		if err != nil {
			log.Println(err.Error())
			return false, err
		}
		sqlQuery := "UPDATE status_pending SET last_updated_at = DATE_ADD(NOW(), INTERVAL 10 MINUTE),response_body = ?, api_log_id = ? WHERE message_id = ?"

		//fmt.Println(sqlQuery)
		res, err := tx.ExecContext(ctx, sqlQuery, string(body), &stats.ApiID, stats.MessageID)
		_ = res
		if err != nil {
			tx.Rollback()
			// defer db.Close()
			log.Println(err.Error())
			return false, err
		}
		tx.Commit()
		// defer db.Close()
		return true, nil
	}

	iexist, _ := validateDeliveryPositionByMessageId(stats.MessageID, "pending")
	if iexist == 0 {
		//insert
		tx, err := db.Begin()
		if err != nil {
			tx.Rollback()
			// defer db.Close()
			log.Println(err.Error())
			return false, err
		}

		sqlQuery := "INSERT INTO status_pending (api_log_id, response_body, contact_id, message_id, conversation_id, conversation_status, channel_id, platform, channel, direction, message_status, `from`, `to`, msisdn, contact_status, `type`, request_body, message_json, contact_json, conversation_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
		res, err := tx.ExecContext(ctx, sqlQuery, &stats.ApiID, string(body), &stats.ContactID, &stats.MessageID, &stats.ConversationID, &stats.ConversationStatus, &stats.ChannelID, &stats.Platform, &stats.Channel, &stats.Direction, &stats.MsgStatus, &stats.From, &stats.To, &stats.MSISDN, &stats.ContactStatus, &stats.Type, &stats.ReqBody, &stats.MsgJson, &stats.ContactJson, &stats.ConvJson)
		if err != nil {
			tx.Rollback()
			// defer db.Close()
			log.Println(err.Error())
			return false, err
		}

		lastID, err := res.LastInsertId()
		lastInsID := int(lastID)
		if err != nil {
			tx.Rollback()
			// defer db.Close()
			return false, err
		}
		log.Println("Success insert, Last ID in status_pending:", lastInsID)
		tx.Commit()
	}
	// defer db.Close()
	return true, nil

}

func (oub *outboundRepo) UpdateStatusSent(stats structs.Status, body []byte) (bool, error) {
	var cnt int
	db := oub.DB
	ctx := context.Background()
	//Count the null token based on existing services that client used
	cntQry := "select count(*) from status_sent ss2 where message_id = ?"
	err := db.QueryRowContext(ctx, cntQry, stats.MessageID).Scan(&cnt)
	// defer db.Close()

	if err != nil {
		log.Println(err.Error())
		return false, err
	}

	if cnt > 0 {
		//update last_updated_at supaya ada jeda 10 menit sebelum di process lagi
		tx, err := db.Begin()
		if err != nil {
			log.Println(err.Error())
			return false, err
		}
		sqlQuery := " UPDATE status_sent SET last_updated_at = DATE_ADD(NOW(), INTERVAL 10 MINUTE) WHERE message_id = ?"
		// sqlQuery := "INSERT INTO status_sent (api_log_id, response_body, contact_id, message_id, conversation_id, conversation_status, channel_id, platform, channel, direction, message_status, `from`, `to`, msisdn, contact_status, `type`, request_body, message_json, contact_json, conversation_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
		//fmt.Println(sqlQuery)
		res, err := tx.ExecContext(ctx, sqlQuery, stats.MessageID, stats.MessageID)
		_ = res
		if err != nil {
			tx.Rollback()
			// defer db.Close()
			log.Println(err.Error())
			return false, err
		}
		tx.Commit()
		_, _ = deleteStatusPendingByMessageId(stats.MessageID, oub.DB, ctx)
		// defer db.Close()
		return true, nil
	}

	iexist, _ := validateDeliveryPositionByMessageId(stats.MessageID, "sent")
	if iexist == 0 {
		//insert
		tx, err := db.Begin()
		if err != nil {
			tx.Rollback()
			// defer db.Close()
			return false, err
		}

		sqlQuery := "INSERT INTO status_sent (api_log_id, response_body, contact_id, message_id, conversation_id, conversation_status, channel_id, platform, channel, direction, message_status, `from`, `to`, msisdn, contact_status, `type`, request_body, message_json, contact_json, conversation_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
		res, err := tx.ExecContext(ctx, sqlQuery, &stats.ApiID, string(body), &stats.ContactID, &stats.MessageID, &stats.ConversationID, &stats.ConversationStatus, &stats.ChannelID, &stats.Platform, &stats.Channel, &stats.Direction, &stats.MsgStatus, &stats.From, &stats.To, &stats.MSISDN, &stats.ContactStatus, &stats.Type, &stats.ReqBody, &stats.MsgJson, &stats.ContactJson, &stats.ConvJson)
		if err != nil {
			tx.Rollback()
			// defer db.Close()
			log.Println(err.Error())
			return false, err
		}

		lastID, err := res.LastInsertId()
		lastInsID := int(lastID)
		if err != nil {
			tx.Rollback()
			// defer db.Close()
			return false, err
		}
		log.Println("Success insert, Last ID in status_sent:", lastInsID)
		tx.Commit()
	}
	_, _ = deleteStatusPendingByMessageId(stats.MessageID, oub.DB, ctx)
	// defer db.Close()
	return true, nil

}

func (oub *outboundRepo) UpdateStatusDelivered(stats structs.Status, body []byte) (bool, error) {
	var cnt int
	db := oub.DB
	ctx := context.Background()
	//Count the null token based on existing services that client used
	cntQry := "select count(*) from status_delivered ss2 where message_id = ?"
	err := db.QueryRowContext(ctx, cntQry, stats.MessageID).Scan(&cnt)
	if err != nil {
		log.Println(err.Error())
		return false, err
	}

	if cnt > 0 {
		//update last_updated_at supaya ada jeda 10 menit sebelum di process lagi
		tx, err := db.Begin()
		if err != nil {
			log.Println(err.Error())
			return false, err
		}
		sqlQuery := "UPDATE status_delivered SET last_updated_at = DATE_ADD(NOW(), INTERVAL 10 MINUTE) WHERE message_id = ?"
		// sqlQuery := "INSERT INTO status_sent (api_log_id, response_body, contact_id, message_id, conversation_id, conversation_status, channel_id, platform, channel, direction, message_status, `from`, `to`, msisdn, contact_status, `type`, request_body, message_json, contact_json, conversation_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
		//fmt.Println(sqlQuery)
		res, err := tx.ExecContext(ctx, sqlQuery, stats.MessageID)
		_ = res
		if err != nil {
			tx.Rollback()
			// defer db.Close()
			log.Println(err.Error())
			return false, err
		}
		tx.Commit()

		_, _ = deleteStatusPendingByMessageId(stats.MessageID, oub.DB, ctx)
		_, _ = deleteStatusSentByMessageId(stats.MessageID, oub.DB, ctx)
		/*
			rspCleansing, errStatPendingCleansing := deleteStatusPendingByMessageId(stats.MessageID)
			if errStatPendingCleansing != nil {
				return false, errStatPendingCleansing
			}

			if rspCleansing == 0 {
				_, errStatSentCleansing := deleteStatusSentByMessageId(stats.MessageID)
				if errStatSentCleansing != nil {
					return false, errStatSentCleansing
				}

			}
		*/
		// defer db.Close()
		return true, nil
	}

	//insert
	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		// defer db.Close()
		return false, err
	}
	iexist, _ := validateDeliveryPositionByMessageId(stats.MessageID, "delivered")
	if iexist == 0 {
		sqlQuery := "INSERT INTO status_delivered (api_log_id, response_body, contact_id, message_id, conversation_id, conversation_status, channel_id, platform, channel, direction, message_status, `from`, `to`, msisdn, contact_status, `type`, request_body, message_json, contact_json, conversation_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
		res, err := tx.ExecContext(ctx, sqlQuery, &stats.ApiID, string(body), &stats.ContactID, &stats.MessageID, &stats.ConversationID, &stats.ConversationStatus, &stats.ChannelID, &stats.Platform, &stats.Channel, &stats.Direction, &stats.MsgStatus, &stats.From, &stats.To, &stats.MSISDN, &stats.ContactStatus, &stats.Type, &stats.ReqBody, &stats.MsgJson, &stats.ContactJson, &stats.ConvJson)
		if err != nil {
			tx.Rollback()
			// defer db.Close()
			log.Println(err.Error())
			return false, err
		}

		lastID, err := res.LastInsertId()
		lastInsID := int(lastID)
		if err != nil {
			tx.Rollback()
			// defer db.Close()
			return false, err
		}
		log.Println("Success insert, Last ID in status_delivered:", lastInsID)
		tx.Commit()
	}

	_, _ = deleteStatusPendingByMessageId(stats.MessageID, oub.DB, ctx)
	_, _ = deleteStatusSentByMessageId(stats.MessageID, oub.DB, ctx)

	/*
		rsp, errStatPending := deleteStatusPending(stats.ID)
		if errStatPending != nil {
			return false, errStatPending
		}


		if rsp == 0 {
			_, errStatSent := deleteStatusSent(stats.ID)
			if errStatSent != nil {
				return false, errStatSent
			}
		}
	*/
	// defer db.Close()
	return true, nil
}

func (oub *outboundRepo) UpdateStatusRead(stats structs.Status, body []byte) (bool, error) {

	db := oub.DB
	ctx := context.Background()
	/*
		var cnt int

		//Count the null token based on existing services that client used
		//cntQry := "select count(*) from status_read ss2 where message_id = ?"
		cntQry := "select count(*) from status_read ss2 where message_id = ?"
		//fmt.Println(cntQry)
		err := db.QueryRow(cntQry, stats.MessageID).Scan(&cnt)
		if err != nil {
			fmt.Println(err.Error())
			return false, err
		}

		if cnt > 0 {
			rspCleansing, errStatPendingCleansing := deleteStatusPendingByMessageId(stats.MessageID)
			if errStatPendingCleansing != nil {
				return false, errStatPendingCleansing
			}

			if rspCleansing == 0 {
				rssCleansing, errStatSentCleansing := deleteStatusSentByMessageId(stats.MessageID)
				if errStatSentCleansing != nil {
					return false, errStatSentCleansing
				}

				if rssCleansing == 0 {
					_, errStatDeliveredCleansing := deleteStatusDeliveredByMessageId(stats.MessageID)
					if errStatDeliveredCleansing != nil {
						return false, errStatDeliveredCleansing
					}
				}
			}
			return true, nil
		}
	*/
	var cnt int

	//Count the null token based on existing services that client used
	//cntQry := "select count(*) from status_read ss2 where message_id = ?"
	cntQry := "select count(*) from status_read ss2 where message_id = ?"
	//fmt.Println(cntQry)
	err := db.QueryRowContext(ctx, cntQry, stats.MessageID).Scan(&cnt)
	if err != nil {
		log.Println(err.Error())
		return false, err
	}

	if cnt == 0 {

		//insert
		tx, err := db.Begin()
		if err != nil {
			log.Println(err.Error())
			// defer db.Close()
			return false, err
		}

		sqlQuery := "INSERT INTO status_read (api_log_id, response_body, contact_id, message_id, conversation_id, conversation_status, channel_id, platform, channel, direction, message_status, `from`, `to`, msisdn, contact_status, `type`, request_body, message_json, contact_json, conversation_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
		res, err := tx.ExecContext(ctx, sqlQuery, &stats.ApiID, string(body), &stats.ContactID, &stats.MessageID, &stats.ConversationID, &stats.ConversationStatus, &stats.ChannelID, &stats.Platform, &stats.Channel, &stats.Direction, &stats.MsgStatus, &stats.From, &stats.To, &stats.MSISDN, &stats.ContactStatus, &stats.Type, &stats.ReqBody, &stats.MsgJson, &stats.ContactJson, &stats.ConvJson)
		if err != nil {
			tx.Rollback()
			// defer db.Close()
			log.Println(err.Error())
			return false, err
		}

		lastID, err := res.LastInsertId()
		lastInsID := int(lastID)
		if err != nil {
			tx.Rollback()
			// defer db.Close()
			return false, err
		}
		fmt.Println("Success insert, Last ID in status_read:", lastInsID)
		tx.Commit()
	}
	_, _ = deleteStatusPendingByMessageId(stats.MessageID, oub.DB, ctx)
	_, _ = deleteStatusSentByMessageId(stats.MessageID, oub.DB, ctx)
	_, _ = deleteStatusDeliveredByMessageId(stats.MessageID, oub.DB, ctx)

	/*
		rsp, errStatPending := deleteStatusPending(stats.ID)
		if errStatPending != nil {
			return false, errStatPending
		}

		if rsp == 0 {
			rss, errStatSent := deleteStatusSent(stats.ID)
			if errStatSent != nil {
				return false, errStatSent
			}

			if rss == 0 {
				_, errStatDelivered := deleteStatusDelivered(stats.ID)
				if errStatDelivered != nil {
					return false, errStatDelivered
				}
			}
		}
	*/
	// defer db.Close()
	return true, nil
}

// delete status_pending
func deleteStatusPending(ID int64, DB *sql.DB, ctx context.Context) (int, error) {
	db := DB
	txDel, err := db.Begin()
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}

	sdp := "delete from status_pending where id = ?"
	rdp, err := txDel.ExecContext(ctx, sdp, ID)
	if err != nil {
		txDel.Rollback()
		return 0, err
	}

	radp, err := rdp.RowsAffected()
	radps := int(radp)
	if err != nil {
		txDel.Rollback()
		return 0, err
	}
	log.Println("Delete status_pending rows affected:", radps)
	txDel.Commit()

	return radps, nil
}

func deleteStatusPendingByMessageId(messageid string, DB *sql.DB, ctx context.Context) (int, error) {
	db := DB
	txDel, err := db.Begin()
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}

	sdp := "delete from status_pending where message_id = ?"
	rdp, err := txDel.ExecContext(ctx, sdp, messageid)
	if err != nil {
		txDel.Rollback()
		return 0, err
	}

	radp, err := rdp.RowsAffected()
	radps := int(radp)
	if err != nil {
		txDel.Rollback()
		return 0, err
	}
	log.Println("Delete status_pending rows affected:", radps)
	txDel.Commit()

	return radps, nil
}

//delete status_sent
// func deleteStatusSent(ID int64) (int, error) {
// 	db := mysql.InitializeMySQL()
// 	txDel, err := db.Begin()
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return 0, err
// 	}

// 	sdp := "delete from status_sent where id = ?"
// 	rdp, err := txDel.Exec(sdp, ID)
// 	if err != nil {
// 		txDel.Rollback()
// 		return 0, err
// 	}

// 	radp, err := rdp.RowsAffected()
// 	radps := int(radp)
// 	if err != nil {
// 		txDel.Rollback()
// 		return 0, err
// 	}
// 	fmt.Println("Delete status_sent rows affected:", radps)
// 	txDel.Commit()

// 	return radps, nil
// }

func deleteStatusSentByMessageId(messageid string, DB *sql.DB, ctx context.Context) (int, error) {
	db := DB
	txDel, err := db.Begin()
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}

	sdp := "delete from status_sent where message_id = ?"
	rdp, err := txDel.ExecContext(ctx, sdp, messageid)
	if err != nil {
		txDel.Rollback()
		return 0, err
	}

	radp, err := rdp.RowsAffected()
	radps := int(radp)
	if err != nil {
		txDel.Rollback()
		return 0, err
	}
	log.Println("Delete status_sent rows affected:", radps)
	txDel.Commit()

	return radps, nil
}

//delete status_delivered
// func deleteStatusDelivered(ID int64) (int, error) {
// 	db := mysql.InitializeMySQL()
// 	txDel, err := db.Begin()
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return 0, err
// 	}

// 	sdp := "delete from status_delivered where id = ?"
// 	rdp, err := txDel.Exec(sdp, ID)
// 	if err != nil {
// 		txDel.Rollback()
// 		return 0, err
// 	}

// 	radp, err := rdp.RowsAffected()
// 	radps := int(radp)
// 	if err != nil {
// 		txDel.Rollback()
// 		return 0, err
// 	}
// 	fmt.Println("Delete status_delivered rows affected:", radps)
// 	txDel.Commit()

// 	return radps, nil
// }

func deleteStatusDeliveredByMessageId(messageid string, DB *sql.DB, ctx context.Context) (int, error) {
	db := DB
	txDel, err := db.Begin()
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}

	sdp := "delete from status_delivered where message_id = ?"
	rdp, err := txDel.ExecContext(ctx, sdp, messageid)
	if err != nil {
		txDel.Rollback()
		return 0, err
	}

	radp, err := rdp.RowsAffected()
	radps := int(radp)
	if err != nil {
		txDel.Rollback()
		return 0, err
	}
	log.Println("Delete status_delivered rows affected:", radps)
	txDel.Commit()

	return radps, nil
}

// func deleteStatusReadByMessageId(messageid string) (int, error) {
// 	db := mysql.InitializeMySQL()
// 	txDel, err := db.Begin()
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return 0, err
// 	}

// 	sdp := "delete from status_read where message_id = ?"
// 	rdp, err := txDel.Exec(sdp, messageid)
// 	if err != nil {
// 		txDel.Rollback()
// 		return 0, err
// 	}

// 	radp, err := rdp.RowsAffected()
// 	radps := int(radp)
// 	if err != nil {
// 		txDel.Rollback()
// 		return 0, err
// 	}
// 	fmt.Println("Delete status_read rows affected:", radps)
// 	txDel.Commit()

// 	return radps, nil
// }

func validateDeliveryPositionByMessageId(messageid string, mbstatus string) (int, error) {
	db := mysql.InitializeMySQL()

	istatus := 0
	/*
	 * status_pending = 0
	 * status_sent = 1
	 * status_delivered = 2
	 */

	switch mbstatus {
	case "pending":
		istatus = 0
	case "sent":
		istatus = 1
	case "delivered":
		istatus = 2
	default:
		istatus = -1
	}

	fmt.Println(istatus)
	var cnt int
	cntQry := "SELECT COUNT(1) FROM conv_status where message_id = ? AND status > ?"
	err := db.QueryRow(cntQry, messageid, istatus).Scan(&cnt)
	if err != nil {
		fmt.Println(err.Error())
		return -1, err
	}

	return cnt, nil
}

func getSessionForStatus(vendoralias string, channel string, DB *sql.DB) (bool, []structs.VendorService, error) {
	var cnt int
	var vss []structs.VendorService
	var qry string
	var check bool

	vs := structs.VendorService{}
	db := DB
	ctx := context.Background()
	//Count the null token based on existing services that client used
	cntQry := "select count(*) from (select cv.vendor_alias, cv.vendor_id, cv.channel, v2.url, vs.uri, vs.`method`, vs.header_prefix, s.token, s.expired_at, s.status as session_status from client_vendors cv left join vendor_services vs on cv.channel = vs.service_name and cv.vendor_alias = vs.vendor_alias and cv.vendor_id = vs.vendor_id and cv.channel = vs.service_name and header_prefix != '' left join sessions s on s.vendor_id = cv.vendor_id and s.client_id = cv.client_id and s.vendor_alias = vs.vendor_alias and s.status = 1 AND s.expired_at > (now() + interval 10 minute) left join vendor v2 on v2.id = cv.vendor_id where cv.vendor_alias = ? AND channel = ? and cv.status = 1 and vs.status = 1) asd where token is null"
	err := db.QueryRowContext(ctx, cntQry, vendoralias, channel).Scan(&cnt)
	// defer db.Close()

	if err != nil {
		fmt.Println(err.Error())
		return false, nil, err
	}

	getChannel := "select cv.client_id, cv.vendor_alias, cv.vendor_id, cv.channel, v2.url, v2.uri_login, v2.username, v2.password, v2.`method` as 'login_method', v2.auth_type as 'login_auth_type', v2.header_prefix as 'login_header_prefix', vs.uri, vs.method, vs.header_prefix, s.token, s.expired_at, s.status as session_status from client_vendors cv  left join vendor_services vs on cv.channel = vs.service_name and cv.vendor_alias = vs.vendor_alias and cv.vendor_id = vs.vendor_id and cv.status = 1 and vs.status = 1 left join sessions s on s.client_id = cv.client_id and s.vendor_alias = cv.vendor_alias and s.vendor_id = cv.vendor_id and s.status = 1 AND s.expired_at > (now() + interval 10 minute) left join vendor v2 on v2.id = cv.vendor_id where cv.vendor_alias = ? AND cv.channel = ? "
	//this query makes 2 times token access (each channel), in the future: please change this query into distinct vendor
	getExpiredChannel := "select cv.client_id, cv.vendor_alias, cv.vendor_id, cv.channel, v2.url, v2.uri_login, v2.username, v2.password, v2.`method` as 'login_method', v2.auth_type as 'login_auth_type', v2.header_prefix as 'login_header_prefix', vs.uri, vs.method, vs.header_prefix, s.token, s.expired_at, s.status as session_status from client_vendors cv  left join vendor_services vs on cv.channel = vs.service_name and cv.vendor_alias = vs.vendor_alias and cv.vendor_id = vs.vendor_id and cv.status = 1 and vs.status = 1 left join sessions s on s.client_id = cv.client_id and s.vendor_alias = cv.vendor_alias and s.vendor_id = cv.vendor_id and s.status = 1 AND s.expired_at > (now() + interval 10 minute) left join vendor v2 on v2.id = cv.vendor_id where vs.vendor_alias = ? AND cv.channel = ? and s.token is null"

	if cnt > 0 {
		qry = getExpiredChannel
		check = false
	} else {
		qry = getChannel
		check = true
	}

	res, err := db.Query(qry, vendoralias, channel)
	defer database.CloseRows(res)
	// defer db.Close()

	if err != nil {
		log.Println(err.Error())
		return false, nil, err
	}

	for res.Next() {
		res.Scan(&vs.ClientID, &vs.Alias, &vs.ID, &vs.Channel, &vs.URL, &vs.URILogin, &vs.Username, &vs.Password, &vs.LoginMethod, &vs.LoginAuthType, &vs.LoginHeaderPrefix, &vs.URI, &vs.Method, &vs.HeaderPrefix, &vs.Token, &vs.ExpiredAt, &vs.Status)
		vss = append(vss, vs)
	}

	return check, vss, nil
}

func getSessions(clientID int, DB *sql.DB, ctx context.Context) (bool, []structs.VendorService, error) {
	var cnt int
	var vss []structs.VendorService
	var qry string
	var check bool

	db := DB
	//Count the null token based on existing services that client used
	cntQry := "select count(*) from (select cv.vendor_alias, cv.vendor_id, cv.channel, v2.url, vs.uri, vs.`method`, vs.header_prefix, s.token, s.expired_at, s.status as session_status from client_vendors cv left join vendor_services vs on cv.channel = vs.service_name and cv.vendor_alias = vs.vendor_alias and cv.vendor_id = vs.vendor_id and cv.channel = vs.service_name left join sessions s on s.vendor_id = cv.vendor_id and s.client_id = cv.client_id and s.vendor_alias = vs.vendor_alias and s.status = 1 AND s.expired_at > (now() + interval 10 minute) left join vendor v2 on v2.id = cv.vendor_id where cv.status = 1 and vs.status = 1 and cv.client_id = ?) asd where token is null"
	err := db.QueryRowContext(ctx, cntQry, clientID).Scan(&cnt)

	if err != nil {
		log.Println(err.Error())
		return false, nil, err
	}

	getChannel := "select cv.vendor_alias, cv.vendor_id, cv.channel, v2.url, v2.uri_login, v2.username, v2.password, v2.`method` as 'login_method', v2.auth_type as 'login_auth_type', v2.header_prefix as 'login_header_prefix', vs.uri, vs.method, vs.header_prefix, s.token, s.expired_at, s.status as session_status, vs.from_sender, vs.reply_to from client_vendors cv  left join vendor_services vs on cv.channel = vs.service_name and cv.vendor_alias = vs.vendor_alias and cv.vendor_id = vs.vendor_id and cv.status = 1 and vs.status = 1 left join sessions s on s.client_id = cv.client_id and s.vendor_alias = cv.vendor_alias and s.vendor_id = cv.vendor_id and s.status = 1 AND s.expired_at > (now() + interval 10 minute) left join vendor v2 on v2.id = cv.vendor_id where cv.client_id = ? AND cv.client_id = ? ;"
	//this query makes 2 times token access (each channel), in the future: please change this query into distinct vendor
	getExpiredChannel := "select cv.vendor_alias, cv.vendor_id, cv.channel, v2.url, v2.uri_login, v2.username, v2.password, v2.`method` as 'login_method', v2.auth_type as 'login_auth_type', v2.header_prefix as 'login_header_prefix', vs.uri, vs.method, vs.header_prefix, s.token, s.expired_at, s.status as session_status, vs.from_sender, vs.reply_to from client_vendors cv  left join vendor_services vs on cv.channel = vs.service_name and cv.vendor_alias = vs.vendor_alias and cv.vendor_id = vs.vendor_id and cv.status = 1 and vs.status = 1 and cv.client_id = ? left join sessions s on s.client_id = cv.client_id and s.vendor_alias = cv.vendor_alias and s.vendor_id = cv.vendor_id and s.status = 1 AND s.expired_at > (now() + interval 10 minute) left join vendor v2 on v2.id = cv.vendor_id where s.token is null AND cv.client_id = ?;"

	if cnt > 0 {
		qry = getExpiredChannel
		check = false
	} else {
		qry = getChannel
		check = true
	}

	res, err := db.QueryContext(ctx, qry, clientID, clientID)
	defer database.CloseRows(res)

	if err != nil {
		log.Println(err.Error())
		return false, nil, err
	}

	for res.Next() {
		vs := structs.VendorService{}
		res.Scan(&vs.Alias, &vs.ID, &vs.Channel, &vs.URL, &vs.URILogin, &vs.Username, &vs.Password, &vs.LoginMethod, &vs.LoginAuthType, &vs.LoginHeaderPrefix, &vs.URI, &vs.Method, &vs.HeaderPrefix, &vs.Token, &vs.ExpiredAt, &vs.Status, &vs.FromSender, &vs.ReplyTo)
		vss = append(vss, vs)
	}

	return check, vss, nil

}

func (oub *outboundRepo) GetTokenStatus(vendoralias string, channel string) ([]structs.VendorService, *structs.ErrorMessage) {
	var vend structs.VendorLogin
	var auth structs.AuthVendor
	var token, exp string
	var errors structs.ErrorMessage

	//should get existing token in sessions table
	//Get existing session
	check, serv, err := getSessionForStatus(vendoralias, channel, oub.DB)
	if err != nil {
		errors.Message = structs.ErrNotFound
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return nil, &errors
	}

	if !check {
		//false! then request token
		for j := range serv {
			strtoken := serv[j].LoginAuthType + " " + common.BasicAuth(serv[j].Username, serv[j].Password)
			_, resp, body, _, err := common.HitAPI(serv[j].URL+serv[j].URILogin, "", serv[j].LoginMethod, strtoken, "", time.Duration(0), oub.DB)
			if err != nil {
				errors.Message = structs.ErrNotFound
				errors.SysMessage = err.Error()
				errors.Code = http.StatusInternalServerError
				return nil, &errors
			}

			if resp.StatusCode == 200 {
				json.Unmarshal(body, &auth)
				for j := range auth.Users {
					fmt.Println(auth.Users[j].Token)
					token = auth.Users[j].Token
					exp = auth.Users[j].ExpiredAt
				}

				db := oub.DB
				ctx := context.Background()
				tx, err := db.Begin()
				if err != nil {
					log.Println(err.Error())
					errors.Message = structs.DBErr
					errors.Data = vend.Alias
					errors.SysMessage = err.Error()
					errors.Code = http.StatusInternalServerError
					return nil, &errors
				}

				sqlQuery := "insert into sessions (client_id, vendor_id, vendor_alias, token, expired_at) values (?, ?, ?, ?, ?)"
				res, err := tx.ExecContext(ctx, sqlQuery, &serv[j].ClientID, &serv[j].ID, &serv[j].Alias, token, exp)
				if err != nil {
					tx.Rollback()
					log.Println(err.Error())
					// defer db.Close()
					errors.Message = structs.QueryErr
					errors.Data = vend.Alias
					errors.SysMessage = err.Error()
					errors.Code = http.StatusInternalServerError
					return nil, &errors
				}

				lastID, err := res.LastInsertId()
				lastInsID := int(lastID)
				if err != nil {
					tx.Rollback()
					// defer db.Close()
					errors.Message = structs.LastIDErr
					errors.Data = fmt.Sprint(lastInsID)
					errors.SysMessage = err.Error()
					errors.Code = http.StatusInternalServerError
					return nil, &errors
				}
				tx.Commit()

			} else {
				body, _ := ioutil.ReadAll(resp.Body)
				log.Println("something wrong with the vendor's API:", string(body), resp.Status)
				errors.Message = structs.LastIDErr
				errors.Data = string(body)
				errors.SysMessage = err.Error()
				errors.Code = http.StatusInternalServerError
				return nil, &errors
			}
		}

		_, servFix, err := getSessionForStatus(vendoralias, channel, oub.DB)
		if err != nil {
			errors.Message = "error when get sessions after hit to vendor"
			errors.Data = ""
			errors.SysMessage = err.Error()
			errors.Code = http.StatusInternalServerError
			return nil, &errors
		}
		errors.Message = structs.Success
		errors.Data = "success"
		errors.Code = http.StatusOK
		return servFix, &errors
	} else {
		//True! send back the token
		errors.Message = structs.Success
		errors.Data = "success"
		errors.Code = http.StatusOK
		return serv, &errors
	}

}

// GetToken is the func to get the token (first time or not)
func (oub *outboundRepo) GetToken(clientID int) ([]structs.VendorService, *structs.ErrorMessage) {
	var vend structs.VendorLogin
	var auth structs.AuthVendor
	var authToken structs.Damcorp_Auth_Response
	var token, exp string
	var errors structs.ErrorMessage

	ctx := context.Background()

	//should get existing token in sessions table
	//Get existing session
	check, serv, err := getSessions(clientID, oub.DB, ctx)
	if err != nil {
		errors.Message = structs.ErrNotFound
		errors.SysMessage = err.Error()
		errors.Code = http.StatusInternalServerError
		return nil, &errors
	}

	if !check {
		//false! then request token
		for j := range serv {
			strtoken := ""
			bodyReq := ""
			bodyStr := make(map[string]interface{})
			strtoken = serv[j].LoginAuthType + " " + common.BasicAuth(serv[j].Username, serv[j].Password)
			if serv[j].Alias == "DAMCORP" {
				strtoken = ""
				bodyStr["email"] = serv[j].Username
				bodyStr["password"] = serv[j].Password
				jsonMsg, _ := json.Marshal(bodyStr)

				bodyReq = string(jsonMsg)
			}

			_, resp, body, _, err := common.HitAPI(serv[j].URL+serv[j].URILogin, bodyReq, serv[j].LoginMethod, strtoken, "", time.Duration(0), oub.DB)
			if err != nil {
				errors.Message = structs.ErrNotFound
				errors.SysMessage = err.Error()
				errors.Code = http.StatusInternalServerError
				return nil, &errors
			}

			if resp.StatusCode == 200 {
				if serv[j].Alias == "DAMCORP" {
					loc, _ := time.LoadLocation("Asia/Jakarta")
					json.Unmarshal(body, &authToken)
					token = authToken.Data.Token
					exp = time.Now().In(loc).Add(time.Hour * 12).Format("2006-01-02 15:04:05")
					// exp = time.Now().Local().Add(time.Hour * 12).Format("2006-01-02 15:04:05")
				}
				json.Unmarshal(body, &auth)
				for j := range auth.Users {
					//fmt.Println(auth.Users[j].Token)
					token = auth.Users[j].Token
					exp = auth.Users[j].ExpiredAt
				}

				db := oub.DB
				tx, err := db.Begin()
				if err != nil {
					log.Println(err.Error())
					errors.Message = structs.DBErr
					errors.Data = vend.Alias
					errors.SysMessage = err.Error()
					errors.Code = http.StatusInternalServerError
					return nil, &errors
				}

				sqlQuery := "insert into sessions (client_id, vendor_id, vendor_alias, token, expired_at) values (?, ?, ?, ?, ?)"
				res, err := tx.ExecContext(ctx, sqlQuery, clientID, &serv[j].ID, &serv[j].Alias, token, exp)
				if err != nil {
					tx.Rollback()
					errors.Message = structs.QueryErr
					errors.Data = vend.Alias
					errors.SysMessage = err.Error()
					errors.Code = http.StatusInternalServerError
					return nil, &errors
				}

				lastID, err := res.LastInsertId()
				lastInsID := int(lastID)
				if err != nil {
					tx.Rollback()
					errors.Message = structs.LastIDErr
					errors.Data = fmt.Sprint(lastInsID)
					errors.SysMessage = err.Error()
					errors.Code = http.StatusInternalServerError
					return nil, &errors
				}
				tx.Commit()

			} else {
				body, _ := ioutil.ReadAll(resp.Body)
				log.Println("something wrong with the vendor's API:", serv[j].Alias, "#", serv[j].URL, string(body), resp.Status)
				errors.Message = structs.LastIDErr
				errors.Data = string(body)
				errors.SysMessage = resp.Status
				errors.Code = http.StatusInternalServerError
				//return nil, &errors
			}
		}
		_, servFix, err := getSessions(clientID, oub.DB, ctx)
		if err != nil {
			errors.Message = "error when get sessions after hit to vendor"
			errors.Data = ""
			errors.SysMessage = err.Error()
			errors.Code = http.StatusInternalServerError
			return nil, &errors
		}
		errors.Message = structs.Success
		errors.Data = "success"
		errors.Code = http.StatusOK
		return servFix, &errors
	} else {
		//True! send back the token
		errors.Message = structs.Success
		errors.Data = "success"
		errors.Code = http.StatusOK
		return serv, &errors
	}
}

func (oub *outboundRepo) CreateOutboundBulkLog(outb structs.OutboundBulkLog) (*structs.OutboundBulkLog, error) {

	db := oub.DB
	tx, err := db.Begin()
	// defer db.Close()

	//start checking insert
	if err != nil {
		log.Println("error using transaction with error : ", err.Error())
		return &outb, err
	}

	sqlQuery := "INSERT INTO outbound_bulk_request_log (`client_id`,`request_body`,`status`, `created_at`) VALUES(?,?,?,NOW())"

	res, err := tx.Exec(sqlQuery, &outb.ClientId, &outb.RequestBody, &outb.Status)

	if err != nil {
		tx.Rollback()
		log.Println("error in repo when insert to database with error : ", err.Error())
		return &outb, err
	}

	lastID, _ := res.LastInsertId()
	lastInsID := int(lastID)

	tx.Commit()

	outb.Id = lastInsID

	return &outb, nil
}
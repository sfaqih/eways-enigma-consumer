package repositories

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"

	"gitlab.com/wgroup1/enigma/interfaces"
	"gitlab.com/wgroup1/enigma/structs"
)

type logRepo struct{
	DB *sql.DB
}

func NewLogRepository(db *sql.DB) interfaces.LogRepository {
	return &logRepo{
		DB: db,
	}
}

func (lg *logRepo) InsertLogBulk(l []*structs.HTTPRequest) *structs.ErrorMessage {
	var errs structs.ErrorMessage
	var valueStrings []string

	vals := []interface{}{}

	db := lg.DB
	ctx := context.Background()
	tx, err := db.Begin()

	//start checking insert
	if err != nil {
		log.Println("error in begin transaction : ", err.Error())
		errs.Code = http.StatusInternalServerError
		errs.SysMessage = "Error connection to database  : " + err.Error()
		errs.Message = structs.QueryErr
	}

	sqlQuery := "insert into api_logs (url, method, request_header, request_body, response_status, response_header, response_body) values "

	iMaxInsert := 200
	iCnt := 0

	for i, lg := range l {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?)")
		vals = append(vals, &lg.Url, &lg.Method, &lg.RequestHeader, &lg.RequestBody, &lg.ResponseStatus, &lg.RequestHeader, &lg.ResponseBody)



		iCnt++

		if iCnt == iMaxInsert || i + 1 == len(l) {
			queryVals := strings.Join(valueStrings, ",")  
			sqlQueryBulk := sqlQuery + queryVals
	
			stmt, errstmt := tx.Prepare(sqlQueryBulk)
			if errstmt != nil {
				// errs.Code = http.StatusInternalServerError
				// errs.SysMessage = "Error if prepare sql transaction : " + err.Error()
				// errs.Message = structs.QueryErr
				log.Println(err.Error())
				continue
			}
	
			_, err = stmt.ExecContext(ctx, vals...)
			if err != nil {
				// errs.Code = http.StatusInternalServerError
				// errs.SysMessage = "Error if execute bulk insert : " + err.Error()
				// errs.Message = structs.QueryErr
				log.Println(err.Error())
				continue
			}
			// vals = []interface{}{}
			iCnt = 1
		}
	}

	errs.Message = structs.Success
	errs.Code = http.StatusOK
	tx.Commit()
	return &errs
}
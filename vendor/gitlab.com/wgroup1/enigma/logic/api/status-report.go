package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/wgroup1/enigma/common"
	"gitlab.com/wgroup1/enigma/repositories/redis"
	service "gitlab.com/wgroup1/enigma/services"
	"gitlab.com/wgroup1/enigma/structs"
)

type statusReportLogic struct{
	statusReportService service.StatusReportService
	inbService          service.InboundService
	DB *sql.DB
}


type StatusReportLogic interface {
	SaveStatusReport(w http.ResponseWriter, r *http.Request)
	SaveStatusReportPlain(w http.ResponseWriter, r *http.Request)
	SaveStatusSparkPost(w http.ResponseWriter, r *http.Request)

	SaveStatusSparkPostConsumer(w http.ResponseWriter, r *http.Request)
	ProcessStatusSparkPostConsumer(statusreportsparkposts []structs.StatusReportSparkPost) bool
}

func NewStatusReportLogic(service service.StatusReportService, inboundService service.InboundService, db *sql.DB) StatusReportLogic {
	return &statusReportLogic{
		statusReportService: service,
		inbService: inboundService,
		DB: db,
	}
}

func (ssp *statusReportLogic) SaveStatusReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var errs []structs.ErrorMessage
	var stsSparkPostAll []structs.StatusReportSparkPost
	_ = json.NewDecoder(r.Body).Decode(&stsSparkPostAll)

	errsave := ssp.statusReportService.SaveStatusReport(stsSparkPostAll)
	errs = append(errs, *errsave)

	common.JSONErrs(w, &errs)
}

func (ssp *statusReportLogic) SaveStatusReportPlain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var errs []structs.ErrorMessage
	var stsSparkPostAll []structs.StatusReportSparkPost
	_ = json.NewDecoder(r.Body).Decode(&stsSparkPostAll)

	errsave := ssp.statusReportService.SaveStatusReportPlain(stsSparkPostAll)
	errs = append(errs, *errsave)

	common.JSONErrs(w, &errs)
}

func (ssp *statusReportLogic) SaveStatusSparkPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	var errs []structs.ErrorMessage
	var stsSparkPostAll []structs.StatusReportSparkPost
	_ = json.NewDecoder(r.Body).Decode(&stsSparkPostAll)

	errsave := ssp.statusReportService.SaveStatusReportPlain(stsSparkPostAll)
	errs = append(errs, *errsave)

	common.JSONErrs(w, &errs)
}

func (ssp *statusReportLogic) SaveStatusSparkPostConsumer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("status", "200")

	var err structs.ErrorMessageSparkPost
	var errs []structs.ErrorMessageSparkPost
	var stsSparkPostAll []structs.StatusReportSparkPost
	_ = json.NewDecoder(r.Body).Decode(&stsSparkPostAll)

	stsSparkPostAllByte, _ := json.Marshal(stsSparkPostAll)
	redis.RPush(common.SPARKPOST_QUEUE_REPORT, stsSparkPostAllByte, common.REDIS_DB)

	err.Status = 200
	err.Data = string(stsSparkPostAllByte)
	err.Message = "success queue"
	//err.RespMessage = string(stsSparkPostAllByte)

	errs = append(errs, err)

	common.JSONErrsSparkPost(w, &errs)

}

func (ssp *statusReportLogic) ProcessStatusSparkPostConsumer(statusreportsparkposts []structs.StatusReportSparkPost) bool {

	rtnbool := true
	startUpdateStatusGenerate := time.Now()
	errsave := ssp.statusReportService.SaveStatusReportSparkPost(statusreportsparkposts)
	if errsave.Code != 200 {
		rtnbool = false
	}

	/*
		webhooks, err := inbService.GetWebhook("SparkPostStatus", "", "", "")
		_ = err
		statusreportsparkpostsByte, _ := json.Marshal(statusreportsparkposts)

		for i := range webhooks {
			_, _, body, _, errhit := common.HitAPI(webhooks[i].URL, string(statusreportsparkpostsByte), webhooks[i].Method, webhooks[i].Token, "", time.Duration(0))
			_ = body
			_ = errhit
		}
	*/
	webhooks, err := ssp.inbService.GetWebhook("SparkPostStatus", "enigma", "galactic", "SparkPostStatus")
	_ = err
	statusreportsparkpostsByte, _ := json.Marshal(statusreportsparkposts)

	for i := range webhooks {
		_, _, body, _, errhit := common.HitAPI(webhooks[i].URL, string(statusreportsparkpostsByte), webhooks[i].Method, webhooks[i].Token, "", time.Duration(0), ssp.DB)
		_ = body
		if errhit != nil {
			fmt.Println(errhit)
		}

	}

	elapsedUpdateStatusGenerate := time.Since(startUpdateStatusGenerate)
	fmt.Println("Process StatusSparkPost Consumer Completed Took :", elapsedUpdateStatusGenerate)

	return rtnbool
}

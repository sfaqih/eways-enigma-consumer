package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/wgroup1/enigma/interfaces"
	"gitlab.com/wgroup1/enigma/structs"
)

type StatusReportService interface {
	SaveStatusReport(stsSparkPostAll []structs.StatusReportSparkPost) *structs.ErrorMessage
	SaveStatusReportPlain(stsSparkPostAll []structs.StatusReportSparkPost) *structs.ErrorMessage
	SaveStatusReportSparkPost(stsSparkPostAll []structs.StatusReportSparkPost) *structs.ErrorMessage
}

type statusReportService struct{}

var (
	statusReportRepo interfaces.StatusReportRepository
)

func NewStatusReportService(repository interfaces.StatusReportRepository) StatusReportService {
	statusReportRepo = repository
	return &statusReportService{}
}

func (*statusReportService) SaveStatusReport(stsSparkPostAll []structs.StatusReportSparkPost) *structs.ErrorMessage {
	var errors structs.ErrorMessage
	var err *structs.ErrorMessage

	sts := structs.StatusReport{}
	stsnew := structs.StatusReport{}

	for _, stsSparkPost := range stsSparkPostAll {
		if (structs.TrackEvent{} == stsSparkPost.Msys.TrackEvent) {
			fmt.Println("Message Event")
			sts.Channel = "Email"
			sts.Email = stsSparkPost.Msys.MessageEvent.RcptTo
			sts.Destination = stsSparkPost.Msys.MessageEvent.RcptTo
			sts.Status = stsSparkPost.Msys.MessageEvent.Type

			sts.Event = stsSparkPost.Msys.MessageEvent.EventID
			tstp, _ := strconv.ParseInt(stsSparkPost.Msys.MessageEvent.Timestamp, 10, 64)
			sts.Timestamp = tstp
			sts.MessageID = stsSparkPost.Msys.MessageEvent.MessageID
			sts.Event = stsSparkPost.Msys.MessageEvent.EventID
			sts.Email = stsSparkPost.Msys.MessageEvent.RcptTo
			var iCpgInt64 int64
			if stsSparkPost.Msys.MessageEvent.TemplateID != "" {
				iCpgInt64, _ = strconv.ParseInt(stsSparkPost.Msys.MessageEvent.TemplateID, 10, 64)
			} else {
				iCpgInt64, _ = strconv.ParseInt("0", 10, 64)
			}
			sts.MjCampaignId = iCpgInt64

			var iCtcInt64 int64
			if stsSparkPost.Msys.MessageEvent.CustomerID != "" {
				iCtcInt64, _ = strconv.ParseInt(stsSparkPost.Msys.MessageEvent.CustomerID, 10, 64)
			} else {
				iCtcInt64, _ = strconv.ParseInt("0", 10, 64)
			}
			sts.MjContactId = iCtcInt64
			sts.CustomCampaign = stsSparkPost.Msys.MessageEvent.TemplateID
			sts.SmtpReply = stsSparkPost.Msys.MessageEvent.MsgFrom
			byte_plain, err_plain := json.Marshal(stsSparkPost)
			_ = err_plain
			sts.StatusJson = string(byte_plain)
		} else if (structs.MessageEvent{} == stsSparkPost.Msys.MessageEvent) {
			fmt.Println("Track Event")
			sts.Channel = "Email"
			sts.Email = stsSparkPost.Msys.TrackEvent.RcptTo
			sts.Destination = stsSparkPost.Msys.TrackEvent.RcptTo
			sts.Status = stsSparkPost.Msys.TrackEvent.Type

			sts.Event = stsSparkPost.Msys.TrackEvent.EventID
			tstp, _ := strconv.ParseInt(stsSparkPost.Msys.TrackEvent.Timestamp, 10, 64)
			sts.Timestamp = tstp
			sts.MessageID = stsSparkPost.Msys.TrackEvent.MessageID
			sts.Event = stsSparkPost.Msys.TrackEvent.EventID
			sts.Email = stsSparkPost.Msys.TrackEvent.RcptTo
			var iCpgInt64 int64
			if stsSparkPost.Msys.TrackEvent.TemplateID != "" {
				iCpgInt64, _ = strconv.ParseInt(stsSparkPost.Msys.TrackEvent.TemplateID, 10, 64)
			} else {
				iCpgInt64, _ = strconv.ParseInt("0", 10, 64)
			}
			sts.MjCampaignId = iCpgInt64

			var iCtcInt64 int64
			if stsSparkPost.Msys.TrackEvent.CustomerID != "" {
				iCtcInt64, _ = strconv.ParseInt(stsSparkPost.Msys.TrackEvent.CustomerID, 10, 64)
			} else {
				iCtcInt64, _ = strconv.ParseInt("0", 10, 64)
			}
			sts.MjContactId = iCtcInt64
			sts.CustomCampaign = stsSparkPost.Msys.TrackEvent.TemplateID
			sts.SmtpReply = stsSparkPost.Msys.TrackEvent.MsgFrom
			byte_plain, err_plain := json.Marshal(stsSparkPost)
			_ = err_plain
			sts.StatusJson = string(byte_plain)
		} else {
			fmt.Println("Unknown Event")
			stsSparkPostAllByte, _ := json.Marshal(stsSparkPostAll)
			errors.Message = "Unknown Event"
			errors.RespMessage = "Unknown Event"
			errors.ReqMessage = string(stsSparkPostAllByte)

			return &errors
		}

		stsnew = statusReportRepo.GetMessageByMessageId(sts)
		stsnew.Type = sts.Status

		err = statusReportRepo.AddStatusLogReport(stsnew, 0)

		if err != nil {
			errors.Message = err.Message
			errors.RespMessage = err.RespMessage
			errors.ReqMessage = err.Data
			errors.Code = err.Code
			return &errors

		}
	}

	//statusjsonbyte, _ := json.Marshal(sts)
	errors.Message = structs.Success
	//errors.ReqMessage = string(statusjsonbyte)
	errors.Code = http.StatusOK

	return &errors
}

func (*statusReportService) SaveStatusReportPlain(stsSparkPostAll []structs.StatusReportSparkPost) *structs.ErrorMessage {

	var errors structs.ErrorMessage
	var err *structs.ErrorMessage

	sts := structs.StatusReport{}
	stsnew := structs.StatusReport{}

	for _, stsSparkPost := range stsSparkPostAll {
		if (structs.TrackEvent{} == stsSparkPost.Msys.TrackEvent) {
			//fmt.Println("Message Event")
			sts.Channel = "Email"
			sts.Email = stsSparkPost.Msys.MessageEvent.RcptTo
			sts.Destination = stsSparkPost.Msys.MessageEvent.RcptTo
			sts.Status = stsSparkPost.Msys.MessageEvent.Type

			sts.Event = stsSparkPost.Msys.MessageEvent.EventID
			tstp, _ := strconv.ParseInt(stsSparkPost.Msys.MessageEvent.Timestamp, 10, 64)
			sts.Timestamp = tstp
			sts.MessageID = stsSparkPost.Msys.MessageEvent.MessageID
			sts.Event = stsSparkPost.Msys.MessageEvent.EventID
			sts.Email = stsSparkPost.Msys.MessageEvent.RcptTo
			var iCpgInt64 int64
			if stsSparkPost.Msys.MessageEvent.TemplateID != "" {
				iCpgInt64, _ = strconv.ParseInt(stsSparkPost.Msys.MessageEvent.TemplateID, 10, 64)
			} else {
				iCpgInt64, _ = strconv.ParseInt("0", 10, 64)
			}
			sts.MjCampaignId = iCpgInt64

			var iCtcInt64 int64
			if stsSparkPost.Msys.MessageEvent.CustomerID != "" {
				iCtcInt64, _ = strconv.ParseInt(stsSparkPost.Msys.MessageEvent.CustomerID, 10, 64)
			} else {
				iCtcInt64, _ = strconv.ParseInt("0", 10, 64)
			}
			sts.MjContactId = iCtcInt64
			sts.CustomCampaign = stsSparkPost.Msys.MessageEvent.TemplateID
			sts.SmtpReply = stsSparkPost.Msys.MessageEvent.MsgFrom
			byte_plain, err_plain := json.Marshal(stsSparkPost)
			_ = err_plain
			sts.StatusJson = string(byte_plain)
		} else if (structs.MessageEvent{} == stsSparkPost.Msys.MessageEvent) {
			//fmt.Println("Track Event")
			sts.Channel = "Email"
			sts.Email = stsSparkPost.Msys.TrackEvent.RcptTo
			sts.Destination = stsSparkPost.Msys.TrackEvent.RcptTo
			sts.Status = stsSparkPost.Msys.TrackEvent.Type

			sts.Event = stsSparkPost.Msys.TrackEvent.EventID
			tstp, _ := strconv.ParseInt(stsSparkPost.Msys.TrackEvent.Timestamp, 10, 64)
			sts.Timestamp = tstp
			sts.MessageID = stsSparkPost.Msys.TrackEvent.MessageID
			sts.Event = stsSparkPost.Msys.TrackEvent.EventID
			sts.Email = stsSparkPost.Msys.TrackEvent.RcptTo
			var iCpgInt64 int64
			if stsSparkPost.Msys.TrackEvent.TemplateID != "" {
				iCpgInt64, _ = strconv.ParseInt(stsSparkPost.Msys.TrackEvent.TemplateID, 10, 64)
			} else {
				iCpgInt64, _ = strconv.ParseInt("0", 10, 64)
			}
			sts.MjCampaignId = iCpgInt64

			var iCtcInt64 int64
			if stsSparkPost.Msys.TrackEvent.CustomerID != "" {
				iCtcInt64, _ = strconv.ParseInt(stsSparkPost.Msys.TrackEvent.CustomerID, 10, 64)
			} else {
				iCtcInt64, _ = strconv.ParseInt("0", 10, 64)
			}
			sts.MjContactId = iCtcInt64
			sts.CustomCampaign = stsSparkPost.Msys.TrackEvent.TemplateID
			sts.SmtpReply = stsSparkPost.Msys.TrackEvent.MsgFrom
			byte_plain, err_plain := json.Marshal(stsSparkPost)
			_ = err_plain
			sts.StatusJson = string(byte_plain)
		} else {
			//fmt.Println("Unknown Event")
			byte_plain, err_plain := json.Marshal(stsSparkPost)
			_ = err_plain
			sts.StatusJson = string(byte_plain)
		}

		stsnew.Channel = sts.Channel
		stsnew.Email = sts.Email
		stsnew.Destination = sts.Destination
		stsnew.Status = sts.Status

		stsnew.Event = sts.Event
		stsnew.Timestamp = sts.Timestamp
		stsnew.MessageID = sts.MessageID
		stsnew.Event = sts.Event
		stsnew.Email = sts.Email
		stsnew.MjCampaignId = sts.MjCampaignId

		stsnew.MjContactId = sts.MjContactId
		stsnew.CustomCampaign = sts.CustomCampaign
		stsnew.SmtpReply = sts.SmtpReply
		stsnew.StatusJson = sts.StatusJson

		err = statusReportRepo.AddStatusLogReportPlain(stsnew, 0)

		if err != nil {
			errors.Message = err.Message
			errors.RespMessage = err.RespMessage
			errors.ReqMessage = err.Data
			errors.Code = err.Code
			return &errors

		}
	}

	//statusjsonbyte, _ := json.Marshal(sts)
	errors.Message = structs.Success
	//errors.ReqMessage = string(statusjsonbyte)
	errors.Code = http.StatusOK

	return &errors
}

func (*statusReportService) SaveStatusReportSparkPost(stsSparkPostAll []structs.StatusReportSparkPost) *structs.ErrorMessage {

	var errors structs.ErrorMessage
	var err *structs.ErrorMessage

	sts := structs.StatusReport{}
	stsnew := structs.StatusReport{}

	for _, stsSparkPost := range stsSparkPostAll {
		if (structs.TrackEvent{} == stsSparkPost.Msys.TrackEvent) {
			//fmt.Println("Message Event")
			sts.Channel = "Email"
			sts.Email = stsSparkPost.Msys.MessageEvent.RcptTo
			sts.Destination = stsSparkPost.Msys.MessageEvent.RcptTo
			sts.Status = stsSparkPost.Msys.MessageEvent.Type

			sts.Event = stsSparkPost.Msys.MessageEvent.EventID
			tstp, _ := strconv.ParseInt(stsSparkPost.Msys.MessageEvent.Timestamp, 10, 64)
			sts.Timestamp = tstp
			sts.MessageID = stsSparkPost.Msys.MessageEvent.MessageID
			sts.Event = stsSparkPost.Msys.MessageEvent.EventID
			sts.Email = stsSparkPost.Msys.MessageEvent.RcptTo
			var iCpgInt64 int64
			if stsSparkPost.Msys.MessageEvent.TemplateID != "" {
				iCpgInt64, _ = strconv.ParseInt(stsSparkPost.Msys.MessageEvent.TemplateID, 10, 64)
			} else {
				iCpgInt64, _ = strconv.ParseInt("0", 10, 64)
			}
			sts.MjCampaignId = iCpgInt64

			var iCtcInt64 int64
			if stsSparkPost.Msys.MessageEvent.CustomerID != "" {
				iCtcInt64, _ = strconv.ParseInt(stsSparkPost.Msys.MessageEvent.CustomerID, 10, 64)
			} else {
				iCtcInt64, _ = strconv.ParseInt("0", 10, 64)
			}
			sts.MjContactId = iCtcInt64
			sts.CustomCampaign = stsSparkPost.Msys.MessageEvent.TemplateID
			sts.SmtpReply = stsSparkPost.Msys.MessageEvent.MsgFrom
			byte_plain, err_plain := json.Marshal(stsSparkPost)
			_ = err_plain
			sts.StatusJson = string(byte_plain)
		} else if (structs.MessageEvent{} == stsSparkPost.Msys.MessageEvent) {
			//fmt.Println("Track Event")
			sts.Channel = "Email"
			sts.Email = stsSparkPost.Msys.TrackEvent.RcptTo
			sts.Destination = stsSparkPost.Msys.TrackEvent.RcptTo
			sts.Status = stsSparkPost.Msys.TrackEvent.Type

			sts.Event = stsSparkPost.Msys.TrackEvent.EventID
			tstp, _ := strconv.ParseInt(stsSparkPost.Msys.TrackEvent.Timestamp, 10, 64)
			sts.Timestamp = tstp
			sts.MessageID = stsSparkPost.Msys.TrackEvent.MessageID
			sts.Event = stsSparkPost.Msys.TrackEvent.EventID
			sts.Email = stsSparkPost.Msys.TrackEvent.RcptTo
			var iCpgInt64 int64
			if stsSparkPost.Msys.TrackEvent.TemplateID != "" {
				iCpgInt64, _ = strconv.ParseInt(stsSparkPost.Msys.TrackEvent.TemplateID, 10, 64)
			} else {
				iCpgInt64, _ = strconv.ParseInt("0", 10, 64)
			}
			sts.MjCampaignId = iCpgInt64

			var iCtcInt64 int64
			if stsSparkPost.Msys.TrackEvent.CustomerID != "" {
				iCtcInt64, _ = strconv.ParseInt(stsSparkPost.Msys.TrackEvent.CustomerID, 10, 64)
			} else {
				iCtcInt64, _ = strconv.ParseInt("0", 10, 64)
			}
			sts.MjContactId = iCtcInt64
			sts.CustomCampaign = stsSparkPost.Msys.TrackEvent.TemplateID
			sts.SmtpReply = stsSparkPost.Msys.TrackEvent.MsgFrom
			byte_plain, err_plain := json.Marshal(stsSparkPost)
			_ = err_plain
			sts.StatusJson = string(byte_plain)
		} else {
			//fmt.Println("Unknown Event")
			byte_plain, err_plain := json.Marshal(stsSparkPost)
			_ = err_plain
			sts.StatusJson = string(byte_plain)
		}

		stsnew.Channel = sts.Channel
		stsnew.Email = sts.Email
		stsnew.Destination = sts.Destination
		stsnew.Status = sts.Status

		stsnew.Event = sts.Event
		stsnew.Timestamp = sts.Timestamp
		stsnew.MessageID = sts.MessageID
		stsnew.Event = sts.Event
		stsnew.Email = sts.Email
		stsnew.MjCampaignId = sts.MjCampaignId

		stsnew.MjContactId = sts.MjContactId
		stsnew.CustomCampaign = sts.CustomCampaign
		stsnew.SmtpReply = sts.SmtpReply
		stsnew.StatusJson = sts.StatusJson

		err = statusReportRepo.AddStatusLogReportPlain(stsnew, 0)

		if err != nil {
			errors.Message = err.Message
			errors.RespMessage = err.RespMessage
			errors.ReqMessage = err.Data
			errors.Code = err.Code
			return &errors

		}
	}

	//statusjsonbyte, _ := json.Marshal(sts)
	errors.Message = structs.Success
	//errors.ReqMessage = string(statusjsonbyte)
	errors.Code = http.StatusOK

	return &errors
}

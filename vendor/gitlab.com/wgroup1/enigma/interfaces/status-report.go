package interfaces

import "gitlab.com/wgroup1/enigma/structs"

type StatusReportRepository interface {
	GetMessageByMessageId(sts structs.StatusReport) structs.StatusReport
	AddStatusLogReport(sts structs.StatusReport, conv_id int) *structs.ErrorMessage
	AddStatusLogReportPlain(sts structs.StatusReport, conv_id int) *structs.ErrorMessage
}

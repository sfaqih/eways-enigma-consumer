package interfaces

import (
	"gitlab.com/wgroup1/enigma/structs"
)

type FlowRepository interface {
	CreateFlow(fl structs.Flow) *structs.ErrorMessage
	GetRunningFlow(sts structs.StatusReport) (structs.Flow, *structs.ErrorMessage)
	GetRunningFlowN8N(sts structs.StatusReport) (structs.Flow, *structs.ErrorMessage)
}

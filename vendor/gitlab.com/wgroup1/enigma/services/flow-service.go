package service

import (
	"gitlab.com/wgroup1/enigma/interfaces"
	"gitlab.com/wgroup1/enigma/structs"
)

type FlowService interface {
	CreateFlow(fl structs.Flow) *structs.ErrorMessage
	//Validate(fl *structs.Flow) (*structs.Flow, *structs.ErrorMessage)
}

type flowService struct{}

var (
	flowRepo interfaces.FlowRepository
)

func NewFlowService(repository interfaces.FlowRepository) FlowService {
	flowRepo = repository
	return &flowService{}
}

func (*flowService) CreateFlow(fl structs.Flow) *structs.ErrorMessage {
	return flowRepo.CreateFlow(fl)
}

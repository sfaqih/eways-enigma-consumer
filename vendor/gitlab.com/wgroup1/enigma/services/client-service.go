package service

import (
	"gitlab.com/wgroup1/enigma/interfaces"
)

type ClientService interface {
	GetClientID(clientCode string) int
}

type clientService struct{}

var (
	clientRepo interfaces.ClientRepository
)

func NewClientService(repository interfaces.ClientRepository) ClientService {
	clientRepo = repository
	return &clientService{}
}

func (*clientService) GetClientID(clientCode string) int {
	return clientRepo.GetClientID(clientCode)
}

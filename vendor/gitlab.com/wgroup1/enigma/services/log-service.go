package service

import (
	"log"

	"gitlab.com/wgroup1/enigma/interfaces"
	"gitlab.com/wgroup1/enigma/structs"
)

type logService struct{
	logRepo interfaces.LogRepository
}

type LogService interface {
	InsertLogBulk(l []*structs.HTTPRequest)
}

// var logRepo interfaces.LogRepository?

func NewLogService(repository interfaces.LogRepository) LogService {
	// logRepo = repository
	return &logService{
		logRepo: repository,
	}
}


func (lg *logService) InsertLogBulk(l []*structs.HTTPRequest) {
	err := lg.logRepo.InsertLogBulk(l)

	if err.Code != 200 {
		log.Println("error when insert api_log bulk : ", err.Message)
	}else {
		log.Println("success insert api_log bulk")
	}

}
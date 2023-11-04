package interfaces

import "gitlab.com/wgroup1/enigma/structs"

type LogRepository interface {
	InsertLogBulk(l []*structs.HTTPRequest) *structs.ErrorMessage
}
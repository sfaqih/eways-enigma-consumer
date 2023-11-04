package interfaces

import "gitlab.com/wgroup1/enigma/structs"

type InboundWARepository interface {
	CreateInboundWA(inb structs.Message, cli int) *structs.ErrorMessage
	CreateInboundWADamcorp(iw []byte, vss []structs.WebHook) *structs.ErrorMessage
}

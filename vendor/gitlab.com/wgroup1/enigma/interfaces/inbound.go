package interfaces

import (
	"context"

	"gitlab.com/wgroup1/enigma/structs"
)

type InboundRepository interface {
	//GetDynamicStructByChannel(inboundChannel string, fromsender string, tosender string) ([]structs.VendorService, *structs.ErrorMessage)
	GetWebhook(inboundChannel string, fromsender string, tosender string, trx_type string) ([]structs.WebHook, *structs.ErrorMessage)
	CreateInbound(inm structs.Inbound, iiw structs.InternalInboundWa, respStr string, cli int, ctx context.Context) *structs.ErrorMessage
}

package interfaces

import (
	"gitlab.com/wgroup1/enigma/structs"
)

type OutboundRepository interface {
	GetToken(clientID int) ([]structs.VendorService, *structs.ErrorMessage)
	GetTokenStatus(vendoralias string, channel string) ([]structs.VendorService, *structs.ErrorMessage)
	CreateOutbound(outm structs.OutMessage, reqStr string, respStr string, cli int) *structs.ErrorMessage
	GetMessageStatus() ([]structs.Status, *structs.ErrorMessage)
	UpdateStatusSent(stats structs.Status, body []byte) (bool, error)
	UpdateStatusDelivered(stats structs.Status, body []byte) (bool, error)
	UpdateStatusRead(stats structs.Status, body []byte) (bool, error)
	UpdateStatusFailed(stats structs.Status, body []byte) (bool, error)
	UpdateStatusPending(stats structs.Status, body []byte) (bool, error)
	UpdateStatusRejected(stats structs.Status, body []byte) (bool, error)
	UpdateStatusReceived(stats structs.Status, body []byte) (bool, error)
	CreateOutboundBulk(outm []*structs.HTTPRequest, cli int) *structs.ErrorMessage
	CreateOutboundBulkLog(outb structs.OutboundBulkLog) (*structs.OutboundBulkLog, error)
}

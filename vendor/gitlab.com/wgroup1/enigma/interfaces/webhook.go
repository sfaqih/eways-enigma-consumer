package interfaces

import "gitlab.com/wgroup1/enigma/structs"

type WebHookRepository interface {
	GetURL(clientID int) []structs.WebHook
	GetWebHooks(ClientID int, page string, limit string) []structs.WebHookPlain
	GetWebHook(ClientID int, webhookId int) structs.WebHookPlain
	AddWebHook(wh structs.WebHookObj, respStr string) *structs.ErrorMessage
	UpdateWebHook(wh structs.WebHookObj, respStr string) *structs.ErrorMessage
}

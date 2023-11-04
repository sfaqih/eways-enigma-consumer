package interfaces

import "gitlab.com/wgroup1/enigma/structs"

type ConversationRepository interface {
	ReceiveConversation(conv structs.Incoming) (*structs.ErrorMessage, []*structs.WebHook)
	// GetWebhookConversation(conv *structs.Incoming) []*structs.WebHook
}

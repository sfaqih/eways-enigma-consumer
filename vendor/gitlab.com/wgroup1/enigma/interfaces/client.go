package interfaces

type ClientRepository interface {
	GetClientID(clientCode string) int
}

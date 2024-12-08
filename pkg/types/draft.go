package types

type DraftServiceInterface interface {
	HandleEvent(event *Event) (bool, error)
	Disconnect()
}

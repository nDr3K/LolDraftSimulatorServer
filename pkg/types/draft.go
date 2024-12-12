package types

type DraftServiceInterface interface {
	HandleEvent(event *Event, sendStateFunc func(*Lobby)) (bool, error)
}

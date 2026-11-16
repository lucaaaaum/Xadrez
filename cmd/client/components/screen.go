package components

import "context"

type Event struct {
	Id      string
	Type    string
	Payload any
}

type Screen interface {
	Update(ctx context.Context) (Screen, error)
	Render() error
	HandleEvent(event Event)
}

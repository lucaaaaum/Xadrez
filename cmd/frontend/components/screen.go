package components

import "context"

type Screen interface {
	Update(ctx context.Context) (Screen, error)
	Render() error
}

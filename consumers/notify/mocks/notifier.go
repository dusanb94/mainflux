package mocks

import (
	"github.com/mainflux/mainflux/consumers/notify"
	"github.com/mainflux/mainflux/pkg/messaging"
)

var _ notify.Notifier = (*notifier)(nil)

type notifier struct{}

// NewNotifier returns a new Notifier mock.
func NewNotifier() notify.Notifier {
	return notifier{}
}

func (n notifier) Notify(from string, to []string, msg messaging.Message) error {
	return nil
}

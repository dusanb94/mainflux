package notify

import (
	"errors"

	"github.com/mainflux/mainflux/pkg/messaging"
)

// ErrNotify wraps sending notification errors,
var ErrNotify = errors.New("Error sending notification")

// Notifier represents an API for sending notification.
type Notifier interface {
	Notify(from string, to []string, msg messaging.Message) error
}

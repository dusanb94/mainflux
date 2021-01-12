package notify

import (
	"github.com/mainflux/mainflux/pkg/messaging"
)

// Notifier represents an API for sending notification.
type Notifier interface {
	Notify(from string, to []string, msg messaging.Message) error
}

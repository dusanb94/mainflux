// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package consumers

// Consumer specifies message writing API.
type Consumer interface {
	// MessageConsumer method is used to save published message. A non-nil
	// error is returned to indicate  operation failure.
	Consume(messages interface{}) error
}

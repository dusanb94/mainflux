// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package smpp

import (
	"crypto/tls"
	"time"
)

// Config represents SMPP transmitter configuration.
type Config struct {
	Addr               string        // Server address in form of host:port.
	User               string        // Username.
	Passwd             string        // Password.
	SystemType         string        // System type, default empty.
	EnquireLink        time.Duration // Enquire link interval, default 10s.
	EnquireLinkTimeout time.Duration // Time after last EnquireLink response when connection considered down
	RespTimeout        time.Duration // Response timeout, default 1s.
	BindInterval       time.Duration // Binding retry interval
	TLS                *tls.Config   // TLS client settings, optional.
}

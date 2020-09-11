// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/coap"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	protocol = "coap"

// 	senMLJSON gocoap.MediaType = 110
// 	senMLCBOR gocoap.MediaType = 112
)

// var (
// 	errBadRequest        = errors.New("bad request")
// 	errBadOption         = errors.New("bad option")
// 	errMalformedSubtopic = errors.New("malformed subtopic")
// 	channelRegExp        = regexp.MustCompile(`^/?channels/([\w\-]+)/messages(/[^?]*)?(\?.*)?$`)
// )

var (
	// 	auth       mainflux.ThingsServiceClient
	logger log.Logger

// 	pingPeriod time.Duration
)

//MakeHTTPHandler creates handler for version endpoint.
func MakeHTTPHandler() http.Handler {
	b := bone.New()
	b.GetFunc("/version", mainflux.Version(protocol))
	b.Handle("/metrics", promhttp.Handler())

	return b
}

// MakeCoAPHandler creates handler for CoAP messages.
func MakeCoAPHandler(svc coap.Service, l log.Logger) mux.HandlerFunc {
	logger = l

	return handler(svc)
}

func handler(svc coap.Service) func(w mux.ResponseWriter, m *mux.Message) {
	return func(w mux.ResponseWriter, m *mux.Message) {
		if m.Options == nil {
			logger.Warn("Nil options")
			return // Handle not return! defer sendresp?
		}
		msg, err := decodeMessage(m)
		if err != nil {
			logger.Warn(fmt.Sprintf("Error parsing path: %s", err))
			return
		}
		key, err := parseKey(m)
		if err != nil {
			logger.Warn(fmt.Sprintf("Error parsing auth: %s", err))
			return
		}
		endpoint := fmt.Sprintf("%s.%s", msg.Channel, msg.Subtopic)

		customResp := message.Message{
			Code:    codes.Content,
			Token:   m.Token,
			Context: m.Context,
			Options: make(message.Options, 0, 16),
		}
		switch m.Code {
		case codes.GET:
			obs, err := m.Options.Observe()
			if err != nil {
				logger.Warn(fmt.Sprintf("Error reading observe option: %s", err))
				break
			}
			if obs == 0 {
				o := coap.NewObserver(w.Client(), m.Token)
				svc.Subscribe(key, endpoint, o)
				break
			}
			svc.Unsubscribe(key, endpoint, m.Token.String())
		case codes.POST:
			svc.Publish(key, msg)
		}

		if err := w.Client().WriteMessage(&customResp); err != nil {
			logger.Warn(fmt.Sprintf("Can't set response: %v", err))
		}
	}
}

func decodeMessage(msg *mux.Message) (messaging.Message, error) {
	path, err := msg.Options.Path()
	if err != nil {
		return messaging.Message{}, err
	}
	ret := messaging.Message{
		Protocol: protocol,
		Channel:  parseID(path),
		Subtopic: parseSubtopic(path),
		Payload:  []byte{},
		Created:  time.Now().UnixNano(),
	}

	if msg.Body != nil {
		var err error
		var n int
		buff := make([]byte, 4096)
		for err != io.EOF {
			n, err = msg.Body.Read(buff)
			if err != nil && err != io.EOF {
				return ret, err
			}
			ret.Payload = append(ret.Payload, buff[:n]...)
		}
	}
	return ret, nil
}

func parseID(path string) string {
	vars := strings.Split(path, "/")
	if len(vars) > 1 {
		return vars[1]
	}
	return ""
}

func parseKey(msg *mux.Message) (string, error) {
	auth, err := msg.Options.GetString(message.URIQuery)
	if err != nil {
		return "", err
	}
	vars := strings.Split(auth, "=")
	if len(vars) != 2 || vars[0] != "auth" {
		return "", errors.New("failed auth")
	}
	return vars[1], nil
}

func parseSubtopic(path string) string {
	pos := 0
	for i, c := range path {
		if c == '/' {
			pos++
		}
		if pos == 3 {
			return strings.ReplaceAll(path[i+1:], "/", ".")
		}
	}
	return ""
}

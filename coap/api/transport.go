// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

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
			return
		}
		path, err := m.Options.Path()
		if err != nil {
			logger.Warn(fmt.Sprintf("Error parsing path: %s", err))
			return
		}
		chanID := parseID(path)
		key, err := parseKey(m)
		if err != nil {
			logger.Warn(fmt.Sprintf("Error parsing auth: %s", err))
			return
		}
		// subtopic := parseSubtopic(path)
		fmt.Println(chanID)
		fmt.Println(m.Options.GetString(message.URIQuery))

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
				svc.Subscribe(key, chanID, o)
				break
			}
			svc.Unsubscribe(key, chanID, m.Token.String())
		case codes.POST:
			svc.Publish(key, messaging.Message{Payload: []byte("")})
		}

		if err := w.Client().WriteMessage(&customResp); err != nil {
			logger.Warn(fmt.Sprintf("Can't set response: %v", err))
		}
	}
}

// func getPath(opts message.Options) string {
// 	path, err := opts.Path()
// 	if err != nil {
// 		fmt.Printf("cannot get path: %v", err)
// 		return ""
// 	}
// 	return path
// }

// func parsePath(opts message.Options) (string, error) {
// 	path, err := opts.Path()
// 	if err != nil {
// 		return "", err
// 	}

// 	return path, nil
// }

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
			return path[i:]
		}
	}
	return ""
}

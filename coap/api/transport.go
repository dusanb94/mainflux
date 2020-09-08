// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"bytes"
	"fmt"
	"net/http"

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
func MakeCoAPHandler(svc coap.Service, l log.Logger) *mux.Router {
	logger = l
	r := mux.NewRouter()
	r.Handle("/channels", mux.HandlerFunc(handler(svc)))

	return r
}

func handler(svc coap.Service) func(w mux.ResponseWriter, m *mux.Message) {
	return func(w mux.ResponseWriter, m *mux.Message) {
		customResp := message.Message{
			Code:    codes.Content,
			Token:   m.Token,
			Context: m.Context,
			Options: make(message.Options, 0, 16),
			Body:    bytes.NewReader([]byte("B hello world")),
		}
		switch m.Code {
		case codes.GET:
			obs, err := m.Options.Observe()
			if err != nil {
				logger.Warn(fmt.Sprintf("Error reading observe option"))
			}
			endpoint, _ := m.Options.Path()
			if obs == 0 {
				o := coap.NewObserver(w.Client(), m.Token)
				fmt.Println("subscribed")
				svc.Subscribe(endpoint, o)
				return
			}
			svc.Unsubscribe(endpoint, m.Token.String())
		case codes.POST:
			svc.Publish(messaging.Message{Payload: []byte("")})
		}

		if err := w.Client().WriteMessage(&customResp); err != nil {
			logger.Warn(fmt.Sprintf("Can't set response: %v", err))
		}
	}
}

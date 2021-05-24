// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"fmt"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/readers"
)

type apiReq interface {
	validate() error
}

type listMessagesReq struct {
	pageMeta readers.PageMetadata
}

type query struct {
	ChannelID   string                 `json:"-" schema:"channel_id"`
	Offset      uint64                 `json:"-" schema:"offset"`
	Limit       uint64                 `json:"-" schema:"limit"`
	Subtopic    string                 `json:"-" schema:"subtopic"`
	Publisher   string                 `json:"-" schema:"publsher"`
	Protocol    string                 `json:"-" schema:"protocol"`
	Comparator  string                 `json:"-" schema:"comparator"`
	Name        string                 `json:"name,omitempty" schema:"name"`
	Value       float64                `json:"v,omitempty" schema:"v"`
	StringValue string                 `json:"vs,omitempty" schema:"vs"`
	DataValue   string                 `json:"vd,omitempty" schema:"vd"`
	BoolValue   bool                   `json:"vb,omitempty" schema:"vb"`
	From        float64                `json:"from,omitempty" schema:"from"`
	To          float64                `json:"to,omitempty" schema:"to"`
	Format      string                 `json:"format,omitempty" schema:"format"`
	Query       map[string]interface{} `json:"query,omitempty"`
}

func (q query) toMap() (map[string]interface{}, error) {
	data, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]interface{})
	if err := json.Unmarshal(data, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (q query) toPageMeta() (readers.PageMetadata, error) {
	m, err := q.toMap()
	if err != nil {
		return readers.PageMetadata{}, err
	}
	ret := readers.PageMetadata{
		Offset:    q.Offset,
		Limit:     q.Limit,
		Publisher: q.Publisher,
		Protocol:  q.Protocol,
		Subtopic:  q.Subtopic,
		Format:    q.Format,
		Query:     m,
	}
	return ret, nil
}

func (req listMessagesReq) validate() error {
	fmt.Println(req.pageMeta.Limit)
	fmt.Println(req.pageMeta.Offset)
	if req.pageMeta.Limit < 1 || req.pageMeta.Offset < 0 {
		return errors.ErrInvalidQueryParams
	}
	// if req.pageMeta.Comparator != "" &&
	// 	req.pageMeta.Comparator != readers.EqualKey &&
	// 	req.pageMeta.Comparator != readers.LowerThanKey &&
	// 	req.pageMeta.Comparator != readers.LowerThanEqualKey &&
	// 	req.pageMeta.Comparator != readers.GreaterThanKey &&
	// 	req.pageMeta.Comparator != readers.GreaterThanEqualKey {
	// 	return errors.ErrInvalidQueryParams
	// }

	return nil
}

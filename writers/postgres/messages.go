// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq" // required for DB access
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/transformers/vega"
	"github.com/mainflux/mainflux/writers"
)

const errInvalid = "invalid_text_representation"

var (
	// ErrInvalidMessage indicates that service received message that
	// doesn't fit required format.
	ErrInvalidMessage = errors.New("invalid message representation")
	errSaveMessage    = errors.New("failed to save message to postgres database")
	errTransRollback  = errors.New("failed to rollback transaction")
)

var _ writers.MessageRepository = (*postgresRepo)(nil)

type postgresRepo struct {
	db *sqlx.DB
}

// New returns new PostgreSQL writer.
// func New(db *sqlx.DB) writers.MessageRepository {
// 	return &postgresRepo{db: db}
// }

func (pr postgresRepo) Save(msg vega.Message) (err error) {
	q := `INSERT INTO messages (dev, mei, iccid, sn, isnp, num, mutc, reason, dutc,
	bat, temp, water, s_magnet, s_blocked, s_leakage, s_blowout VALUES (:dev, :mei, :iccid, :sn, :isnp, :num, :mutc, :reason, :dutc,
	:bat, :temp, :water, :s_magnet, :s_blocked, :s_leakage, :s_blowout);`

	dbMsg := toDBMessage(msg)
	if err != nil {
		return errors.Wrap(errSaveMessage, err)
	}

	if _, err := pr.db.NamedExec(q, dbMsg); err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok {
			switch pqErr.Code.Name() {
			case errInvalid:
				return errors.Wrap(errSaveMessage, ErrInvalidMessage)
			}
		}

		return errors.Wrap(errSaveMessage, err)
	}
	return err
}

type dbMessage struct {
	Dev      string  `db:"dev"`
	IMEI     string  `db:"imei"`
	ICCID    string  `db:"iccd"`
	SN       string  `db:"sn"`
	Isnp     string  `db:"isnp"`
	Num      int     `db:"num"`
	MUTC     float64 `db:"mutc"`
	Reason   string  `db:"reason"`
	DUTC     int     `db:"dutc"`
	Bat      int     `db:"bat"`
	Temp     float64 `db:"temp"`
	Water    string  `db:"water"`
	SMagnet  int     `db:"s_magnet"`
	SBlocked int     `db:"s_blocked"`
	SLeakage int     `db:"s_leakage"`
	SBlowout int     `db:"s_blowout"`
}

func toDBMessage(msg vega.Message) dbMessage {
	return dbMessage{
		Dev:      msg.D.Dev,
		IMEI:     msg.D.IMEI,
		ICCID:    msg.D.ICCID,
		SN:       msg.D.SN,
		Isnp:     msg.D.Isnp,
		Num:      msg.D.Num,
		MUTC:     msg.D.MUTC,
		Reason:   msg.D.Reason,
		DUTC:     msg.D.DUTC,
		Bat:      msg.D.Bat,
		Temp:     msg.D.Temp,
		Water:    msg.D.Water,
		SMagnet:  msg.D.SMagnet,
		SBlocked: msg.D.SBlocked,
		SLeakage: msg.D.SLeakage,
		SBlowout: msg.D.SBlowout,
	}
}

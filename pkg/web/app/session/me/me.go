package me

import (
	"net/http"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/session"
)

type Session interface {
	IsAuthed() bool
	ID() ID
}

type ses struct {
	isAuthed bool
	id       ID
}

func (s *ses) IsAuthed() bool {
	return s.isAuthed
}

func (s *ses) ID() ID {
	return s.id
}

func (s *ses) MarshalBinary() ([]byte, error) {
	bs := make([]byte, 17, 17)
	bs[0] = byte('t')
	if !s.isAuthed {
		bs[0] = byte('f')
	}
	err := s.id.MarshalBinaryTo(bs[1:])
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (s *ses) UnmarshalBinary(data []byte) error {
	s.isAuthed = string(data[0:1]) == `t`
	id := &ID{}
	e := id.UnmarshalBinary(data[1:])
	if e != nil {
		return e
	}
	PanicIfZeroID(*id)
	s.id = *id
	return nil
}

func Get(tlbx app.Tlbx) Session {
	s := session.Get(tlbx)
	ses := &ses{}
	if !s.Exists() {
		// if session doesnt exist create a new unauthed one
		ses.isAuthed = false
		ses.id = tlbx.NewID()
		bs, err := ses.MarshalBinary()
		PanicOn(err)
		s.Set(bs)
	} else {
		PanicOn(ses.UnmarshalBinary(s.Get()))
	}
	return ses
}

func Del(tlbx app.Tlbx) {
	session.Get(tlbx).Del()
}

func AuthedExists(tlbx app.Tlbx) bool {
	return session.Get(tlbx).Exists() && Get(tlbx).IsAuthed()
}

func AuthedGet(tlbx app.Tlbx) ID {
	ses := Get(tlbx)
	app.ReturnIf(!ses.IsAuthed(), http.StatusUnauthorized, "")
	return ses.ID()
}

func AuthedSet(tlbx app.Tlbx, me ID) {
	ses := &ses{
		isAuthed: true,
		id:       me,
	}
	bs, err := ses.MarshalBinary()
	PanicOn(err)
	session.Get(tlbx).Set(bs)
}

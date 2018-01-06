package main

import (
	"net/mail"
	"time"

	"github.com/mjl-/enmime"
)

type email struct {
	Date     time.Time
	Envelope enmime.Envelope
}

func (m email) AddrList(key string) []*mail.Address {
	l, _ := m.Envelope.AddressList(key)
	return l
}

func (m email) AddrListString(key string) string {
	l, _ := m.Envelope.AddressList(key)
	return addressListString(l)
}

func (m email) FirstAddr(key string) *mail.Address {
	l, _ := m.Envelope.AddressList(key)
	if len(l) == 0 {
		return nil
	}
	return l[0]
}

func (m email) FirstName(key string) string {
	a := m.FirstAddr(key)
	if a != nil {
		return a.Name
	}
	return ""
}

func (m email) FirstEmail(key string) string {
	a := m.FirstAddr(key)
	if a != nil {
		return a.Address
	}
	return ""
}

func (m email) Header(key string) string {
	return m.Envelope.GetHeader(key)
}

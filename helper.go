package main

import (
	"net/mail"
	"net/textproto"
	"os"
	"strings"

	"github.com/mjl-/enmime"
)

func addressListString(la []*mail.Address) string {
	ls := make([]string, len(la))
	for i, a := range la {
		ls[i] = a.String()
	}
	return strings.Join(ls, ", ")
}

func emptyMail() email {
	header := textproto.MIMEHeader(map[string][]string{
		"From": {settings.Address},
	})
	return email{
		Envelope: enmime.Envelope{
			Header: &header,
		},
	}
}

func quote(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, "\n", "\n> ", -1)
	if s != "" {
		s = "> " + s + "\n"
	}
	return s
}

func settingsPath() string {
	appdata := os.Getenv("APPDATA") // windows, but more helpful than just homedir
	if appdata == "" {
		home := os.Getenv("HOME") // unix
		if home == "" {
			home = os.Getenv("home") // plan 9
		}
		appdata = home + "/lib"
	}
	return appdata + "/duitmail/settings.json"
}

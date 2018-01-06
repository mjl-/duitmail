package main

import (
	"9fans.net/go/draw"
	"github.com/mjl-/duit"
)

type server struct {
	Username, Password, Host, Port string
	TLS                            bool
}

type serverUI struct {
	server server

	host, port, user, password *duit.Field
	tls                        *duit.Checkbox

	*duit.Grid
}

func newServerUI(bold *draw.Font, what string, server server) *serverUI {
	ui := &serverUI{
		server: server,

		host: &duit.Field{
			Placeholder: "localhost",
			Text:        server.Host,
		},
		port: &duit.Field{
			Placeholder: "", // xxx
			Text:        server.Port,
		},
		user: &duit.Field{
			Placeholder: "user name...",
			Text:        server.Username,
		},
		password: &duit.Field{
			Placeholder: "password...",
			Password:    true,
			Text:        server.Password,
		},
		tls: &duit.Checkbox{
			Checked: server.TLS,
		},
	}

	ui.Grid = &duit.Grid{
		Columns: 2,
		Padding: []duit.Space{
			duit.SpaceXY(4, 2),
			duit.SpaceXY(4, 2),
		},
		Halign: []duit.Halign{
			duit.HalignRight,
			duit.HalignLeft,
		},
		Valign: []duit.Valign{
			duit.ValignMiddle,
			duit.ValignMiddle,
		},
		Kids: duit.NewKids(
			&duit.Label{},
			&duit.Label{Text: what, Font: bold},
			&duit.Label{Text: "host"},
			ui.host,
			&duit.Label{Text: "port"},
			ui.port,
			&duit.Label{Text: "user"},
			ui.user,
			&duit.Label{Text: "password"},
			ui.password,
			ui.tls,
			&duit.Label{Text: "require TLS"},
		),
	}
	return ui
}

func (ui *serverUI) Server() server {
	return server{
		Host:     ui.host.Text,
		Port:     ui.port.Text,
		Username: ui.user.Text,
		Password: ui.password.Text,
		TLS:      ui.tls.Checked,
	}
}

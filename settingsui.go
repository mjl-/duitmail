package main

import (
	"bytes"
	"encoding/json"
	"net/mail"
	"os"
	"path"

	"9fans.net/go/draw"
	"github.com/mjl-/duit"
	fa "github.com/mjl-/fontawesome5"
)

type mailboxSettings struct {
	Name       string // displayed in UI
	Address    string // in "name <email>", default From address
	Downloads  string // where downloads are stored
	Signature  string // default signature for outgoing emails
	IMAP, SMTP server
}

func (s mailboxSettings) AddressEmail() string {
	a, _ := mail.ParseAddress(s.Address)
	if a != nil {
		return a.Address
	}
	return ""
}

type mailboxSettingsUI struct {
	mailboxSettings mailboxSettings

	name, address, downloads   *duit.Field
	signature                  *duit.Edit
	imapServerUI, smtpServerUI *serverUI

	*duit.Box
}

func newMailboxSettingsUI(bold, awesome *draw.Font, stop chan struct{}, dui *duit.DUI, mbSet mailboxSettings) *mailboxSettingsUI {
	ui := &mailboxSettingsUI{
		mailboxSettings: mbSet,

		name:      &duit.Field{Text: mbSet.Name},
		address:   &duit.Field{Text: mbSet.Address, Placeholder: "firstname lastname <name@examle.org>"},
		downloads: &duit.Field{Text: mbSet.Downloads},
		signature: duit.NewEdit(bytes.NewReader([]byte(mbSet.Signature))),

		imapServerUI: newServerUI(bold, "IMAP", mbSet.IMAP),
		smtpServerUI: newServerUI(bold, "SMTP", mbSet.SMTP),
	}

	emailSettingsUI := &duit.Grid{
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
			&duit.Label{Text: "mailbox name"},
			ui.name,
			&duit.Label{Text: "email address"},
			ui.address,
			&duit.Label{Text: "downloads"},
			ui.downloads,
			&duit.Label{Text: "signature"},
			&duit.Box{
				Height: dui.Scale(100),
				Kids:   duit.NewKids(ui.signature),
			},
		),
	}

	ui.Box = &duit.Box{
		Kids: duit.NewKids(
			duit.CenterUI(
				duit.SpaceXY(10, 10),
				&duit.Box{
					Kids: duit.NewKids(
						&duit.Tabs{
							Buttongroup: &duit.Buttongroup{
								Texts: []string{
									"Email",
									"Incoming",
									"Outgoing",
								},
							},
							UIs: []duit.UI{
								emailSettingsUI,
								ui.imapServerUI,
								ui.smtpServerUI,
							},
						},
						duit.CenterUI(
							duit.SpaceXY(10, 10),
							&duit.Button{
								Icon: duit.Icon{
									Font: awesome,
									Rune: fa.Save,
								},
								Text:     "save",
								Colorset: &dui.Primary,
								Click: func(r *duit.Result) {
									ui.saveSettings()
									close(stop)
								},
							},
						),
					),
				},
			),
		),
	}

	return ui
}

func (ui *mailboxSettingsUI) saveSettings() {
	settings := mailboxSettings{
		Name:      ui.name.Text,
		Address:   ui.address.Text,
		Downloads: ui.downloads.Text,
		Signature: ui.signature.Text(),
		IMAP:      ui.imapServerUI.Server(),
		SMTP:      ui.smtpServerUI.Server(),
	}
	p := settingsPath()
	os.MkdirAll(path.Dir(p), os.ModePerm)
	f, err := os.Create(p)
	check(err, "create config file")
	err = json.NewEncoder(f).Encode(settings)
	check(err, "writing settings")
	err = f.Close()
	check(err, "closing settings file")
}

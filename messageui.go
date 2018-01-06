package main

import (
	"bytes"
	"image"
	"io"
	"log"
	"net/mail"
	"net/textproto"
	"os"

	"github.com/mjl-/duit"
	"github.com/mjl-/enmime"
)

type messageUI struct {
	MailboxUI *mailboxUI
	Email     email
	*duit.Box
}

func newMessageUI(mbUI *mailboxUI, m email) *messageUI {
	edit := duit.NewEdit(bytes.NewReader([]byte(m.Envelope.Text)))

	// xxx pretty horrible, we cannot determine size before hand, and we only get an utf8reader from enmime
	var attachments duit.UI = &duit.Box{}
	if len(m.Envelope.Attachments) > 0 {
		var kids []*duit.Kid
		for i, p := range m.Envelope.Attachments {
			func(p *enmime.Part) {
				filename := &duit.Field{Text: p.FileName}
				size := "size?" // fmt.Sprintf("%.1fm", size)
				kids = append(kids,
					duit.NewKids(
						&duit.Label{Text: size},
						&duit.Button{
							Text: "save",
							Click: func(r *duit.Result) {
								path := settings.Downloads + "/" + filename.Text
								f, err := os.Create(path)
								check(err, "creating file")
								_, err = io.Copy(f, p.Utf8Reader)
								check(err, "writing to downloads")
								err = f.Close()
								check(err, "closing download")
								kids[i*3+1].UI = &duit.Label{Text: ""}
								r.Layout = true
							},
						},
						filename,
					)...,
				)
			}(p)
		}
		attachments = &duit.Grid{
			Columns: 3,
			Padding: duit.NSpace(3, duit.SpaceXY(4, 2)),
			Halign:  []duit.Halign{duit.HalignLeft, duit.HalignRight, duit.HalignRight},
			Valign:  []duit.Valign{duit.ValignMiddle, duit.ValignMiddle, duit.ValignMiddle},
			Kids:    kids,
		}
	}

	mailQuote := func() string {
		t := edit.Selection()
		if t == "" {
			t = edit.Text()
		}
		t = quote(t)
		if t != "" {
			return t + "\n"
		}
		return t
	}

	return &messageUI{
		MailboxUI: mbUI,
		Email:     m,
		Box: &duit.Box{
			Margin: image.Pt(0, 15),
			Kids: duit.NewKids(
				&duit.Box{
					Padding: duit.SpaceXY(duit.ScrollbarSize, 4),
					Margin:  image.Pt(4, 2),
					Kids: duit.NewKids(
						&duit.Button{
							Text: "archive",
							Click: func(r *duit.Result) {
								log.Printf("todo: archive email...")
							},
						},
						&duit.Button{
							Text: "reply",
							Click: func(r *duit.Result) {
								to := m.AddrListString("Reply-To")
								if to == "" {
									to = m.AddrListString("From")
								}
								header := textproto.MIMEHeader(map[string][]string{
									"From":    {settings.Address},
									"To":      {to},
									"Subject": {"re: " + m.Header("Subject")},
								})
								newMail := email{
									Envelope: enmime.Envelope{
										Text:   mailQuote(),
										Header: &header,
									},
								}
								go compose(newMail, m.Header("Message-ID"))
							},
						},
						&duit.Button{
							Text: "reply all",
							Click: func(r *duit.Result) {
								to := m.AddrList("Reply-To")
								if len(to) == 0 {
									to = m.AddrList("From")
								}
								self := settings.AddressEmail()
								filterSelf := func(l []*mail.Address) (r []*mail.Address) {
									for _, a := range l {
										if a.Address != self {
											r = append(r, a)
										}
									}
									return
								}
								to = append(to, filterSelf(m.AddrList("To"))...)
								cc := filterSelf(m.AddrList("Cc"))
								bcc := filterSelf(m.AddrList("Bcc"))
								header := textproto.MIMEHeader(map[string][]string{
									"From":    {settings.Address},
									"To":      {addressListString(to)},
									"Cc":      {addressListString(cc)},
									"Bcc":     {addressListString(bcc)},
									"Subject": {"re: " + m.Header("Subject")},
								})
								newMail := email{
									Envelope: enmime.Envelope{
										Text:   mailQuote(),
										Header: &header,
									},
								}
								go compose(newMail, m.Header("Message-ID"))
							},
						},
						&duit.Button{
							Text: "forward",
							Click: func(r *duit.Result) {
								header := textproto.MIMEHeader(map[string][]string{
									"From":    {settings.Address},
									"Subject": {"fwd: " + m.Header("Subject")},
								})
								newMail := email{
									Envelope: enmime.Envelope{
										Text:   mailQuote(),
										Header: &header,
									},
								}
								go compose(newMail, m.Header("Message-ID"))
							},
						},
						&duit.Button{
							Text: "delete",
							Click: func(r *duit.Result) {
								log.Printf("todo: delete email...")
							},
						},
						&duit.Button{
							Text: "spam",
							Click: func(r *duit.Result) {
								log.Printf("todo: delete & mark as spam...")
							},
						},
					),
				},
				&duit.Grid{
					Width:   -1,
					Columns: 2,
					Padding: []duit.Space{duit.Space{Top: 2, Right: 2, Bottom: 2, Left: duit.ScrollbarSize}, duit.Space{Top: 2, Right: 0, Bottom: 2, Left: 2}},
					Kids: duit.NewKids(
						&duit.Label{Text: "Date:"},
						&duit.Label{Text: m.Header("Date")},
						&duit.Label{Text: "From:"},
						&duit.Box{Width: -1, Kids: duit.NewKids(&duit.Label{Text: m.AddrListString("From")})},
						&duit.Label{Text: "To:"},
						&duit.Label{Text: m.AddrListString("Tp")},
						&duit.Label{Text: "Reply-To:"},
						&duit.Label{Text: m.AddrListString("Reply-To")},
						&duit.Label{Text: "Cc:"},
						&duit.Label{Text: m.AddrListString("Cc")},
						&duit.Label{Text: "Bcc:"},
						&duit.Label{Text: m.AddrListString("Bcc")},
						&duit.Label{Text: "Subject:"},
						&duit.Label{Text: m.Header("Subject")},
					),
				},
				attachments,
				edit,
			),
		},
	}
}

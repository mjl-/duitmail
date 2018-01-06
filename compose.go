package main

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"net/mail"

	"github.com/mjl-/duit"
)

func compose(m email, inReplyTo string) {
	dui, err := duit.NewDUI("compose", "700x700")
	check(err, "new ui")

	fromUI := &duit.Field{Text: m.AddrListString("From")}
	replyToUI := &duit.Field{Text: m.AddrListString("Reply-To")}
	toUI := &duit.Field{Text: m.AddrListString("To")}
	ccUI := &duit.Field{Text: m.AddrListString("Cc")}
	bccUI := &duit.Field{Text: m.AddrListString("Bcc")}
	subjectUI := &duit.Field{Text: m.Header("Subject")}

	text := m.Envelope.Text
	cursor := int64(len(text))
	if settings.Signature != "" {
		if text == "" {
			text = "\n"
		}
		text += "\n--\n" + settings.Signature
	}
	edit := duit.NewEdit(bytes.NewReader([]byte(text)))
	edit.SetCursor(cursor, -1)

	dui.Top = &duit.Box{
		Kids: duit.NewKids(
			&duit.Box{
				Padding: duit.SpaceXY(4, 4),
				Margin:  image.Pt(4, 2),
				Kids: duit.NewKids(
					&duit.Button{
						Text:    "send",
						Primary: true,
						Click: func(r *duit.Result) {
							var err error
							parseAddressList := func(what, s string) []*mail.Address {
								if s == "" {
									return nil
								}
								l, e := mail.ParseAddressList(s)
								if e != nil {
									err = fmt.Errorf("parsing %s: %s", what, e)
								}
								return l
							}
							from := parseAddressList("from", fromUI.Text)
							replyTo := parseAddressList("reply-to", replyToUI.Text)
							to := parseAddressList("to", toUI.Text)
							cc := parseAddressList("cc", ccUI.Text)
							bcc := parseAddressList("bcc", bccUI.Text)
							if err != nil {
								log.Printf("bad email, not sent: %s\n", err)
								return
							}
							subject := subjectUI.Text
							body := edit.Text()
							log.Printf("sending email...\n")
							go func() {
								err = sendmail(settings.SMTP, from, replyTo, to, cc, bcc, inReplyTo, subject, body)
								if err != nil {
									log.Printf("sendmail: %s\n", err)
								} else {
									log.Printf("email sent\n")
									// xxx copy to "sent" mailbox?
								}
							}()
						},
					},
					&duit.Button{
						Text: "attach",
						Click: func(r *duit.Result) {
							log.Printf("todo: attach file...")
						},
					},
					&duit.Button{
						Text: "cancel",
						Click: func(r *duit.Result) {
							log.Printf("todo: close window without killing the entire program...")
						},
					},
				),
			},
			&duit.Grid{
				Width:   -1,
				Padding: []duit.Space{duit.SpaceXY(4, 2), duit.SpaceXY(4, 2)},
				Columns: 2,
				Kids: duit.NewKids(
					&duit.Label{Text: "From"},
					fromUI,
					&duit.Label{Text: "Reply-To"},
					replyToUI,
					&duit.Label{Text: "To"},
					toUI,
					&duit.Label{Text: "Cc"},
					ccUI,
					&duit.Label{Text: "Bcc"},
					bccUI,
					&duit.Label{Text: "Subject"},
					subjectUI,
				),
			},
			edit,
		),
	}
	dui.Render()

	for {
		select {
		case e := <-dui.Events:
			dui.Event(e)
		}
	}
}

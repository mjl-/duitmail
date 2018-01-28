package main

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"net/mail"
	"os"

	"github.com/mjl-/duit"
	fa "github.com/mjl-/fontawesome5"
)

func compose(m email, inReplyTo string) {
	dui, err := duit.NewDUI("mail-compose", nil)
	check(err, "new ui")

	fontawesome, _ := dui.Display.OpenFont(os.Getenv("fontawesome"))

	icon := func(c rune) duit.Icon {
		return duit.Icon{
			Font: fontawesome,
			Rune: c,
		}
	}

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
	edit.SetCursor(duit.Cursor{Cur: cursor, Start: -1})

	stop := make(chan struct{})

	dui.Top.UI = &duit.Box{
		Kids: duit.NewKids(
			&duit.Box{
				Padding: duit.SpaceXY(4, 4),
				Margin:  image.Pt(4, 2),
				Kids: duit.NewKids(
					&duit.Button{
						Icon:     icon(fa.PaperPlane),
						Text:     "send",
						Colorset: &dui.Primary,
						Click: func() (e duit.Event) {
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
							return
						},
					},
					&duit.Button{
						Icon: icon(fa.Paperclip),
						Text: "attach",
						Click: func() (e duit.Event) {
							log.Printf("todo: attach file...")
							return
						},
					},
					&duit.Button{
						Icon: icon(fa.Times),
						Text: "cancel",
						Click: func() (e duit.Event) {
							close(stop)
							return
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
		case e := <-dui.Inputs:
			dui.Input(e)

		case <-dui.Done:
			return

		case <-stop:
			dui.Close()
			return
		}
	}
}

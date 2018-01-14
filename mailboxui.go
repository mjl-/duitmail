package main

import (
	"image"
	"time"

	"github.com/mjl-/duit"
	fa "github.com/mjl-/fontawesome5"
)

type mailboxUI struct {
	Mailbox mailbox
	duit.Box
}

type date struct {
	year  int
	month time.Month
	day   int
}

func newDate(year int, month time.Month, day int) date {
	return date{year, month, day}
}

func newMailboxUI(mb mailbox) *mailboxUI {
	messageRows := make([]*duit.Gridrow, len(mb.Emails))
	today := newDate(time.Now().Date())
	for i, m := range mb.Emails {
		d := newDate(m.Date.Date())
		var date string
		if d == today {
			date = m.Date.Format("15:04")
		} else {
			date = m.Date.Format("2006-01-02")
		}
		messageRows[i] = &duit.Gridrow{
			Values: []string{
				date,
				m.FirstEmail("From"),
				m.Header("Subject"),
			},
		}
	}

	mbUI := &mailboxUI{
		Mailbox: mb,
	}

	noMessageUI := duit.NewMiddle(duit.SpaceXY(10, 10), &duit.Label{Text: "select a message on the left"})

	messageBox := &duit.Box{
		Width:  -1,
		Height: -1,
		Kids:   duit.NewKids(noMessageUI),
	}

	var messageList *duit.Gridlist
	messageList = &duit.Gridlist{
		Striped: true,
		Header: duit.Gridrow{
			Values: []string{
				"date",
				"from",
				"subject",
			},
		},
		Rows:    messageRows,
		Padding: duit.SpaceXY(4, 2),
		Changed: func(index int, r *duit.Event) {
			defer mainDUI.MarkLayout(messageBox)
			row := messageList.Rows[index]
			var nui duit.UI = noMessageUI
			if row.Selected {
				if row.Value == nil {
					nui = newMessageUI(mbUI, mb.Emails[index])
				} else {
					nui = row.Value.(*messageUI)
				}
			}
			messageBox.Kids = duit.NewKids(nui)
		},
	}
	mbUI.Box = duit.Box{
		Kids: duit.NewKids(
			&duit.Split{
				Gutter: 1,
				Split: func(width int) []int {
					return []int{width / 2, width - width/2}
				},
				Kids: duit.NewKids(
					duit.NewBox(
						&duit.Box{
							Padding: duit.SpaceXY(duit.ScrollbarSize, 4),
							Margin:  image.Pt(4, 2),
							Kids: duit.NewKids(
								&duit.Button{
									Icon: icon(fa.Edit),
									Text: "new mail",
									Click: func(r *duit.Event) {
										go compose(emptyMail(), "")
									},
								},
								&duit.Label{
									Text: "mailbox connection status...",
								},
							),
						},
						duit.NewScroll(messageList),
					),
					messageBox,
				),
			},
		),
	}
	mbUI.Box.Kids[0].ID = "messages"
	return mbUI
}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net/mail"
	"os"
	"sort"
	"strings"

	"9fans.net/go/draw"
	"github.com/mjl-/duit"
	"github.com/mjl-/enmime"
	fa "github.com/mjl-/fontawesome5"
)

var (
	settings        mailboxSettings
	mainDUI         *duit.DUI
	mainFontawesome *draw.Font
)

func check(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s\n", msg, err)
	}
}

type mailbox struct {
	Name   string
	Emails []email
}

func icon(c rune) duit.Icon {
	return duit.Icon{
		Font: mainFontawesome,
		Rune: c,
	}
}

func openSettings() {
	dui, err := duit.NewDUI("mail settings", "600x500")
	check(err, "new dui")

	// xxx find better way of dealing with fonts in a dui...
	var bold *draw.Font
	if os.Getenv("boldfont") != "" {
		bold, _ = dui.Display.OpenFont(os.Getenv("boldfont"))
	}
	if bold == nil {
		bold = dui.Display.DefaultFont
	}

	awesome, _ := dui.Display.OpenFont(os.Getenv("fontawesome"))

	stop := make(chan struct{})
	dui.Top = newMailboxSettingsUI(bold, awesome, stop, dui, settings)
	dui.Render()
	for {
		select {
		case e := <-dui.Events:
			dui.Event(e)

		case <-dui.Done:
			return

		case <-stop:
			dui.Close()
			return
		}
	}
}

func fetchMail(r *duit.Result) {
	log.Printf("fetchMail, not yet")
}

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		log.Println("usage: duitmail")
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(flag.Args()) != 0 {
		flag.Usage()
		os.Exit(2)
	}

	f, err := os.Open(settingsPath())
	if os.IsNotExist(err) {
		// xxx some default settings? or open settings screen immediately/instead?
	} else {
		check(err, "opening config file")
		err = json.NewDecoder(f).Decode(&settings)
		check(err, "parsing config file")
	}

	dui, err := duit.NewDUI("mail", "1200x700")
	check(err, "new dui")

	mainDUI = dui
	mainFontawesome, err = dui.Display.OpenFont(os.Getenv("fontawesome"))
	if err != nil {
		log.Printf("icons (fontawesome) not available: %s\n", err)
	}

	// xxx this must be replaced with imap stuff
	var mailboxes []mailbox
	l, err := ioutil.ReadDir("local/mailboxes")
	check(err, "listing mailboxes")
	for _, fi := range l {
		// log.Printf("mailbox %s\n", fi.Name())
		mb := mailbox{
			Name: fi.Name(),
		}
		ll, err := ioutil.ReadDir("local/mailboxes/" + fi.Name())
		check(err, "listing mails")
		for _, ffi := range ll {
			p := fmt.Sprintf("local/mailboxes/%s/%s", fi.Name(), ffi.Name())
			// log.Printf("parsing %s\n", p)
			f, err := os.Open(p)
			check(err, "open email")
			envelope, err := enmime.ReadEnvelope(f)
			check(err, "parsing email")

			dateHeader := envelope.Header.Get("date")
			dateHeader = strings.TrimSpace(strings.Split(dateHeader, "(")[0])
			tm, err := mail.ParseDate(dateHeader)
			check(err, fmt.Sprintf("parsing date %q", dateHeader))

			email := email{
				Date:     tm,
				Envelope: *envelope,
			}
			mb.Emails = append(mb.Emails, email)
		}

		sort.Slice(mb.Emails, func(i, j int) bool {
			return mb.Emails[i].Date.After(mb.Emails[j].Date)
		})

		mailboxes = append(mailboxes, mb)
	}

	mailboxValues := make([]*duit.ListValue, len(mailboxes))
	for i, mb := range mailboxes {
		mbUI := newMailboxUI(mb)
		mailboxValues[i] = &duit.ListValue{
			Text:  mb.Name,
			Value: mbUI,
		}
	}

	noMailboxUI := duit.NewMiddle(&duit.Label{Text: "select a mailbox on the left"})
	mailboxBox := &duit.Box{
		Width:  -1,
		Height: -1,
		Kids:   duit.NewKids(noMailboxUI), // kids are replaced on mailbox selection
	}

	var mailboxList *duit.List
	mailboxList = &duit.List{
		Values: mailboxValues,
		Changed: func(index int, r *duit.Result) {
			lv := mailboxList.Values[index]
			var nui duit.UI = noMailboxUI
			if lv.Selected {
				// xxx create lazily?
				nui = lv.Value.(*mailboxUI)
			}
			mailboxBox.Kids = duit.NewKids(nui)
			r.Layout = true
		},
	}

	dui.Top = &duit.Box{
		Kids: duit.NewKids(
			&duit.Horizontal{
				Split: func(width int) []int {
					col1 := dui.Scale(150)
					return []int{col1, width - col1}
				},
				Kids: duit.NewKids(
					duit.NewBox(
						&duit.Box{
							Padding: duit.SpaceXY(duit.ScrollbarSize, 4),
							Margin:  image.Pt(4, 2),
							Kids: duit.NewKids(
								&duit.Button{
									Icon: icon(fa.Cogs),
									Text: "settings",
									Click: func(r *duit.Result) {
										go openSettings()
									},
								},
								&duit.Button{
									Icon:  icon(fa.Sync),
									Text:  "refresh",
									Click: fetchMail,
								},
							),
						},
						&duit.Scroll{
							Child: mailboxList,
						},
					),
					mailboxBox,
				),
			},
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

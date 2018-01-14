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
	"bytes"

	"github.com/mxk/go-imap/imap"
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
	dui.Top.UI = newMailboxSettingsUI(bold, awesome, stop, dui, settings)
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

func fetchMail(r *duit.Event) {
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

	noMailboxUI := duit.NewMiddle(duit.SpaceXY(10, 10), &duit.Label{Text: "select a mailbox on the left"})
	mailboxBox := &duit.Box{
		Width:  -1,
		Height: -1,
		Kids:   duit.NewKids(noMailboxUI), // kids are replaced on mailbox selection
	}

	var mailboxList *duit.List
	mailboxList = &duit.List{
		Values: mailboxValues,
		Changed: func(index int, r *duit.Event) {
			defer dui.MarkLayout(nil) // xxx more specific?
			lv := mailboxList.Values[index]
			var nui duit.UI = noMailboxUI
			if lv.Selected {
				// xxx create lazily?
				nui = lv.Value.(*mailboxUI)
			}
			mailboxBox.Kids = duit.NewKids(nui)
		},
	}

	status := &duit.Label{}

	dui.Top = duit.Kid{
		ID: "mailboxes",
		UI: &duit.Split{
			Gutter: 1,
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
							status,
							&duit.Button{
								Icon: icon(fa.Cogs),
								Text: "settings",
								Click: func(r *duit.Event) {
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
					duit.NewScroll(mailboxList),
				),
				mailboxBox,
			),
		},
	}
	dui.Render()

	if false {
		go func() {
			dui.Call <- func() {
				status.Text = "connecting..."
				dui.MarkLayout(status)
			}

			lcheck, handle := errorHandler(func(err error) {
				dui.Call <- func() {
					status.Text = "error: " + err.Error()
					dui.MarkLayout(status)
				}
			})
			defer handle()

			addr := fmt.Sprintf("%s:%s", settings.IMAP.Host, settings.IMAP.Port)
			var err error
			var c *imap.Client
			if settings.IMAP.TLS {
				c, err = imap.DialTLS(addr, nil)
			} else {
				c, err = imap.Dial(addr)
			}
			lcheck(err, "dial imap")

			_, err = c.Login(settings.IMAP.Username, settings.IMAP.Password)
			lcheck(err, "login")

			cmd, err := imap.Wait(c.LSub("", "%"))
			lcheck(err, "listing subscribed mailboxes")

			for _, rsp := range c.Data {
				log.Println("imap server:", rsp)
			}

			var mailboxNames []string
			selIndex := -1
			for i, line := range cmd.Data {
				mb := line.MailboxInfo()
				if mb.Name == "INBOX" {
					selIndex = i
				}
				mailboxNames = append(mailboxNames, mb.Name)
			}
			if selIndex < 0 {
				if len(mailboxNames) > 0 {
					selIndex = 0
				} else {
					lcheck(fmt.Errorf("no mailboxes"), "finding mailbox to open")
				}
			}

			dui.Call <- func() {
				status.Text = "messages..."
				dui.MarkLayout(status)
			}

			cmd, err = c.Select(mailboxNames[selIndex], false)
			lcheck(err, "select mailbox")

			seq, err := imap.NewSeqSet("1:*")
			lcheck(err, "newseqset")
			cmd, err = c.Fetch(seq, "RFC822.HEADER")
			lcheck(err, "message fetch")

			count := 0
			for cmd.InProgress() {
				// Wait for the next response (no timeout)
				c.Recv(-1)

				// Process command data
				for _, rsp := range cmd.Data {
					count++
					header := imap.AsBytes(rsp.MessageInfo().Attrs["RFC822.HEADER"])
					if msg, _ := mail.ReadMessage(bytes.NewReader(header)); msg != nil {
						fmt.Println("|--", msg.Header.Get("Subject"))
					}
				}
				cmd.Data = nil

				// Process unilateral server data
				for _, rsp := range c.Data {
					fmt.Println("Server data:", rsp)
				}
				c.Data = nil
			}

			// Check command completion status
			if rsp, err := cmd.Result(imap.OK); err != nil {
				if err == imap.ErrAborted {
					fmt.Println("Fetch command aborted")
				} else {
					fmt.Println("Fetch error:", rsp.Info)
				}
			}

			values := make([]*duit.ListValue, len(mailboxNames))
			for i,name := range mailboxNames {
				values[i] =  &duit.ListValue{
					Text: name,
					Selected: false, // xxx
				}
			}

			dui.Call <- func() {
				status.Text = fmt.Sprintf("%d messages", count)
				mailboxList.Values = values
				dui.MarkLayout(mailboxList)
			}
		}()
	}

	for {
		select {
		case e := <-dui.Inputs:
			dui.Input(e)

		case <-dui.Done:
			return
		}
	}
}

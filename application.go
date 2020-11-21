package main

import (
	ui "github.com/VladimirMarkelov/clui"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"io"
	"io/ioutil"
	"log"
)

func readMail(r imap.Literal) string {
	// Create a new mail reader
	mr, err := mail.CreateReader(r)
	if err != nil {
		log.Fatal(err)
	}

	// Print some info about the message
	header := mr.Header
	if date, err := header.Date(); err == nil {
		log.Println("Date:", date)
	}
	if from, err := header.AddressList("From"); err == nil {
		log.Println("From:", from)
	}
	if to, err := header.AddressList("To"); err == nil {
		log.Println("To:", to)
	}
	if subject, err := header.Subject(); err == nil {
		log.Println("Subject:", subject)
	}

	// Process each message's part
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		switch p.Header.(type) {
		case *mail.InlineHeader:
			// This is the message's text (can be plain-text or HTML)
			b, _ := ioutil.ReadAll(p.Body)
			return string(b)
			log.Println("Got text: %v", string(b))
		}
	}

	return ""
}

func main() {
	c, err := client.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		log.Fatal(err)
	}

	defer c.Logout()

	if err := c.Login("crutchcorntest@gmail.com", "Testtest123!"); err != nil {
		log.Fatal(err)
	}

	done := make(chan error, 1)

	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}

	from := uint32(1)
	to := mbox.Messages
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	ui.InitLibrary()
	defer ui.DeinitLibrary()

	wnd := ui.AddWindow(0, 0, 60, ui.AutoSize, "Emails in inbox")
	wnd.SetSizable(false)

	drawEmailContents := func(val string) {
		textView := ui.CreateTextView(wnd, 50, 12, 1)
		textView.AddText([]string{val})
	}

	drawEmailList := func() {
		frm := ui.CreateFrame(wnd, 50, 12, ui.BorderNone, ui.Fixed)
		frm.SetPack(ui.Vertical)
		frm.SetScrollable(true)

		for msg := range messages {
			btn := ui.CreateButton(frm, 40, ui.AutoSize, msg.Envelope.Subject, 1)

			btn.OnClick(func(ev ui.Event) {
				var section imap.BodySectionName
				r := msg.GetBody(&section)
				var contents = readMail(r)
				drawEmailContents(contents)
			})
		}
	}

	drawEmailList()

	ui.MainLoop()
}
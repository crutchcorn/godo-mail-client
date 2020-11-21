package main

import (
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/manifoldco/promptui"
	"io"
	"io/ioutil"
	"log"
)

func readMail(msg *imap.Message) string {
	var section imap.BodySectionName
	r := msg.GetBody(&section)

	// Create a new mail reader
	mr, err := mail.CreateReader(r)
	if err != nil {
		log.Fatal(err)
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
		}
	}

	return "Did not work"
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

	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}

	from := uint32(1)
	to := mbox.Messages
	seqSet := new(imap.SeqSet)
	seqSet.AddRange(from, to)

	// Get the whole message body
	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope}

	messages := make(chan *imap.Message, mbox.Messages)
	go func() {
		if err := c.Fetch(seqSet, items, messages); err != nil {
			log.Fatal(err)
		}
	}()

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "> {{ .Envelope.Subject | cyan }}",
		Inactive: "  {{ .Envelope.Subject | cyan }}",
		Selected: "* {{ .Envelope.Subject | red | cyan }}",
	}

	messageSlice := make([]*imap.Message, mbox.Messages)

	idx := 0
	for msg := range messages {
		messageSlice[idx] = msg
		idx++
	}

	prompt := promptui.Select{
		Label: "Select email",
		Items: messageSlice,
		Templates: templates,
	}

	i, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	var contents = readMail(messageSlice[i])
	println(contents)
}
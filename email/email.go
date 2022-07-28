package email

import (
	"os"
	"strings"

	"github.com/go-msvc/errors"
	"gopkg.in/gomail.v2"
)

var (
	gmailUsername    string
	gmailAppPassword string
	smtpAddr         string = "smtp.gmail.com"
	smtpPort         int    = 587
)

func init() {
	gmailUsername = os.Getenv("GMAIL_USERNAME")
	if gmailUsername == "" {
		panic("missing env GMAIL_USERNAME, expecting your gmail email account")
	}
	gmailAppPassword = os.Getenv("GMAIL_APP_PASSWORD")
	if gmailAppPassword == "" {
		panic("missing env GMAIL_APP_PASSWORD, expecting password generated for this app in gmail account settings")
	}
}

type Message struct {
	From                Email
	To                  []Email
	Cc                  []Email
	Bcc                 []Email
	Subject             string
	ContentType         string
	Content             string
	AttachmentFilenames []string
}

func (msg *Message) Validate() error {
	if err := msg.From.Validate(); err != nil {
		return errors.Wrapf(err, "invalid from")
	}
	if len(msg.To) < 1 {
		return errors.Errorf("missing to")
	}
	for index, to := range msg.To {
		if err := to.Validate(); err != nil {
			return errors.Wrapf(err, "invalid to[%d]:%+v", index, to)
		}
	}
	for index, cc := range msg.Cc {
		if err := cc.Validate(); err != nil {
			return errors.Wrapf(err, "invalid cc[%d]:%+v", index, cc)
		}
	}
	for index, bcc := range msg.Bcc {
		if err := bcc.Validate(); err != nil {
			return errors.Wrapf(err, "invalid bcc[%d]:%+v", index, bcc)
		}
	}
	msg.Subject = strings.TrimSpace(msg.Subject)
	if msg.Subject == "" {
		return errors.Errorf("missing subject")
	}
	if msg.ContentType == "" {
		return errors.Errorf("missing content-type")
	}
	if msg.Content == "" {
		return errors.Errorf("missing content")
	}
	return nil
}

type Email struct {
	Addr string
	Name string
}

func (email Email) Validate() error {
	if email.Addr == "" {
		return errors.Errorf("missing addr")
	}
	return nil
}

func Send(msg Message) error {
	if err := msg.Validate(); err != nil {
		return errors.Wrapf(err, "cannot send invalid email message")
	}
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(msg.From.Addr, msg.From.Name))

	list := []string{}
	for _, to := range msg.To {
		list = append(list, m.FormatAddress(to.Addr, to.Name))
	}
	m.SetHeader("To", list...)

	if len(msg.Cc) > 0 {
		list = []string{}
		for _, cc := range msg.Cc {
			list = append(list, m.FormatAddress(cc.Addr, cc.Name))
		}
		m.SetHeader("Cc", list...)
	}
	if len(msg.Bcc) > 0 {
		list = []string{}
		for _, bcc := range msg.Bcc {
			list = append(list, m.FormatAddress(bcc.Addr, bcc.Name))
		}
		m.SetHeader("Bcc", list...)
	}
	m.SetHeader("Subject", msg.Subject)
	m.SetBody(msg.ContentType, msg.Content) //"text/html", "Hello <b>Bob</b> and <i>Cora</i>!")
	for _, fn := range msg.AttachmentFilenames {
		m.Attach(fn) //"/home/Alex/lolcat.jpg")
	}

	d := gomail.NewDialer(smtpAddr, smtpPort, gmailUsername, gmailAppPassword)

	if err := d.DialAndSend(m); err != nil {
		return errors.Wrapf(err, "failed to send email")
	}
	return nil
}

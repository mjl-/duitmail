package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
)

func sendmail(smtpConfig server, from, replyTo, to, cc, bcc []*mail.Address, inReplyTo, subject, body string) (err error) {
	type smtpError struct {
		err error
	}
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if ee, ok := e.(*smtpError); ok {
			err = ee.err
			return
		}
		panic(e)
	}()
	xcheck := func(e error, msg string) {
		if e != nil {
			panic(&smtpError{fmt.Errorf("smtp: %s: %s", msg, e)})
		}
	}

	addr := fmt.Sprintf("%s:%s", smtpConfig.Host, smtpConfig.Port)
	c, err := smtp.Dial(addr)
	xcheck(err, "dial")
	defer func() {
		if c != nil {
			c.Close()
		}
	}()

	hostname, _ := os.Hostname()
	if hostname != "" {
		err = c.Hello(hostname)
		xcheck(err, "helo")
	}
	if smtpConfig.TLS {
		config := &tls.Config{ServerName: smtpConfig.Host}
		err = c.StartTLS(config)
		xcheck(err, "start tls")
	}
	if smtpConfig.Username != "" || smtpConfig.Password != "" {
		log.Printf("doing auth\n")
		auth := smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Host)
		err = c.Auth(auth)
		xcheck(err, "authentication")
	}

	for _, a := range from {
		log.Printf("mail from %s, %s\n", a.Address, a)
		err = c.Mail(a.Address)
		xcheck(err, "mail from")
	}
	for _, a := range append(to, append(cc, bcc...)...) {
		log.Printf("rcpt to %s, %s\n", a.Address, a)
		err = c.Rcpt(a.Address)
		xcheck(err, "rcpt to")
	}
	msg := ""
	addAddressList := func(key string, l []*mail.Address) {
		if len(l) > 0 {
			msg += fmt.Sprintf("%s: %s\r\n", key, addressListString(l))
		}
	}
	addAddressList("From", from)
	addAddressList("Reply-To", replyTo)
	addAddressList("To", to)
	addAddressList("Cc", cc)
	if inReplyTo != "" {
		msg += fmt.Sprintf("In-Reply-To: %s\r\n", inReplyTo)
	}
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "User-Agent: duitmail\r\n"
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/plain; charset=utf-8\r\n"
	msg += "Content-Disposition: inline\r\n"
	msg += "Content-Transfer-Encoding: 8bit\r\n" // xxx must encode if we have lines longer than 990 chars...
	msg += "\r\n"
	msg += strings.Replace(body, "\n", "\r\n", -1)

	data, err := c.Data()
	xcheck(err, "smtp data")
	_, err = data.Write([]byte(msg))
	xcheck(err, "writing email body")

	err = data.Close()
	xcheck(err, "flushing data")
	err = c.Close()
	xcheck(err, "closing smtp connection")
	c = nil
	return nil
}

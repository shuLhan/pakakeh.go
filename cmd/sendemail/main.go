// Program sendemail is command line interface that use lib/email and lib/smtp
// to send email.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/shuLhan/share/lib/email"
	"github.com/shuLhan/share/lib/smtp"
)

const (
	envSmtpUsername = "SMTP_USERNAME"
	envSmtpPassword = "SMTP_PASSWORD"
)

func main() {
	var (
		smtpc      *smtp.Client
		smtpRes    *smtp.Response
		mailtx     *smtp.MailTx
		clientOpts smtp.ClientOptions
		msg        email.Message
		err        error

		from         string
		to           string
		subject      string
		fileBodyText string
		fileBodyHtml string
		serverUrl    string

		content []byte
		mailb   []byte

		isHelp bool
	)

	log.SetFlags(0)
	log.SetPrefix("sendemail: ")

	flag.BoolVar(&isHelp, "help", false, "Print the command usage and flags.")
	flag.StringVar(&from, "from", "", "Set the sender address.")
	flag.StringVar(&to, "to", "", "Set the recipients.")
	flag.StringVar(&subject, "subject", "", "Set the subject.")
	flag.StringVar(&fileBodyText, "bodytext", "", "Set the text body from content of file.")
	flag.StringVar(&fileBodyHtml, "bodyhtml", "", "Set the HTML body from content of file.")
	flag.Usage = usage
	flag.Parse()

	if isHelp {
		usage()
		os.Exit(1)
	}

	serverUrl = flag.Arg(0)
	if len(serverUrl) == 0 {
		log.Printf("missing server URL")
		os.Exit(1)
	}

	if len(from) == 0 {
		log.Printf("missing -from flag")
		os.Exit(1)
	}
	err = msg.SetFrom(from)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if len(to) == 0 {
		log.Printf("missing -to flag")
		os.Exit(1)
	}
	err = msg.SetTo(to)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if len(subject) == 0 {
		log.Printf("missing -subject flag")
		os.Exit(1)
	}
	msg.SetSubject(subject)

	if len(fileBodyText) == 0 && len(fileBodyHtml) == 0 {
		log.Printf("missing -bodytext or -bodyhtml")
		os.Exit(1)
	}

	if len(fileBodyText) > 0 {
		content, err = os.ReadFile(fileBodyText)
		if err != nil {
			log.Println(err)
		}
		_ = msg.SetBodyText(content)
	}

	if len(fileBodyHtml) > 0 {
		content, err = os.ReadFile(fileBodyHtml)
		if err != nil {
			log.Println(err)
		}
		_ = msg.SetBodyHtml(content)
	}

	mailb, err = msg.Pack()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	fmt.Printf("mail data:\n%s\n", mailb)

	mailtx = smtp.NewMailTx(from, []string{to}, mailb)

	clientOpts = smtp.ClientOptions{
		ServerUrl:     serverUrl,
		AuthUser:      os.Getenv(envSmtpUsername),
		AuthPass:      os.Getenv(envSmtpPassword),
		AuthMechanism: smtp.SaslMechanismPlain,
	}

	smtpc, err = smtp.NewClient(clientOpts)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	smtpRes, err = smtpc.MailTx(mailtx)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	fmt.Printf("SMTP response: %+v\n", smtpRes)
}

func usage() {
	fmt.Println(`
sendemail - command line interface to send an email using SMTP.

== SYNOPSIS

	sendemail <FLAGS> <SERVER_URL>

== FLAGS

Unless otherwise noted, all of the flags below are required.`)

	fmt.Println()
	flag.PrintDefaults()

	fmt.Println(`
The 'from' and 'to' flags are set using mailbox format, for example
"John <john@domain.tld>, Jane <jane@domain.tld>"

Only one of 'bodytext' and 'bodyhtml' is required.

== SERVER_URL

The SERVER_URL argument define the SMTP server where the email will be submitted.
Its use the following URL format,

	<scheme://><ip_address / domain>[:port]

The scheme can be 'smtps' for SMTP with implicit TLS (port 465) or
'smtp+starttls' for SMTP with STARTTLS (port 587).

The following environment variables are read for authentication with SMTP,

* SMTP_USERNAME: for SMTP user
* SMTP_PASSWORD: for SMTP password

== EXAMPLES

Send an email with message body set and read from HTML file,

	$ SMTP_USERNAME=myuser
	$ SMTP_PASSWORD=mypass
	$ sendemail -from="my@email.tld" \
		-to="John <john@example.com>, Jane <jane@example.com>" \
		-subject="Happy new years!" \
		-bodytext=/path/to/message.html \
		smtps://mail.myserver.com`)
}

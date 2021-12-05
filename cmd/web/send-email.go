package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
)

func listenForMail() {
	// This function will run indefinitely in the background
	go func() {
		for {
			msg := <-app.MailChan
			sendMessage(msg)
		}
	}()
}

func sendMessage(mailData models.MailData) {
	// Create email server
	server := mail.NewSMTPClient()
	server.Host = "localhost"
	server.Port = 1025
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	// Connect to mail server
	client, err := server.Connect()
	if err != nil {
		app.ErrorLog.Println(err)
	}

	// Setup email message
	email := mail.NewMSG()
	email.SetFrom(mailData.From).AddTo(mailData.To).SetSubject(mailData.Subject)
	
	if mailData.Template == "" {
		email.SetBody(mail.TextHTML, mailData.Content)
	} else {
		// Read email template html file
		data, err := ioutil.ReadFile(fmt.Sprintf("./email-templates/%s", mailData.Template))
		if err != nil {
			app.ErrorLog.Println(err)
		}

		// Replace content of email template and set it as body of email message to be sent
		mailTemplate := string(data)
		msgToSend := strings.Replace(mailTemplate, "[%body%]", mailData.Content, 1)
		email.SetBody(mail.TextHTML, msgToSend)
	}

	// Send email from our email server
	err = email.Send(client)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Email sent!")
	}
}
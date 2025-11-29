package utils

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

func SendMail() {

	m := gomail.NewMessage()

	// Kis ki taraf se mail jayegi
	m.SetHeader("From", "muhammad.anas.khalid.13@gmail.com")

	m.SetHeader("To", "211370152@gift.edu.pk")

	m.SetHeader("Subject", "Test Email from Golang")

	m.SetBody(" library se aayi hai ye email", "Hello,\n\nThis is a test email sent from a Golang application using the gomail library.\n\nBest regards,\nGolang App")

	d := gomail.NewDialer("smtp.gmail.com", 587, "muhammad.anas.khalid.13@gmail.com", "atuy fmtq msqm epxh")

	if err := d.DialAndSend(m); err != nil {

		panic(err)
	}

	fmt.Println("Email successfully send ho gayi hai!")
}

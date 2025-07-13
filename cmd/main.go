package main

import (
	job "daily_content_generator/internal/job/scheduler"
	"log"
)

func main() {

	// subject := "Daily Newsletter"
	// content := "Hello, this is your daily newsletter content."

	// err := mailer.SendNewsletter(subject, content)
	// if err != nil {
	// 	panic(err)
	// }

	log.Println("Daily Content Generator is running...")

	job.InitialCronJob()

}

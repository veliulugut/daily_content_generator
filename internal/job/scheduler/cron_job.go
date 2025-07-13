package job

import (
	jobs "daily_content_generator/internal/job"
	"log"

	"github.com/robfig/cron/v3"
)

func InitialCronJob() {
	c := cron.New()

	// 9:00 AM every day
	c.AddFunc("0 9 * * *", func() {
		log.Println("Scheduling daily digest job at 9:00 AM")
		jobs.GenerateAndSendDigest()
	})

	// afternoon 13:00 every day
	c.AddFunc("0 13 * * *", func() {
		log.Println("Scheduling daily digest job at 1:00 PM")
		jobs.GenerateAndSendDigest()
	})

	// evening 21:00 every day
	c.AddFunc("0 21 * * *", func() {
		log.Println("Scheduling daily digest job at 9:00 PM")
		jobs.GenerateAndSendDigest()
	})

	// test job every 1 minute
	c.AddFunc("@every 1m", func() {
		log.Println("Cron: Every 1 minute")
		jobs.GenerateAndSendDigest()
	})

	c.Start()

	select {}

}

package main

import (
	job "daily_content_generator/internal/job/scheduler"
	"log"
)

func main() {
	log.Println("Daily Content Generator is running...")
	log.Println("Starting cron scheduler...")
	job.InitialCronJob()
}

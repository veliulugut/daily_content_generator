package job

import (
	"daily_content_generator/internal/fetcher"
	"daily_content_generator/internal/generator"
	"daily_content_generator/internal/mailer"
	"log"
	"time"
)

func GenerateAndSendDigest() {
	log.Println("Starting daily digest generation...")

	// devTo fetch data
	devtoData, err := fetcher.GetDevToArticles()
	if err != nil {
		log.Printf("Error fetching DevTo articles: %v", err)
	}

	// github fetch data
	githubData, err := fetcher.GetTrendingProjects()
	if err != nil {
		log.Printf("Error fetching GitHub trending projects: %v", err)
	}

	allItems := append(devtoData, githubData...)
	log.Printf("Total items collected: %d (DevTo: %d, GitHub: %d)",
		len(allItems), len(devtoData), len(githubData))

	if len(allItems) == 0 {
		log.Println("No items to send in the digest.")
		return
	}

	// summarize the top 8 most popular content (increased from 5)
	content, err := generator.GenerateContentByPopularity(allItems, 8, "")
	if err != nil {
		log.Printf("Error generating content: %v", err)
		return
	}

	// create email header
	subject := "📰 Daily Digest - " + time.Now().Format("02 Jan 2006")

	//email sending
	if err := mailer.SendNewsletter(subject, content); err != nil {
		log.Printf("Error sending newsletter: %v", err)
		return
	}

	log.Println("Daily digest sent successfully!")
}

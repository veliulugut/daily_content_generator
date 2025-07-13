package fetcher

import (
	"daily_content_generator/internal/generator"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Article struct {
	Title                string   `json:"title"`
	Description          string   `json:"description"`
	URL                  string   `json:"url"`
	TagList              []string `json:"tag_list"`
	PublicReactionsCount int      `json:"public_reactions_count"`
}

func GetDevToArticles() ([]generator.ContentItem, error) {
	log.Println("Fetching DevTo articles...")

	urls := []string{
		"https://dev.to/api/articles?per_page=10&top=7",          // Top articles from last week
		"https://dev.to/api/articles?per_page=10&tag=tutorial",   // Tutorial articles
		"https://dev.to/api/articles?per_page=10&tag=javascript", // JavaScript articles
		"https://dev.to/api/articles?per_page=10&tag=python",     // Python articles
		"https://dev.to/api/articles?per_page=10&tag=webdev",     // Web development
	}

	var allArticles []Article
	seenTitles := make(map[string]bool)

	for _, url := range urls {
		articles, err := fetchArticlesFromURL(url)
		if err != nil {
			log.Printf("Error fetching from %s: %v", url, err)
			continue
		}

		// Add unique articles only
		for _, article := range articles {
			titleKey := strings.ToLower(strings.TrimSpace(article.Title))
			if !seenTitles[titleKey] && article.PublicReactionsCount > 2 {
				seenTitles[titleKey] = true
				allArticles = append(allArticles, article)
			}
		}
	}

	log.Printf("DevTo: Collected %d unique articles", len(allArticles))

	var items []generator.ContentItem
	for _, article := range allArticles {
		summary := formatArticleForNewsletter(article)
		items = append(items, generator.ContentItem{
			Text:       summary,
			Popularity: article.PublicReactionsCount,
		})
	}

	return items, nil
}

func fetchArticlesFromURL(url string) ([]Article, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching articles: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var articles []Article
	if err := json.Unmarshal(body, &articles); err != nil {
		return nil, fmt.Errorf("error unmarshalling articles: %w", err)
	}

	return articles, nil
}

func formatArticleForNewsletter(article Article) string {
	description := strings.TrimSpace(article.Description)
	if len(description) > 150 {
		description = description[:147] + "..."
	}

	summary := fmt.Sprintf("**%s**\n%s", article.Title, description)

	if len(article.TagList) > 0 {
		// Limit to first 3 most relevant tags
		tags := article.TagList
		if len(tags) > 3 {
			tags = tags[:3]
		}
		summary += fmt.Sprintf("\nTags: %v", tags)
	}

	summary += fmt.Sprintf("\n👍 %d reactions", article.PublicReactionsCount)

	return summary
}

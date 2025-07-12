package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Article struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	TagList     []string `json:"tag_list"`
}

func GetDevToArticles() ([]string, error) {
	resp, err := http.Get("https://dev.to/api/articles")
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

	var articleSummaries []string
	for _, article := range articles {
		summary := fmt.Sprintf("**%s**\n%s\n[Read more](%s)", article.Title, article.Description, article.URL)
		if len(article.TagList) > 0 {
			summary += fmt.Sprintf("\nTags: %s", article.TagList)
		}

		articleSummaries = append(articleSummaries, summary)
	}

	return articleSummaries, nil
}

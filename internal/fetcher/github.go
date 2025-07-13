package fetcher

import (
	"daily_content_generator/internal/generator"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type TrendingProject struct {
	Name        string
	Description string
	Language    string
	Stars       string
	TodayStars  string
}

func GetTrendingProjects() ([]generator.ContentItem, error) {
	log.Println("Fetching GitHub trending projects...")

	var items []generator.ContentItem
	seenProjects := make(map[string]bool)

	languages := []string{
		"", // All languages (trending overall)
		"javascript",
		"python",
		"go",
		"typescript",
		"rust",
		"java",
	}

	for _, lang := range languages {
		projects, err := fetchTrendingByLanguage(lang)
		if err != nil {
			langLabel := lang
			if langLabel == "" {
				langLabel = "all"
			}
			log.Printf("Error fetching %s projects: %v", langLabel, err)
			continue
		}

		for _, project := range projects {
			projectKey := strings.ToLower(project.Name)
			if !seenProjects[projectKey] && project.Name != "" {
				seenProjects[projectKey] = true
				projectInfo := formatProjectInfo(project)
				popularity := parseStarCount(project.Stars)

				todayStars := parseStarCount(project.TodayStars)
				popularity += todayStars * 10

				items = append(items, generator.ContentItem{
					Text:       projectInfo,
					Popularity: popularity,
				})
			}
		}
	}

	log.Printf("GitHub: Collected %d projects", len(items))
	return items, nil
}

func fetchTrendingByLanguage(language string) ([]TrendingProject, error) {
	url := "https://github.com/trending"
	if language != "" {
		url += "/" + language
	}
	url += "?since=daily"
	doc, err := fetchGitHubPage(url)
	if err != nil {
		return nil, err
	}

	var projects []TrendingProject
	doc.Find("article.Box-row").Each(func(i int, s *goquery.Selection) {
		if i >= 5 {
			return
		}

		project := extractProjectInfo(s)
		if project.Name != "" && project.Description != "" {
			projects = append(projects, project)
		}
	})

	return projects, nil
}

func fetchGitHubPage(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching page %s: %w", url, err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing document: %w", err)
	}

	return doc, nil
}

func formatProjectInfo(project TrendingProject) string {
	info := fmt.Sprintf("**%s**", project.Name)

	if project.Language != "" {
		info += fmt.Sprintf(" (%s)", project.Language)
	}

	if project.Description != "" {
		// Clean and limit description length
		desc := strings.TrimSpace(project.Description)
		if len(desc) > 120 {
			desc = desc[:117] + "..."
		}
		info += fmt.Sprintf("\n%s", desc)
	}

	if project.Stars != "" {
		info += fmt.Sprintf("\n⭐ %s stars", project.Stars)
	}

	if project.TodayStars != "" {
		info += fmt.Sprintf(" (+%s today)", project.TodayStars)
	}

	return info
}

func extractProjectInfo(s *goquery.Selection) TrendingProject {
	project := TrendingProject{}

	// Extract repository name from h2 a tag
	titleLink := s.Find("h2.h3 a")
	if titleLink.Length() > 0 {
		href, _ := titleLink.Attr("href")
		// Parse name from href (e.g., "/owner/name" -> "owner/name")
		project.Name = strings.Trim(href, "/")
	}

	// Extract description
	project.Description = strings.TrimSpace(s.Find("p.color-fg-muted").Text())

	// Extract programming language
	langSpan := s.Find("span[itemprop='programmingLanguage']")
	if langSpan.Length() > 0 {
		project.Language = strings.TrimSpace(langSpan.Text())
	}

	// Extract star count
	project.Stars = extractStarCount(s)

	// Extract today's stars
	project.TodayStars = extractTodayStars(s)

	return project
}

func extractStarCount(s *goquery.Selection) string {
	starLink := s.Find("a[href*='/stargazers']")
	if starLink.Length() == 0 {
		return ""
	}

	starText := strings.TrimSpace(starLink.Text())
	re := regexp.MustCompile(`star\s+([\d,]+)`)
	matches := re.FindStringSubmatch(starText)

	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractTodayStars(s *goquery.Selection) string {
	todayStarsText := s.Find("span.d-inline-block.float-sm-right").Text()
	if todayStarsText == "" {
		return ""
	}

	re := regexp.MustCompile(`([\d,]+)\s+stars?\s+today`)
	matches := re.FindStringSubmatch(todayStarsText)

	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func parseStarCount(stars string) int {
	clean := strings.ReplaceAll(stars, ",", "")
	n, err := strconv.Atoi(clean)
	if err != nil {
		return 0
	}

	return n
}

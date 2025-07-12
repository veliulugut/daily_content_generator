package fetcher

import (
	"fmt"
	"net/http"
	"regexp"
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

func GetTrendingProjects() ([]string, error) {
	doc, err := fetchGitHubTrendingPage()
	if err != nil {
		return nil, err
	}

	var projects []string
	doc.Find("article.Box-row").Each(func(i int, s *goquery.Selection) {
		project := extractProjectInfo(s)
		if project.Name != "" {
			projectInfo := formatProjectInfo(project)
			projects = append(projects, projectInfo)
		}
	})

	return projects, nil
}

func fetchGitHubTrendingPage() (*goquery.Document, error) {
	res, err := http.Get("https://github.com/trending")
	if err != nil {
		return nil, fmt.Errorf("error fetching trending projects: %w", err)
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
		info += fmt.Sprintf("\n- %s", project.Description)
	}

	if project.Stars != "" {
		info += fmt.Sprintf("\n- ⭐ %s stars", project.Stars)
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

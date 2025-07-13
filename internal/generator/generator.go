package generator

import (
	"daily_content_generator/internal/summarizer"
	"log"
	"math/rand"
	"sort"
	"strings"
	"time"
)

type ContentItem struct {
	Text       string
	Popularity int
}

func GenerateContentByPopularity(allItems []ContentItem, count int, promt string) (string, error) {
	log.Printf("Generating content from %d items...", len(allItems))

	if len(allItems) == 0 {
		return "", nil
	}

	// 1. Remove duplicates and similar content
	allItems = removeDuplicateContent(allItems)

	// 2. Sort by popularity (high to low)
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].Popularity > allItems[j].Popularity
	})

	// 3. Select diverse content from different categories
	selectedItems := selectDiverseContent(allItems, count)
	log.Printf("Selected %d diverse items for newsletter", len(selectedItems))

	var texts []string
	for _, item := range selectedItems {
		texts = append(texts, item.Text)
	}

	input := strings.Join(texts, "\n\n---\n\n")

	result, err := summarizer.SummarizeGeminiContent(input)
	if err != nil {
		log.Printf("Error generating content: %v", err)
		return "", err
	}

	log.Printf("Content generated successfully (%d characters)", len(result))
	return result, nil
}

// removeDuplicateContent removes items with very similar titles or content
func removeDuplicateContent(items []ContentItem) []ContentItem {
	var unique []ContentItem
	seen := make(map[string]bool)

	for _, item := range items {
		// Extract the first line as title for comparison
		lines := strings.Split(item.Text, "\n")
		if len(lines) == 0 {
			continue
		}

		title := strings.ToLower(strings.TrimSpace(lines[0]))
		title = strings.ReplaceAll(title, "*", "") // Remove markdown formatting

		// Skip if too similar to already seen content
		if isContentSimilar(title, seen) {
			continue
		}

		// Mark keywords as seen
		words := strings.Fields(title)
		for _, word := range words {
			if len(word) > 3 { // Only consider meaningful words
				seen[word] = true
			}
		}

		unique = append(unique, item)
	}

	return unique
}

// isContentSimilar checks if content is too similar to already processed content
func isContentSimilar(title string, seen map[string]bool) bool {
	words := strings.Fields(title)
	matchCount := 0

	for _, word := range words {
		if len(word) > 3 && seen[word] {
			matchCount++
		}
	}

	// If more than 50% of meaningful words are already seen, consider it duplicate
	meaningfulWords := 0
	for _, word := range words {
		if len(word) > 3 {
			meaningfulWords++
		}
	}

	if meaningfulWords > 0 && float64(matchCount)/float64(meaningfulWords) > 0.5 {
		return true
	}

	return false
}

// selectDiverseContent ensures we get content from different categories/languages with randomization
func selectDiverseContent(items []ContentItem, count int) []ContentItem {
	if len(items) <= count {
		return items
	}

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	var selected []ContentItem
	githubItems := []ContentItem{}
	devtoItems := []ContentItem{}

	// Separate GitHub and DevTo items
	for _, item := range items {
		if isGitHubItem(item) {
			githubItems = append(githubItems, item)
		} else {
			devtoItems = append(devtoItems, item)
		}
	}

	log.Printf("Separated items: %d GitHub, %d DevTo", len(githubItems), len(devtoItems))

	// Randomly decide the mix ratio
	githubRatio := 0.3 + rand.Float64()*0.4 // Between 30%-70%
	targetGitHub := int(float64(count) * githubRatio)
	targetDevTo := count - targetGitHub

	log.Printf("Random selection target: %d GitHub (%.1f%%), %d DevTo",
		targetGitHub, githubRatio*100, targetDevTo)

	// Randomly shuffle both lists
	shuffleContentItems(githubItems)
	shuffleContentItems(devtoItems)

	// Select from GitHub (ensuring language diversity)
	selectedGitHub := selectDiverseGitHubItems(githubItems, targetGitHub)
	selected = append(selected, selectedGitHub...)

	// Select from DevTo (ensuring topic diversity)
	selectedDevTo := selectDiverseDevToItems(devtoItems, targetDevTo)
	selected = append(selected, selectedDevTo...)

	// If we still need more items, randomly pick from remaining
	if len(selected) < count {
		remaining := []ContentItem{}

		// Add remaining items
		for _, item := range githubItems[len(selectedGitHub):] {
			remaining = append(remaining, item)
		}
		for _, item := range devtoItems[len(selectedDevTo):] {
			remaining = append(remaining, item)
		}

		shuffleContentItems(remaining)
		needed := count - len(selected)
		if needed > len(remaining) {
			needed = len(remaining)
		}

		selected = append(selected, remaining[:needed]...)
	}

	// Final shuffle to mix GitHub and DevTo items
	shuffleContentItems(selected)

	return selected
}

// shuffleContentItems randomly shuffles a slice of ContentItem
func shuffleContentItems(items []ContentItem) {
	for i := len(items) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		items[i], items[j] = items[j], items[i]
	}
}

// isGitHubItem checks if the item is from GitHub
func isGitHubItem(item ContentItem) bool {
	return strings.Contains(item.Text, "⭐") || strings.Contains(item.Text, "stars")
}

// selectDiverseGitHubItems selects GitHub items with language diversity
func selectDiverseGitHubItems(items []ContentItem, count int) []ContentItem {
	if len(items) <= count {
		return items
	}

	var selected []ContentItem
	languageCount := make(map[string]int)

	for _, item := range items {
		if len(selected) >= count {
			break
		}

		// Extract language from item text
		language := extractLanguageFromGitHub(item.Text)

		// Limit items per language to ensure diversity
		maxPerLanguage := max(1, count/4) // At most 1/4 of items from same language
		if languageCount[language] < maxPerLanguage {
			selected = append(selected, item)
			languageCount[language]++
		}
	}

	return selected
}

// selectDiverseDevToItems selects DevTo items with topic diversity
func selectDiverseDevToItems(items []ContentItem, count int) []ContentItem {
	if len(items) <= count {
		return items
	}

	var selected []ContentItem
	tagCount := make(map[string]int)

	for _, item := range items {
		if len(selected) >= count {
			break
		}

		// Extract main topic/tag from item text
		mainTag := extractMainTagFromDevTo(item.Text)

		// Limit items per tag to ensure diversity
		maxPerTag := max(1, count/3) // At most 1/3 of items from same tag
		if tagCount[mainTag] < maxPerTag {
			selected = append(selected, item)
			tagCount[mainTag]++
		}
	}

	return selected
}

// extractLanguageFromGitHub extracts programming language from GitHub item
func extractLanguageFromGitHub(text string) string {
	// Look for (Language) pattern
	if start := strings.Index(text, "("); start != -1 {
		if end := strings.Index(text[start:], ")"); end != -1 {
			return text[start+1 : start+end]
		}
	}
	return "unknown"
}

// extractMainTagFromDevTo extracts main topic from DevTo item
func extractMainTagFromDevTo(text string) string {
	// Look for Tags: pattern
	if tagIndex := strings.Index(text, "Tags:"); tagIndex != -1 {
		tagLine := text[tagIndex+5:]
		if bracketEnd := strings.Index(tagLine, "]"); bracketEnd != -1 {
			tagsPart := tagLine[:bracketEnd]
			if bracketStart := strings.Index(tagsPart, "["); bracketStart != -1 {
				firstTag := strings.Split(tagsPart[bracketStart+1:], " ")[0]
				return strings.TrimSpace(firstTag)
			}
		}
	}

	// Fallback: use first word of title
	lines := strings.Split(text, "\n")
	if len(lines) > 0 {
		words := strings.Fields(lines[0])
		if len(words) > 0 {
			return strings.ToLower(strings.Trim(words[0], "*"))
		}
	}

	return "general"
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

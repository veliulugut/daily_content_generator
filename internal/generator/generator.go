package generator

import (
	"daily_content_generator/internal/summarizer"
	"sort"
	"strings"
)

type ContentItem struct {
	Text       string
	Popularity int
}

func GenerateContentByPopularity(allItems []ContentItem, count int, promt string) (string, error) {
	if len(allItems) == 0 {
		return "", nil
	}

	// 1. sort (high - low popularity)
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].Popularity > allItems[j].Popularity
	})

	if len(allItems) > count {
		allItems = allItems[:count]
	}

	var texts []string
	for _, item := range allItems {
		texts = append(texts, item.Text)
	}

	input := strings.Join(texts, "\n\n")

	return summarizer.SummarizeGeminiContent(input)
}

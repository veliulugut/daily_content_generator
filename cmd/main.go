package main

import (
	"daily_content_generator/internal/fetcher"
	"fmt"
)

func main() {

	// // Example usage of the summarizer
	// summary, err := summarizer.SummarizeGeminiContent("golang 1.24.1 released with new features and improvements")
	// if err != nil {
	// 	fmt.Println("Error summarizing content:", err)
	// }

	// fmt.Println("Summary:", summary)

	// Example usage of the fetcher
	articles, err := fetcher.GetDevToArticles()
	if err != nil {
		panic(err)
	}

	for _, article := range articles {
		fmt.Println(article)
	}
}

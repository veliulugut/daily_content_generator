package main

import (
	"daily_content_generator/internal/summarizer"
	"fmt"
)

func main() {

	summary, err := summarizer.SummarizeGeminiContent("golang 1.24.1 released with new features and improvements")
	if err != nil {
		fmt.Println("Error summarizing content:", err)
	}

	fmt.Println("Summary:", summary)
}

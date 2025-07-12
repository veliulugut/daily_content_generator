package summarizer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func SummarizeGeminiContent(input string) (string, error) {
	apiKey, err := getAPIKey()
	if err != nil {
		return "", err
	}

	reqBody, err := buildRequestBody(input)
	if err != nil {
		return "", err
	}

	respBytes, err := sendGeminiRequest(reqBody, apiKey)
	if err != nil {
		return "", err
	}

	summary, err := extractSummaryFromResponse(respBytes)
	if err != nil {
		return "", err
	}

	return summary, nil
}

func getAPIKey() (string, error) {
	err := godotenv.Load("../.env")
	if err != nil {
		return "", fmt.Errorf("error loading .env file: %w", err)
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("API key not found in environment variables")
	}
	return apiKey, nil
}

func buildRequestBody(input string) ([]byte, error) {
	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": `You are a professional tech editor writing for an audience of experienced software developers.

Your job is to research, curate, and summarize relevant software projects, tools, articles, and developments from platforms like GitHub, Dev.to, Hacker News, Reddit, newsletters, or technical blogs.

You will be given a list of **headlines or topics**. For each item, imagine you are researching it deeply—reading GitHub READMEs, issues, Hacker News comments, or blog posts—and write an original, meaningful editorial summary based on that.

## Your Objective

- Present the content clearly and naturally, like a daily or weekly technical newsletter for developers.
- Each summary should explain **what the project/tool/topic is**, **why it’s interesting**, and **who might benefit from it**.
- Rewrite everything in your own words; do not copy source phrasing.

## Content Structure

- Output the final result as **plain text only**.
- You may start with a brief 1–2 sentence introduction.
- Group content logically using paragraphs or bullet points, but do not use Markdown formatting (no headings, lists, or special markup).
- Ensure the text is readable and meaningful.
- Order items by relevance or thematic relation.

## Style and Tone Guidelines

- Write as a human tech editor, not an AI assistant.
- Use a professional, clear, and natural tone. Avoid promotional or hype language.
- Avoid clichés like “revolutionary”, “game-changer”, or “AI-powered” unless technically justified.
- Do not use emojis.
- Avoid generic or vague statements. Highlight what is genuinely useful or innovative to developers.

## What “Deep Research” Means

- When given a headline, imagine you are browsing the GitHub repo README, checking issues, reading Hacker News comments, or skimming blog posts related to it.
- Use this imagined context to produce meaningful insights and original summaries.
- If a headline does not yield meaningful content, you may omit it.

## Output

Return only the final editorial summaries as **plain text**, without any Markdown or other formatting. Do not include any extra explanations or notes.`,
					},
					{
						"text": input,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.7,
			"maxOutputTokens": 2048,
		},
	}

	return json.Marshal(payload)
}

func sendGeminiRequest(body []byte, apiKey string) ([]byte, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-goog-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func extractSummaryFromResponse(respBytes []byte) (string, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return "", fmt.Errorf("error unmarshalling response: %w", err)
	}

	candidates, ok := result["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return "", fmt.Errorf("no candidates in response: %s", respBytes)
	}

	content, ok := candidates[0].(map[string]interface{})["content"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("no content in response candidates: %s", respBytes)
	}

	parts, ok := content["parts"].([]interface{})
	if !ok || len(parts) == 0 {
		return "", fmt.Errorf("no parts in content: %s", respBytes)
	}

	text, ok := parts[0].(map[string]interface{})["text"].(string)
	if !ok {
		return "", fmt.Errorf("no text in first part of content: %s", respBytes)
	}

	return text, nil
}

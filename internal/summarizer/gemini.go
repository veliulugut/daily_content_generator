package summarizer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

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

	// Extra cleaning to ensure no HTML tags remain
	summary = cleanGeminiResponse(summary)

	return summary, nil
}

func cleanGeminiResponse(content string) string {
	// Remove any HTML tags that might have slipped through
	re := regexp.MustCompile(`<[^>]*>`)
	content = re.ReplaceAllString(content, "")

	// Remove HTML entities
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&gt;", ">")
	content = strings.ReplaceAll(content, "&amp;", "&")
	content = strings.ReplaceAll(content, "&quot;", "\"")
	content = strings.ReplaceAll(content, "&nbsp;", " ")

	// Remove any remaining angle brackets
	content = strings.ReplaceAll(content, "<", "")
	content = strings.ReplaceAll(content, ">", "")

	// Clean up extra whitespace
	content = strings.TrimSpace(content)

	return content
}

func getAPIKey() (string, error) {
	// Try multiple paths for .env file
	envPaths := []string{
		".env",          // When running from root directory
		"../.env",       // When running from cmd directory
		"../../.env",    // When running from internal subdirectory
		"../../../.env", // When running from internal/subdirectory
	}

	var lastErr error
	for _, path := range envPaths {
		err := godotenv.Load(path)
		if err == nil {
			break // Successfully loaded
		}
		lastErr = err
	}

	// If all paths failed, try to continue with system environment variables
	if lastErr != nil {
		// Don't return error immediately, maybe env vars are already set
		log.Printf("Warning: Could not load .env file, using system environment variables: %v", lastErr)
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
						"text": `You are a professional tech newsletter editor creating a digest for developers.

## Critical Format Requirements

Your output must follow this EXACT structure:

🚀 Trending GitHub Projects

ProjectName/RepoName (Language)
Brief 1-2 sentence description highlighting key features and practical value.

AnotherProject (Language)
Different description focusing on unique aspects and use cases.

📖 Developer Articles & Tutorials

Clear Article Title
Educational summary in 1-2 sentences about what developers will learn.

Another Article Title  
Different focus area with practical learning outcomes mentioned.

🛠️ Tools & Libraries

Tool/Library Name
What specific problem it solves and how it improves workflow.

💡 Tech Insights

Industry Topic or Trend
Brief insight about how this affects developers and development practices.

## Content Rules

1. ALWAYS use the 4 emoji section headers exactly as shown above
2. Each project/article gets its own title line followed by description
3. Keep descriptions to 1-2 sentences maximum
4. NO HTML tags, NO markdown formatting, NO extra symbols
5. Use plain text with line breaks for structure
6. Each section should have 2-3 items maximum
7. Focus on different technologies/topics in each item
8. Include programming language in parentheses for GitHub projects
9. Make each item distinct - no repetitive content

## Style Guidelines

- Write clearly and concisely
- Highlight practical value for developers
- Include specific technical details (languages, frameworks, metrics)
- Avoid marketing language and hype
- Focus on what makes each item unique and useful

Return only the formatted content following the structure above.`,
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

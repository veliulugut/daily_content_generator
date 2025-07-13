package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

//go:embed template.html
var templateFS embed.FS

// EmailData holds the data for email template
type EmailData struct {
	Subject string
	Date    string
	Content string
}

// generateEmailTemplate creates a professional HTML email template
func generateEmailTemplate(subject, content string) (string, error) {
	templateContent, err := templateFS.ReadFile("template.html")
	if err != nil {
		return "", fmt.Errorf("failed to read email template: %w", err)
	}

	// Parse template
	tmpl, err := template.New("email").Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse email template: %w", err)
	}

	// Prepare data
	data := EmailData{
		Subject: subject,
		Date:    time.Now().Format("January 2, 2006"),
		Content: content,
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute email template: %w", err)
	}

	return buf.String(), nil
}

func formatContentForEmail(content string) string {
	// Clean HTML tags first
	content = cleanHTML(content)

	// Split content into paragraphs
	paragraphs := strings.Split(content, "\n\n")
	var formattedParagraphs []string

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// Apply markdown formatting
		para = formatMarkdown(para)

		// Wrap in paragraph tags for HTML email
		formattedParagraphs = append(formattedParagraphs, "<p>"+para+"</p>")
	}

	return strings.Join(formattedParagraphs, "\n\n")
}

func cleanHTML(content string) string {
	// Remove all HTML tags using regex - be more aggressive
	re := regexp.MustCompile(`<[^>]*>`)
	content = re.ReplaceAllString(content, "")

	// Remove any remaining HTML-like patterns
	re2 := regexp.MustCompile(`&[a-zA-Z0-9#]+;`)
	content = re2.ReplaceAllString(content, "")

	// Handle common HTML entities
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&gt;", ">")
	content = strings.ReplaceAll(content, "&amp;", "&")
	content = strings.ReplaceAll(content, "&quot;", "\"")
	content = strings.ReplaceAll(content, "&apos;", "'")
	content = strings.ReplaceAll(content, "&nbsp;", " ")

	// Remove any remaining angle brackets
	content = strings.ReplaceAll(content, "<", "")
	content = strings.ReplaceAll(content, ">", "")

	// Remove any remaining HTML-like patterns
	content = strings.ReplaceAll(content, "</p>", "\n\n")
	content = strings.ReplaceAll(content, "<p>", "")
	content = strings.ReplaceAll(content, "<br>", "\n")
	content = strings.ReplaceAll(content, "<br/>", "\n")
	content = strings.ReplaceAll(content, "<br />", "\n")
	content = strings.ReplaceAll(content, "<div>", "")
	content = strings.ReplaceAll(content, "</div>", "\n")
	content = strings.ReplaceAll(content, "<span>", "")
	content = strings.ReplaceAll(content, "</span>", "")

	// Clean up extra whitespace and newlines
	lines := strings.Split(content, "\n")
	var cleanLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Remove multiple spaces
			line = strings.Join(strings.Fields(line), " ")
			cleanLines = append(cleanLines, line)
		}
	}

	// Join lines back and ensure proper paragraph breaks
	result := strings.Join(cleanLines, "\n")

	// Normalize line breaks - replace multiple newlines with double newlines
	result = strings.ReplaceAll(result, "\n\n\n", "\n\n")

	return result
}

func formatMarkdown(text string) string {
	// Handle bold text (**text**)
	for strings.Contains(text, "**") {
		first := strings.Index(text, "**")
		if first == -1 {
			break
		}
		second := strings.Index(text[first+2:], "**")
		if second == -1 {
			break
		}
		second += first + 2

		before := text[:first]
		content := text[first+2 : second]
		after := text[second+2:]
		text = before + "<strong>" + content + "</strong>" + after
	}

	// Handle italic text (*text*)
	for strings.Contains(text, "*") {
		first := strings.Index(text, "*")
		if first == -1 {
			break
		}
		second := strings.Index(text[first+1:], "*")
		if second == -1 {
			break
		}
		second += first + 1

		before := text[:first]
		content := text[first+1 : second]
		after := text[second+1:]
		text = before + "<em>" + content + "</em>" + after
	}

	return text
}

func SendNewsletter(subject, body string) error {

	if err := loadEnvVariables(); err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	from := os.Getenv("MAIL_FROM")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	toList := os.Getenv("MAIL_TO")

	if from == "" || password == "" || smtpHost == "" || smtpPort == "" || toList == "" {
		return fmt.Errorf("missing required email configuration: from=%s, password=%s, smtpHost=%s, smtpPort=%s, toList=%s", from, password, smtpHost, smtpPort, toList)
	}

	to := strings.Split(toList, ",")
	for i, email := range to {
		to[i] = strings.TrimSpace(email)
	}

	formattedContent := formatContentForEmail(body)

	htmlBody, err := generateEmailTemplate(subject, formattedContent)
	if err != nil {
		return fmt.Errorf("failed to generate email template: %w", err)
	}

	message := fmt.Sprintf("Subject: %s\r\n", subject) +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
		htmlBody

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Newsletter sent successfully to %d recipients", len(to))
	return nil
}

func loadEnvVariables() error {
	paths := []string{
		".env",          // When running from root directory
		"../.env",       // When running from cmd directory
		"../../.env",    // When running from internal subdirectory
		"../../../.env", // When running from internal/subdirectory
	}

	var lastErr error
	for _, path := range paths {
		if err := godotenv.Load(path); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}

	log.Printf("Warning: Could not load .env file, using system environment variables: %v", lastErr)
	return nil
}

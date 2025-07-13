# Daily Content Generator 📰

Automated newsletter generator that curates trending content from GitHub and Dev.to.

## ✨ Features

- **🎯 Smart Curation**: Trending GitHub projects + Dev.to articles
- **🎲 Random Mix**: Different content combination each time
- **📧 Email Templates**: Professional HTML newsletters
- **🔄 Auto Scheduling**: Daily delivery at 9 AM, 1 PM, 9 PM

## 🚀 Quick Start

### Setup

1. **Clone & Install**
   ```bash
   git clone https://github.com/veliulugut/daily_content_generator.git
   cd daily_content_generator
   go mod download
   ```

2. **Configure**
   ```bash
   cp .env_example .env
   ```
   
   Edit `.env`:
   ```env
   GEMINI_API_KEY="your_api_key"
   MAIL_FROM="your_email@gmail.com"
   SMTP_PASSWORD="your_app_password"
   MAIL_TO="recipient@email.com"
   ```

3. **Run**
   ```bash
   make run
   ```

## 📧 Gmail Setup

1. Enable 2-Factor Authentication
2. Generate App Password: Google Account → Security → App passwords
3. Use app password in `SMTP_PASSWORD`

## 🆘 Common Issues

**Email not sending?** Check Gmail app password and 2FA
**No content?** Check internet connection and API limits

---

**Made with ❤️ for developers**

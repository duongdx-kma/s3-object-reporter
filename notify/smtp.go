package notify

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"text/template"

	"s3-object-reporter/reports"
)

// EmailTemplate: html
const EmailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; }
        table { width: 100%%; border-collapse: collapse; margin-top: 10px; }
        th, td { border: 1px solid black; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <h2>{{ .ReportName }}</h2>
    <p><strong>Report Date:</strong> {{ .ReportDate }}</p>
    <table>
        <thead>
            <tr>
                <th>#</th>
                <th>Service Name</th>
                <th>Service Path</th>
                <th>Count</th>
            </tr>
        </thead>
        <tbody>
            {{ range $index, $service := .Services }}
            <tr>
                <td>{{ inc $index }}</td>
                <td>{{ $service.ServiceName }}</td>
                <td>{{ $service.ServicePath }}</td>
                <td>{{ $service.Count }}</td>
            </tr>
            {{ end }}
        </tbody>
    </table>
</body>
</html>
`

func inc(n int) int {
	return n + 1
}

func SMTP(source reports.Report) error {
	subject := source.ReportName

	from := os.Getenv("SMTP_FROM_EMAIL")
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASS")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	addr := fmt.Sprintf("%s:%s", host, port)

	recipients := strings.Split(os.Getenv("SMTP_TO_EMAIL"), ",")

	// Render HTML template
	template, err := template.New("email").Funcs(template.FuncMap{"inc": inc}).Parse(EmailTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	err = template.Execute(&body, source)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Format email message
	for _, recipient := range recipients {
		to := strings.TrimSpace(recipient)
		msg := []byte(fmt.Sprintf(
			"From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
			os.Getenv("SMTP_FROM_NAME"),
			from,
			to,
			subject,
			body.String(),
		))

		auth := smtp.PlainAuth("", user, password, host)

		err := smtp.SendMail(addr, auth, from, []string{to}, msg)
		if err != nil {
			return fmt.Errorf("SMTP: %v %v", host, err)
		}
	}
	return nil
}

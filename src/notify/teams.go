package notify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"s3-object-reporter/reports"
)

type TeamsMessage struct {
	Type        string       `json:"type"`
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	ContentType string       `json:"contentType"`
	Content     AdaptiveCard `json:"content"`
}

type AdaptiveCard struct {
	Schema  string        `json:"$schema"`
	Type    string        `json:"type"`
	Version string        `json:"version"`
	MSTeams MSTeams       `json:"msteams"`
	Body    []interface{} `json:"body"`
}

type MSTeams struct {
	Width string `json:"width"`
}

type TextBlock struct {
	Type   string `json:"type"`
	Text   string `json:"text"`
	Weight string `json:"weight,omitempty"`
	Size   string `json:"size,omitempty"`
	Wrap   bool   `json:"wrap"`
}

type Row struct {
	Type    string   `json:"type"`
	Columns []Column `json:"columns"`
}

type Column struct {
	Type  string      `json:"type"`
	Width string      `json:"width"`
	Items []TextBlock `json:"items"`
}

// SendToTeamsAlert sends a report to Microsoft Teams webhook
func SendToTeamsAlert(report reports.Report) error {
	webhookURL := os.Getenv("TEAMS_WEBHOOK_URL")
	if webhookURL == "" {
		fmt.Println("ERROR: Teams webhook URL is not set")
		return errors.New("Teams webhook URL is not set")
	}

	// Header
	header := []interface{}{
		TextBlock{Type: "TextBlock", Text: fmt.Sprintf("📢 **Report: %s**", report.ReportName), Weight: "bolder", Size: "Large"},
		TextBlock{Type: "TextBlock", Text: fmt.Sprintf("📅 **Date:** %s", report.ReportDate), Wrap: true},
		TextBlock{Type: "TextBlock", Text: "---", Wrap: true},
		TextBlock{Type: "TextBlock", Text: "**📌 Service Report:**", Weight: "bolder", Size: "Medium"},
	}

	// Table Header Row
	tableRows := []Row{
		{
			Type: "ColumnSet",
			Columns: []Column{
				{Type: "Column", Width: "50", Items: []TextBlock{{Type: "TextBlock", Text: "Service Name", Weight: "bolder"}}},
				{Type: "Column", Width: "30", Items: []TextBlock{{Type: "TextBlock", Text: "Count", Weight: "bolder"}}},
			},
		},
	}

	// Table Content
	for _, service := range report.Services {
		row := Row{
			Type: "ColumnSet",
			Columns: []Column{
				{Type: "Column", Width: "50", Items: []TextBlock{{Type: "TextBlock", Text: service.ServiceName, Wrap: true}}},
				{Type: "Column", Width: "30", Items: []TextBlock{{Type: "TextBlock", Text: fmt.Sprintf("%d", service.Count), Wrap: true}}},
			},
		}
		tableRows = append(tableRows, row)
	}

	// Adaptive Card
	message := TeamsMessage{
		Type: "message",
		Attachments: []Attachment{
			{
				ContentType: "application/vnd.microsoft.card.adaptive",
				Content: AdaptiveCard{
					Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
					Type:    "AdaptiveCard",
					Version: "1.4",
					MSTeams: MSTeams{Width: "Full"},
					Body:    append(header, convertRowsToAdaptiveElements(tableRows)...),
				},
			},
		},
	}

	// Marshal JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		fmt.Println("ERROR: Failed to marshal JSON:", err)
		return err
	}

	// Debug JSON Output
	fmt.Println("DEBUG: JSON Message:\n", string(jsonData))

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("ERROR: Failed to create HTTP request:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("ERROR: Failed to send request to Teams webhook:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		fmt.Printf("ERROR: Teams webhook error: received status code %d\n", resp.StatusCode)
		return fmt.Errorf("Teams webhook error: received status code %d", resp.StatusCode)
	}

	fmt.Println("DEBUG: Message sent successfully to Teams")
	return nil
}

// Convert Rows to Adaptive Elements
func convertRowsToAdaptiveElements(rows []Row) []interface{} {
	var elements []interface{}
	for _, row := range rows {
		elements = append(elements, row)
	}
	return elements
}

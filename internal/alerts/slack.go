package alerts

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

// slackMessage is the payload for a Slack incoming webhook
type slackMessage struct {
    Text        string            `json:"text,omitempty"`
    Attachments []slackAttachment `json:"attachments,omitempty"`
    Blocks      []slackBlock      `json:"blocks,omitempty"`
}

// Slack attachment structures for richer formatting (left color bar)
type slackAttachment struct {
    Color  string       `json:"color,omitempty"`
    Title  string       `json:"title,omitempty"`
    Text   string       `json:"text,omitempty"`
    Fields []slackField `json:"fields,omitempty"`
    MrkdwnIn []string   `json:"mrkdwn_in,omitempty"`
}

type slackField struct {
    Title string `json:"title"`
    Value string `json:"value"`
    Short bool   `json:"short"`
}

// Slack Block Kit minimal structures (used for header above attachment)
type slackBlockText struct {
    Type  string `json:"type"`
    Text  string `json:"text"`
    Emoji bool   `json:"emoji,omitempty"`
}

type slackBlock struct {
    Type string          `json:"type"`
    Text *slackBlockText `json:"text,omitempty"`
}

// SendSlackMessage posts a simple text message to a Slack webhook URL.
// It is safe to call in background tasks; timeouts and errors are returned but should not crash callers.
func SendSlackMessage(ctx context.Context, webhookURL, text string) error {
    if webhookURL == "" || text == "" {
        return fmt.Errorf("missing webhook URL or text")
    }

    payload := slackMessage{Text: text}
    body, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("marshal slack payload: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
    if err != nil {
        return fmt.Errorf("create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("post to slack: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
    }
    return nil
}

// SendSlackAlert posts a message using Slack attachments with a colored bar on the left.
// The attachment will use a red color to highlight alert severity.
func SendSlackAlert(ctx context.Context, webhookURL, title string, fields map[string]string, extraText string) error {
    if webhookURL == "" {
        return fmt.Errorf("missing webhook URL")
    }

    // Build a markdown text with clear section spacing
    project := fields["Project"]
    application := fields["Application"]
    namespace := fields["Namespace"]
    metric := fields["Metric"]
    reason := fields["Reason"]
    if reason == "" {
        reason = fields["Error"]
    }

    text := fmt.Sprintf("*Project*: %s\n*Application*: %s\n*Namespace*: %s\n\n*Metric*: %s\n*Reason*: %s",
        project, application, namespace, metric, reason,
    )
    if extraText != "" {
        text = fmt.Sprintf("%s\n\n%s", text, extraText)
    }

    payload := slackMessage{
        Blocks: []slackBlock{
            {
                Type: "section",
                Text: &slackBlockText{Type: "mrkdwn", Text: "*K8S Monitoring App Alert*"},
            },
            {Type: "divider"},
        },
        Attachments: []slackAttachment{
            {
                Color:    "danger",
                Title:    title,
                Text:     text,
                MrkdwnIn: []string{"text"},
            },
        },
    }

    body, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("marshal slack payload: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
    if err != nil {
        return fmt.Errorf("create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("post to slack: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
    }
    return nil
}

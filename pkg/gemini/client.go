package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Client struct {
	model *genai.GenerativeModel
}

func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	model := client.GenerativeModel("gemini-2.0-flash-001")
    
    // Set response MIME type to JSON
    model.ResponseMIMEType = "application/json"
    
	return &Client{model: model}, nil
}

type NegotiationResponse struct {
	Decision        string `json:"decision"`
	Reasoning       string `json:"reasoning"`
	ResponseContent string `json:"response_content"`
}

func (c *Client) GenerateNegotiationResponse(ctx context.Context, itemPrice int, minPrice int, views int, content string, durationDays int) (string, string, string, error) {
	// 1. Construct Prompt
	viewsContext := "Normal"
	if views > 100 { viewsContext = "High (Popular)" }
	if views < 10 { viewsContext = "Low (Unpopular)" }
	
	urgencyContext := "Fresh Listing"
	if durationDays > 14 { urgencyContext = "Old Listing (Urgent to sell)" }

	promptText := fmt.Sprintf(`
You are a smart negotiation agent for a flea market app seller.
Your goal is to negotiate the price of a listed item.

**Item Context:**
- Listing Price: ¥%d
- Effective MAP (Minimum Acceptable Price): ¥%d (Do NOT accept below this)
- Market Popularity: %s (Views: %d)
- Listing Urgency: %s (Days Listed: %d)

**User Message:**
"%s"

**Strategy Instructions:**
1. If the user's offer is below MAP, you MUST REJECT or Counter above MAP.
2. If Popularity is High, comprise less. If Urgency is High (Old), comprise more.
3. Output format must be strictly structured using the definition below.

Please output JSON:
{
  "decision": "ACCEPT" | "REJECT" | "COUNTER",
  "reasoning": "Internal reasoning for the seller...",
  "response_content": "Message to the buyer..."
}
`, itemPrice, minPrice, viewsContext, views, urgencyContext, durationDays, content)

	// 2. Call Gemini API
    // Note: Schema definition in Generative AI Go SDK is slightly different or experimental. 
    // Usually ResponseMIMEType is enough for simple JSON.
    // We will stick to ResponseMIMEType and internal prompting for compatibility.

	resp, err := c.model.GenerateContent(ctx, genai.Text(promptText))
	if err != nil {
		return "", "", "", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", "", "", fmt.Errorf("empty response from Gemini")
	}

	// 3. Parse Response
	part := resp.Candidates[0].Content.Parts[0]
	var txt string
	if t, ok := part.(genai.Text); ok {
		txt = string(t)
	} else {
		return "", "", "", fmt.Errorf("unexpected response type")
	}

	var parsedResp NegotiationResponse
	if err := json.Unmarshal([]byte(txt), &parsedResp); err != nil {
		// Log raw text for debugging if needed
		fmt.Printf("Raw Gemini Response: %s\n", txt)
		return "", "", "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	return parsedResp.Decision, parsedResp.Reasoning, parsedResp.ResponseContent, nil
}

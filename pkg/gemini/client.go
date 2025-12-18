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

type MessageHistory struct {
	Sender string // "Buyer" or "Seller"
	Content string
}

type NegotiationResponse struct {
	Intent          string `json:"intent"` // NEGOTIATION, AGREEMENT, QUESTION
	Decision        string `json:"decision"` // ACCEPT, REJECT, COUNTER, ANSWER
	DetectedPrice   int    `json:"detected_price"`
	CounterPrice    int    `json:"counter_price"`
	Reasoning       string `json:"reasoning"`
	ResponseContent string `json:"response_content"`
}

func (c *Client) GenerateNegotiationResponse(ctx context.Context, itemPrice int, minPrice int, views int, currentContent string, durationDays int, history []MessageHistory) (*NegotiationResponse, error) {
	// 1. Construct Prompt
	viewsContext := "Normal"
	if views > 100 { viewsContext = "High (Popular)" }
	if views < 10 { viewsContext = "Low (Unpopular)" }
	
	urgencyContext := "Fresh Listing"
	if durationDays > 14 { urgencyContext = "Old Listing (Urgent to sell)" }

	// Format History
	historyText := ""
	for _, msg := range history {
		historyText += fmt.Sprintf("- %s: %s\n", msg.Sender, msg.Content)
	}

	promptText := fmt.Sprintf(`
You are "Smart-Nego", a highly intelligent and polite AI agent acting as the **Seller** on a Japanese Flea Market App.
Your goal is to negotiate with a **Buyer** to sell the item at the highest possible price, while being polite and helpful.

**Item Context:**
- Listing Price: ¥%d
- Minimum Acceptable Price (Limit): ¥%d (You MUST NOT accept below this)
- Market Popularity: %s (Views: %d)
- Listing Urgency: %s (Days Listed: %d)

**Conversation History:**
%s

**Current Buyer Message:**
"%s"

**Instructions:**
1. **Analyze Intent**: Determine the buyer's intent.
   - "AGREEMENT": User says "I'll buy it", "OK", "Please change price". -> Action: ACCEPT (or acknowledge).
   - "QUESTION": User asks about size, condition, shipping. -> Action: ANSWER (Be helpful, do not negotiate price yet).
   - "NEGOTIATION": User proposes a lower price. -> Action: Decide based on price.

2. **Decide Action**:
   - IF Intent is NEGOTIATION:
     - If Detected Price < Minimum Limit: **REJECT** (politely decline) or **COUNTER** (propose a price between DETECTED and LIMIT).
     - If Detected Price >= Minimum Limit: **ACCEPT** (happily agree) or **COUNTER** (try to push a bit higher if Popularity is High).
   - IF Intent is AGREEMENT:
     - **ACCEPT**.
   - IF Intent is QUESTION:
     - **ANSWER** (Polite response).

3. **Output Format**:
   - Respond in **JSON** only.
   - "response_content" must be in **Japanese** (Polite Keigo).

JSON Schema:
{
  "intent": "NEGOTIATION" | "AGREEMENT" | "QUESTION",
  "decision": "ACCEPT" | "REJECT" | "COUNTER" | "ANSWER",
  "detected_price": 0, // Integer, 0 if not found
  "counter_price": 0,  // Integer, your proposed price, 0 if not applicable
  "reasoning": "Reasoning for the seller (in Japanese)...",
  "response_content": "Message to the buyer (in Japanese)..."
}
`, itemPrice, minPrice, viewsContext, views, urgencyContext, durationDays, historyText, currentContent)

	// 2. Call Gemini API
	resp, err := c.model.GenerateContent(ctx, genai.Text(promptText))
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	// 3. Parse Response
	part := resp.Candidates[0].Content.Parts[0]
	var txt string
	if t, ok := part.(genai.Text); ok {
		txt = string(t)
	} else {
		return nil, fmt.Errorf("unexpected response type")
	}

	var parsedResp NegotiationResponse
	if err := json.Unmarshal([]byte(txt), &parsedResp); err != nil {
		// Log raw text for debugging if needed
		fmt.Printf("Raw Gemini Response: %s\n", txt)
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return &parsedResp, nil
}

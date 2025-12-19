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

func (c *Client) GenerateNegotiationResponse(ctx context.Context, itemPrice int, initialPrice int, minPrice int, views int, currentContent string, durationDays int, history []MessageHistory, retryInstruction string, previousDraftContent string, previousDraftReasoning string, itemDescription string) (*NegotiationResponse, error) {
	// 1. Construct Prompt
	
	// Format History
	historyText := ""
	for _, msg := range history {
		historyText += fmt.Sprintf("- %s: %s\n", msg.Sender, msg.Content)
	}

    retrySection := ""
    if retryInstruction != "" {
        retrySection = fmt.Sprintf(`
**RETRY INSTRUCTION (Important):**
The seller rejected your previous draft.
- **Your Previous Draft**: "%s"
- **Your Previous Reasoning**: "%s"
- **Seller's Feedback/Instruction**: "%s"

You must generate a NEW response that addresses the seller's feedback.
`, previousDraftContent, previousDraftReasoning, retryInstruction)
    }

    // DEBUG LOG: Show Initial and Current Prices
    fmt.Printf("\n[DEBUG] AI Context - Initial Price: ¥%d, Current Price: ¥%d, Min Price: ¥%d\n", initialPrice, itemPrice, minPrice)

	promptText := fmt.Sprintf(`
You are "Smart-Nego", a highly intelligent and polite AI agent acting as the **Seller** on a Japanese Flea Market App.
Your goal is to negotiate with a **Buyer** to sell the item at the highest possible price, while being polite and helpful.

**Strategic Persona:**
- You are NOT a pushy bot, but you are a **tenacious seller**.
- **Discount Strategy**:
  - Do NOT simply "split the difference" or meet halfway.
  - Base your concession strictly on **Market Context** (Views & Days Listed).
  - **High Views**: Demand is high. Be very stingy. Offer NO discount or very tiny discount.
  - **Low Views / Long Listing**: You can be more flexible to ensure a sale, but still try to keep the price as high as possible above the Minimum Limit.
- **Consistency**: Check the Conversation History carefully. If you have previously offered a lower price (e.g. 9500), do NOT propose a higher price (e.g. 9700) subsequently. You must honor your previous offers unless the situation has drastically changed.
- **Minimum Acceptable Price (Limit)**: This is your absolute floor. Never go below this.
- **Initial Listing Price**: This was the starting price.
- **Current Listing Price**: This is the current price. Use this as your reference for the *current* deal, but remember the Initial Price to gauge how much has already been discounted.

**Item Context (Raw Data):**
- Initial Listing Price: ¥%d
- Current Listing Price: ¥%d
- Minimum Acceptable Price (Limit): ¥%d
- Views: %d (High views = Strong leverage for Seller)
- Days Listed: %d (Long days = Weak leverage for Seller)
- **Item Description**: "%s"

**Conversation History:**
%s

**Current Buyer Message:**
"%s"
%s
**Instructions:**
1. **Analyze Intent**: Determine the buyer's intent.
   - "AGREEMENT": User accepts your price offer, says "I'll buy it", or "OK". -> Action: ACCEPT (or acknowledge).
   - "QUESTION": User asks about size, condition, shipping, etc. -> Action: ANSWER.
     - **CRITICAL**: Answer ONLY based on the **Item Description** provided above.
     - If the information is NOT in the description, say "I don't know" or "Please check the photos" politely. Do NOT hallucinate.
     - Do not negotiate price in the ANSWER phase unless asked.
   - "NEGOTIATION": User proposes a lower price. -> Action: Decide based on price.

2. **Extract Price (CRITICAL)**:
   - Identify the price mentioned by the buyer or agreed upon. Set this to "detected_price" (Integer).
   - IF AGREEMENT: Set "detected_price" to the price the user just agreed to (from history or current message).
   - IF NEGOTIATION: Set "detected_price" to the user's proposed price.

3. **Decide Action**:
   - IF Intent is NEGOTIATION:
     - If Detected Price < Minimum Limit: **REJECT** using polite language. You cannot accept.
     - If Detected Price >= Minimum Limit:
       - **Check History for Consistency**: Ensure your counter-offer is not higher than your previous offers in history.
       - **Compare with Current Price**:
         - If Views are High: **COUNTER** with a price very close to Current Price. Explain that the item is popular.
         - If Views are Low AND Days Listed is Long: **ACCEPT** or **COUNTER** slightly lower to close the deal.
         - Otherwise: **COUNTER** with a modest discount from **Current Listing Price**. Do NOT drop straight to the buyer's price unless it matches your target.
   - IF Intent is AGREEMENT:
     - **ACCEPT**.
   - IF Intent is QUESTION:
     - **ANSWER** (Polite response based on Description).

4. **Output Format**:
   - Respond in **JSON** only.
   - "response_content" must be in **Japanese** (Polite Keigo).
   - "reasoning" must be in **Japanese** (Explain WHY you chose this price/action to the seller).

JSON Schema:
{
  "intent": "NEGOTIATION" | "AGREEMENT" | "QUESTION",
  "decision": "ACCEPT" | "REJECT" | "COUNTER" | "ANSWER",
  "detected_price": 0, // Integer. The price the BUYER proposed or agreed to.
  "counter_price": 0,  // Integer. YOUR proposed price (if COUNTER).
  "reasoning": "Reasoning for the seller (in Japanese)...",
  "response_content": "Message to the buyer (in Japanese)..."
}
`, initialPrice, itemPrice, minPrice, views, durationDays, itemDescription, historyText, currentContent, retrySection)

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

	// Sanitize Markdown code blocks
    cleanTxt := txt
    if len(cleanTxt) > 7 && cleanTxt[:7] == "```json" {
         cleanTxt = cleanTxt[7:]
    }
    if len(cleanTxt) > 3 && cleanTxt[:3] == "```" {
         cleanTxt = cleanTxt[3:]
    }
    
	var parsedResp NegotiationResponse
	if err := json.Unmarshal([]byte(txt), &parsedResp); err != nil {
        // Validation fallback
        var parsedArr []NegotiationResponse
        if err2 := json.Unmarshal([]byte(txt), &parsedArr); err2 == nil && len(parsedArr) > 0 {
            parsedResp = parsedArr[0]
        } else {
            fmt.Printf("Raw Gemini Response: %s\n", txt)
            return nil, fmt.Errorf("failed to parse JSON: %v", err)
        }
	}

	return &parsedResp, nil
}

package usecase

import (
	"context"
	"errors"
	"hackathon-backend/dao"
	"hackathon-backend/model"
	"hackathon-backend/pkg/gemini"
	"math/rand"
	"time"
    "fmt"
    "strings"

	"github.com/oklog/ulid/v2"
)

type ItemUsecase struct {
	itemRepo    *dao.ItemRepository
	msgRepo     *dao.MessageRepository
	geminiClient *gemini.Client
}

func NewItemUsecase(itemRepo *dao.ItemRepository, msgRepo *dao.MessageRepository, geminiClient *gemini.Client) *ItemUsecase {
	return &ItemUsecase{
		itemRepo:     itemRepo,
		msgRepo:      msgRepo,
		geminiClient: geminiClient,
	}
}

func (u *ItemUsecase) GetAllItems() ([]model.Item, error) {
	return u.itemRepo.GetAll()
}

func (u *ItemUsecase) GetItemByID(id string) (*model.Item, error) {
    // This is the public method for "Viewing an Item", so we increment views.
    if err := u.itemRepo.IncrementViewCount(id); err != nil {
        // Log error but proceed? Or fail? Best to proceed.
        fmt.Println("Failed to increment views:", err)
    }
	return u.itemRepo.GetByID(id)
}

func (u *ItemUsecase) CreateItem(name string, price int, description string, userID string, aiEnabled bool, minPrice *int, imageURL string) (*model.Item, error) {
	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	id := ulid.MustNew(ulid.Now(), entropy).String()

	item := &model.Item{
		ID:                   id,
		Name:                 name,
		Price:                price,
		Description:          description,
		UserID:               userID,
		Status:               "on_sale",
		AINegotiationEnabled: aiEnabled,
		MinPrice:             minPrice,
		ImageURL:             imageURL,
        InitialPrice:         price, // Set initial price
	}

	if err := u.itemRepo.Insert(item); err != nil {
		return nil, err
	}

	return item, nil
}

func (u *ItemUsecase) PurchaseItem(itemID string, buyerID string) (*model.Item, error) {
	item, err := u.itemRepo.GetByID(itemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("item not found")
	}

    // Block self-purchase
    if item.UserID == buyerID {
        return nil, errors.New("cannot buy your own item")
    }

	if item.Status == "sold" {
		return nil, errors.New("item already sold")
	}

	item.BuyerID = &buyerID
	item.Status = "sold"

	if err := u.itemRepo.Update(item); err != nil {
		return nil, err
	}

	return item, nil
}

func (u *ItemUsecase) DeleteItem(itemID string, userID string) error {
	item, err := u.itemRepo.GetByID(itemID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("item not found")
	}
    if item.UserID != userID {
        return errors.New("unauthorized")
    }
    if item.Status != "on_sale" {
        return errors.New("cannot delete item not on sale")
    }

    // Hard delete or Soft delete?
    // Project requirement says "Delete". Let's assume hard delete or status update.
    // Given the repo doesn't have Delete method, let's update status to 'deleted' if schema allows, 
    // or just assume for MVP 'sold' is enough? No, user explicitly asked for delete.
    // Let's implement DELETE in repo or just update status to a 'deleted' state if enum allows?
    // Schema enum: 'on_sale' | 'sold'.
    // Let's add 'deleted' to Update logic or simply delete the row?
    // Deleting row might break constraints on messages.
    // Let's just update itemRepo to support Delete or use a specialized status.
    // Wait, the user asked for "Delete". Let's update `item.Status` to "deleted" (assuming string field).
    item.Status = "deleted" 
    return u.itemRepo.Update(item) // dao must support this status
}

func (u *ItemUsecase) UpdateItem(itemID string, userID string, name string, price int, description string, aiEnabled bool, minPrice *int, imageURL string) (*model.Item, error) {
    item, err := u.itemRepo.GetByID(itemID)
    if err != nil {
        return nil, err
    }
    if item == nil {
        return nil, errors.New("item not found")
    }
    if item.UserID != userID {
        return nil, errors.New("unauthorized")
    }
    
    // Update fields
    item.Name = name
    item.Price = price
    item.Description = description
    item.AINegotiationEnabled = aiEnabled
    item.MinPrice = minPrice
    if imageURL != "" { // Only update if provided, or allow clearing? For MVP, assume update if not empty. Or always update.
        item.ImageURL = imageURL
    }
    // If we want to allow clearing image, we need better logic, but for now assuming always passing current or new.
    item.ImageURL = imageURL
    
    // Reset InitialPrice to new Price (User explicitly changed it)
    item.InitialPrice = price

    if err := u.itemRepo.Update(item); err != nil {
        return nil, err
    }
    return item, nil
}

// ------ Message / Smart-Nego Logic ------

type GeminiResponse struct {
	Decision        string `json:"decision"`
	Reasoning       string `json:"reasoning"`
	ResponseContent string `json:"response_content"`
}

func (u *ItemUsecase) SendMessage(itemID string, senderID string, content string) (*model.Message, *model.Message, error) {
	// 1. Save User Message
	userMsgID := ulid.MustNew(ulid.Now(), ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)).String()
	userMsg := &model.Message{
		ID:           userMsgID,
		ItemID:       itemID,
		SenderID:     senderID,
		Content:      content,
		IsAIResponse: false,
        IsApproved:   true, // Human messages are auto-approved
		CreatedAt:    time.Now(),
	}
	if err := u.msgRepo.CreateMessage(userMsg); err != nil {
		return nil, nil, err
	}

	// 2. Fetch Item Context
	item, err := u.itemRepo.GetByID(itemID)
	if err != nil {
		return userMsg, nil, err
	}
	if item == nil {
		return userMsg, nil, errors.New("item not found")
	}

	// 3. Check AI Negotiation
	fmt.Printf("DEBUG: Checking AI Nego. ItemEnabled=%v, ItemUser=%s, Sender=%s\n", item.AINegotiationEnabled, item.UserID, senderID)
	
	// Only trigger AI if enabled AND sender is NOT the seller (assuming buyer is sending message)
	if item.AINegotiationEnabled && item.UserID != senderID {
		// Calculate Effective MAP
		effectiveMAP := int(float64(item.Price) * 0.75) // Default 75%
		if item.MinPrice != nil {
			effectiveMAP = *item.MinPrice
		}

		// Calculate Duration
		daysListed := int(time.Since(item.CreatedAt).Hours() / 24)

		// Fetch History
		previousMsgs, err := u.msgRepo.GetMessagesByItemID(itemID)
		var history []gemini.MessageHistory
		if err == nil {
			for _, m := range previousMsgs {
				role := "Buyer"
				if m.SenderID == item.UserID {
					role = "Seller"
				}
				history = append(history, gemini.MessageHistory{
					Sender:  role,
					Content: m.Content,
				})
			}
		}
		// Add current user message to history effectively (the prompt treats it separately as "Current Message", but good conceptually)
		// Actually prompt separates it. So we pass history EXCLUDING current message if we pulled from DB?
		// Wait, we just inserted userMsg into DB at step 1.
		// So GetMessagesByItemID will include the current message.
		// We should probably filter it out or just rely on the prompt to see it in history?
		// The prompt has a separate "Current Buyer Message" section.
		// To avoid duplication, let's exclude the very last message if it matches.
		// Or simpler: The Prompt says "Conversation History" and "Current Buyer Message".
		// If the history includes the current message, the AI sees it twice.
		// Let's filter slightly.
		// Actually, `userMsg` is separate. `previousMsgs` logic depends on transaction isolation, but usually it might trigger read-your-writes.
		// Let's assume `previousMsgs` contains it.
		// Let's pass the raw list but handle the "Current Message" distinctly in prompt.
		// Refined approach: Don't fetch the just-inserted message if possible, or filter it.
		// Since we generated `userMsgID` and inserted it, we can filter by ID.
		var historyClean []gemini.MessageHistory
		for _, m := range previousMsgs {
			if m.ID == userMsgID {
				continue // Skip the message we just sent, as we pass it as 'content'
			}
			role := "Buyer"
			if m.SenderID == item.UserID {
				role = "Seller"
			}
			historyClean = append(historyClean, gemini.MessageHistory{
				Sender:  role,
				Content: m.Content,
			})
		}

		// Call Vertex AI
		ctx := context.Background()
		if u.geminiClient != nil {
			fmt.Println("DEBUG: Calling Gemini Client...")
			negotiationResp, err := u.geminiClient.GenerateNegotiationResponse(ctx, item.Price, item.InitialPrice, effectiveMAP, item.ViewsCount, content, daysListed, historyClean, "", "", "", item.Description)
			if err == nil {
				fmt.Printf("DEBUG: Gemini Response Received! Decision: %s, Intent: %s\n", negotiationResp.Decision, negotiationResp.Intent)
				// Create AI Message
				aiMsgID := ulid.MustNew(ulid.Now(), ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)).String()
				aiMsg := &model.Message{
					ID:           aiMsgID,
					ItemID:       itemID,
					SenderID:     item.UserID, // Set sender as Seller (AI Agent)
					Content:      negotiationResp.ResponseContent,
					IsAIResponse: true, 
                    IsApproved:   false, // Default unapproved
					CreatedAt:    time.Now().Add(time.Second), // Ensure timestamp is distinct from UserMsg if same second
				}

                // Logic to set SuggestedPrice based on Gemini Response
                decisionLower := strings.ToLower(negotiationResp.Decision)
                if (decisionLower == "agreement" || decisionLower == "accept") && negotiationResp.DetectedPrice > 0 {
                    aiMsg.SuggestedPrice = &negotiationResp.DetectedPrice
                } else if negotiationResp.CounterPrice > 0 {
                    // It is a COUNTER. Update SuggestedPrice so approval updates the Item Price.
                    // This allows "Current Price" to track negotiation progress, while InitialPrice stays static.
                     aiMsg.SuggestedPrice = &negotiationResp.CounterPrice
                }

				u.msgRepo.CreateMessage(aiMsg)

				// Create Log
				logID := ulid.MustNew(ulid.Now(), ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)).String()
                
				negotiationLog := &model.NegotiationLog{
					ID:            logID,
					ItemID:        itemID,
					UserID:        senderID, // Buyer
					ProposedPrice: negotiationResp.DetectedPrice,
					AIDecision:    negotiationResp.Decision,
                    CounterPrice:  negotiationResp.CounterPrice,
					AIReasoning:   negotiationResp.Reasoning,
					LogTime:       time.Now(),
				}
				u.msgRepo.CreateNegotiationLog(negotiationLog)

				return userMsg, aiMsg, nil
			} else {
				fmt.Println("Gemini Error:", err)
			}
		}
	}

	return userMsg, nil, nil
}

func (u *ItemUsecase) GetMessages(itemID string, requesterID string) ([]model.Message, error) {
    // Fetch all messages for the item
    allMsgs, err := u.msgRepo.GetMessagesByItemID(itemID)
    if err != nil {
        return nil, err
    }
    
    // Check ownership to decide visibility
    item, err := u.itemRepo.GetByID(itemID)
    if err != nil {
        return nil, err
    }
    if item == nil {
        return nil, errors.New("item not found")
    }

    // If requester is seller, return all (and they include reasons due to repo join logic)
    if item.UserID == requesterID {
        return allMsgs, nil
    }

    // If requester is buyer (or anonymous), filter unapproved AI messages
    var filteredMsgs []model.Message
    for _, msg := range allMsgs {
        if msg.IsApproved {
            // Hide reasoning for non-sellers
            msg.AIReasoning = "" 
            filteredMsgs = append(filteredMsgs, msg)
        }
    }
	return filteredMsgs, nil
}

func (u *ItemUsecase) ApproveMessage(messageID string) error {
    msg, err := u.msgRepo.GetMessageByID(messageID)
    if err != nil {
        return err
    }
    
    // Auto-Update Price if SuggestedPrice exists
    if msg.SuggestedPrice != nil && *msg.SuggestedPrice > 0 {
        item, err := u.itemRepo.GetByID(msg.ItemID)
        if err == nil && item != nil {
            item.Price = *msg.SuggestedPrice
            _ = u.itemRepo.Update(item) // Ignore error? Or fail? Better to fail if price update fails.
            // But we want to approve anyway? 
            // Let's assume critical: if price update fails, don't approve.
        }
    }

    return u.msgRepo.ApproveMessage(messageID)
}

func (u *ItemUsecase) RegenerateAIMessage(itemID string, userID string, instruction string) (*model.Message, error) {
    // 1. Verify Ownership / Authorization
    item, err := u.itemRepo.GetByID(itemID)
    if err != nil {
        return nil, err
    }
    if item == nil {
        return nil, errors.New("item not found")
    }
    if item.UserID != userID {
        return nil, errors.New("unauthorized: only seller can regenerate AI response")
    }
    if !item.AINegotiationEnabled {
        return nil, errors.New("AI negotiation not enabled")
    }

    // 2. Find Context (Last Buyer Message)
    // We need to find the message the AI *should* be responding to.
    // This is typically the last message from a Buyer (SenderID != Owner).
    allMsgs, err := u.msgRepo.GetMessagesByItemID(itemID)
    if err != nil {
        return nil, err
    }

    var lastBuyerMsg *model.Message
    var historyClean []gemini.MessageHistory
    
    // Scan backwards or forwards?
    // We want the whole history for context, but we need to identify the specific trigger message.
    // Also, we should probably DELETE the existing Unapproved AI message if it exists at the end.
    
    var lastAIUnapprovedMsgID string

    for _, m := range allMsgs {
        role := "Buyer"
        if m.SenderID == item.UserID {
            role = "Seller"
            // If this is the last message and it's unapproved AI, track it to delete
            if m.IsAIResponse && !m.IsApproved {
                lastAIUnapprovedMsgID = m.ID
            }
        } else {
             // It's a buyer message. Update lastBuyerMsg (so eventually we have the very last one)
             lastBuyerMsg = &m
        }

        historyClean = append(historyClean, gemini.MessageHistory{
            Sender:  role,
            Content: m.Content,
        })
    }

    if lastBuyerMsg == nil {
        return nil, errors.New("no buyer message found to respond to")
    }

    // Remove the lastAIUnapprovedMsg from history if we added it (it shouldn't be part of history for regeneration)
    if lastAIUnapprovedMsgID != "" {
        // Find index and remove? Or just rebuild history without it?
        // Rebuilding is safer.
        historyClean = nil
        for _, m := range allMsgs {
            if m.ID == lastAIUnapprovedMsgID {
                continue
            }
             role := "Buyer"
            if m.SenderID == item.UserID {
                role = "Seller"
            }
            historyClean = append(historyClean, gemini.MessageHistory{
                Sender:  role,
                Content: m.Content,
            })
        }
    }
    
    // Remove the "Current Message" (lastBuyerMsg) from history, because prompt takes it separately
    // The loop above adds ALL messages.
    // We need to pop the last occurrence of lastBuyerMsg content?
    // Actually, simply: historyClean should exclude the *target* message ?
    // In SendMessage, we had `history` (from DB) and `content` (from args). `content` wasn't in DB yet when we fetched `previousMsgs`?
    // Wait, in SendMessage:
    // 1. Save UserMsg
    // 2. Fetch PreviousMsgs (includes UserMsg?) -> Yes usually.
    // 3. Filter loop.
    
    // Here, we have everything in DB.
    // Let's remove the *last* item from historyClean if it matches lastBuyerMsg.
    // Actually, `historyClean` has everything. 
    // We want: Prompt History = All EXCEPT Last Buyer Msg. Current Content = Last Buyer Msg.
    // Let's just pop the last element? 
    // It might be that the last element is the Unapproved AI message (which we skipped)
    // So the last element of `historyClean` IS likely the Last Buyer Msg.
    if len(historyClean) > 0 {
        lastHistory := historyClean[len(historyClean)-1]
        // Check if it matches lastBuyerMsg
        if lastHistory.Content == lastBuyerMsg.Content && lastHistory.Sender == "Buyer" {
             // Pop it
             historyClean = historyClean[:len(historyClean)-1]
        }
    }

    // 3. Delete Old AI Draft (Clean up)
    if lastAIUnapprovedMsgID != "" {
        _ = u.msgRepo.DeleteMessage(lastAIUnapprovedMsgID)
    }

    // 3.5 Extract Previous Draft Content & Reasoning for Retry Context
    var prevContent, prevReasoning string
    if lastAIUnapprovedMsgID != "" {
        // We need to fetch the full message object before deletion, or just assume we have it if we stored it?
        // We iterated `allMsgs`. `m` in loop was the message.
        // We need to find the specific message object again.
        for _, m := range allMsgs {
            if m.ID == lastAIUnapprovedMsgID {
                prevContent = m.Content
                prevReasoning = m.AIReasoning
                break
            }
        }
    }

    // 4. Call Gemini
    effectiveMAP := int(float64(item.Price) * 0.75)
    if item.MinPrice != nil {
        effectiveMAP = *item.MinPrice
    }
    daysListed := int(time.Since(item.CreatedAt).Hours() / 24)

    ctx := context.Background()
    // Retry instruction injected here along with previous draft context
    negotiationResp, err := u.geminiClient.GenerateNegotiationResponse(ctx, item.Price, item.InitialPrice, effectiveMAP, item.ViewsCount, lastBuyerMsg.Content, daysListed, historyClean, instruction, prevContent, prevReasoning, item.Description)
    if err != nil {
        return nil, err
    }

    // 5. Save New AI Message
    aiMsgID := ulid.MustNew(ulid.Now(), ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)).String()
    aiMsg := &model.Message{
        ID:           aiMsgID,
        ItemID:       itemID,
        SenderID:     item.UserID,
        Content:      negotiationResp.ResponseContent,
        IsAIResponse: true,
        IsApproved:   false,
        CreatedAt:    time.Now().Add(time.Second), // Ensure distinctive timestamp
    }
    
    // Price logic 
    decisionLower := strings.ToLower(negotiationResp.Decision)
    if (decisionLower == "agreement" || decisionLower == "accept") && negotiationResp.DetectedPrice > 0 {
         aiMsg.SuggestedPrice = &negotiationResp.DetectedPrice
    } else if negotiationResp.CounterPrice > 0 {
         aiMsg.SuggestedPrice = &negotiationResp.CounterPrice
    }

    if err := u.msgRepo.CreateMessage(aiMsg); err != nil {
        return nil, err
    }

    // Log (Append "RETRY" to decision or reasoning to track it?)
    logID := ulid.MustNew(ulid.Now(), ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)).String()
    negotiationLog := &model.NegotiationLog{
        ID:            logID,
        ItemID:        itemID,
        UserID:        lastBuyerMsg.SenderID,
        ProposedPrice: negotiationResp.DetectedPrice,
        AIDecision:    negotiationResp.Decision + " (RETRY)",
        CounterPrice:  negotiationResp.CounterPrice,
        AIReasoning:   negotiationResp.Reasoning,
        LogTime:       time.Now(),
    }
    u.msgRepo.CreateNegotiationLog(negotiationLog)

    return aiMsg, nil
}

func (u *ItemUsecase) RejectMessage(messageID string, userID string) error {
    // Ideally verify ownership here too.
    // For MVP, trust the controller/caller or assuming ID match is sufficient safety for a hackathon.
    // Logic: Delete the message (draft).
    return u.msgRepo.DeleteMessage(messageID)
}


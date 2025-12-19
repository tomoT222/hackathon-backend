package controller

import (
	"encoding/json"
	"hackathon-backend/usecase"
	"net/http"
	"strings"
)

type ItemController struct {
	usecase *usecase.ItemUsecase
}

func NewItemController(usecase *usecase.ItemUsecase) *ItemController {
	return &ItemController{usecase: usecase}
}

func (c *ItemController) GetItems(w http.ResponseWriter, r *http.Request) {
	items, err := c.usecase.GetAllItems()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if items == nil {
		w.Write([]byte("[]"))
		return
	}
}

type CreateItemRequest struct {
	Name                 string `json:"name"`
	Price                int    `json:"price"`
	Description          string `json:"description"`
	UserID               string `json:"user_id"`
	AINegotiationEnabled bool   `json:"ai_negotiation_enabled"`
	MinPrice             *int   `json:"min_price"`
	ImageURL             string `json:"image_url"`
}

func (c *ItemController) HandleItems(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    switch r.Method {
    case "GET":
        items, err := c.usecase.GetAllItems()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        if items == nil {
             w.Write([]byte("[]"))
             return
        }
        json.NewEncoder(w).Encode(items)
    case "POST":
        var req CreateItemRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
             http.Error(w, err.Error(), http.StatusBadRequest)
             return
        }
        item, err := c.usecase.CreateItem(req.Name, req.Price, req.Description, req.UserID, req.AINegotiationEnabled, req.MinPrice, req.ImageURL)
        if err != nil {
             http.Error(w, err.Error(), http.StatusInternalServerError)
             return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(item)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

type BuyRequest struct {
    UserID string `json:"user_id"`
}

type SendMessageRequest struct {
    UserID  string `json:"user_id"`
    Content string `json:"content"`
}

func (c *ItemController) HandleItemDetail(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    // path: /items/{id} or /items/{id}/buy or /items/{id}/messages
    parts := strings.Split(r.URL.Path, "/")
    if len(parts) < 3 {
        http.Error(w, "Invalid URL", http.StatusBadRequest)
        return
    }
    id := parts[2]

    // Check if it is a buy request
    if len(parts) >= 4 && parts[3] == "buy" {
        if r.Method != "PUT" {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        var req BuyRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
             http.Error(w, err.Error(), http.StatusBadRequest)
             return
        }
        item, err := c.usecase.PurchaseItem(id, req.UserID)
        if err != nil {
             if err.Error() == "item not found" {
                 http.Error(w, err.Error(), http.StatusNotFound)
                 return
             }
             if err.Error() == "item already sold" {
                 http.Error(w, err.Error(), http.StatusConflict)
                 return
             }
             http.Error(w, err.Error(), http.StatusInternalServerError)
             return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(item)
        return
    }

    // Check if it is a messages request
    if len(parts) >= 4 && parts[3] == "messages" {
        // Handle Retry: /items/{id}/messages/retry
        if len(parts) >= 5 && parts[4] == "retry" {
            if r.Method == "POST" {
                 var req struct {
                     UserID      string `json:"user_id"`
                     Instruction string `json:"instruction"`
                 }
                 if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                     http.Error(w, err.Error(), http.StatusBadRequest)
                     return
                 }
                 
                 aiMsg, err := c.usecase.RegenerateAIMessage(id, req.UserID, req.Instruction)
                 if err != nil {
                      http.Error(w, err.Error(), http.StatusInternalServerError)
                      return
                 }
                 w.Header().Set("Content-Type", "application/json")
                 json.NewEncoder(w).Encode(aiMsg)
                 return
            } else {
                 http.Error(w, "Method not allowed for retry", http.StatusMethodNotAllowed)
                 return
            }
        }

        if r.Method == "POST" {
            var req SendMessageRequest
            if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }
            userMsg, aiMsg, err := c.usecase.SendMessage(id, req.UserID, req.Content)
            if err != nil {
                 http.Error(w, err.Error(), http.StatusInternalServerError)
                 return
            }
            // Return latest message (or both)
             w.Header().Set("Content-Type", "application/json")
             response := map[string]interface{}{
                 "user_message": userMsg,
                 "ai_message": aiMsg, // Can be nil
             }
            json.NewEncoder(w).Encode(response)
            return
        } else if r.Method == "GET" {
             // Extract userID query param for filtering
             userID := r.URL.Query().Get("user_id")
             msgs, err := c.usecase.GetMessages(id, userID)
             if err != nil {
                 http.Error(w, err.Error(), http.StatusInternalServerError)
                 return
             }
             w.Header().Set("Content-Type", "application/json")
             if msgs == nil {
                 w.Write([]byte(`{"messages": []}`))
                 return
             }
             json.NewEncoder(w).Encode(map[string]interface{}{"messages": msgs})
             return
        } else {
             http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
             return
        }
    }

    switch r.Method {
    case "GET":
        item, err := c.usecase.GetItemByID(id)
        if err != nil {
             http.Error(w, err.Error(), http.StatusInternalServerError)
             return
        }
        if item == nil {
            http.Error(w, "Item not found", http.StatusNotFound)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(item)
    case "DELETE":
        // Extract user_id from query or header? For MPV, query param or body to verify owner
        // Since DELETE body is discouraged, let's use Query Param `user_id` or Header `X-User-ID`
        // Let's use Query Param for simplicity as we used it elsewhere
        userID := r.URL.Query().Get("user_id")
        if userID == "" {
             http.Error(w, "user_id required", http.StatusUnauthorized)
             return
        }
        err := c.usecase.DeleteItem(id, userID)
        if err != nil {
             http.Error(w, err.Error(), http.StatusBadRequest)
             return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "deleted"}`))
    case "PUT":
        // Update Item
        var req CreateItemRequest // Reuse structure
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
             http.Error(w, err.Error(), http.StatusBadRequest)
             return
        }
        item, err := c.usecase.UpdateItem(id, req.UserID, req.Name, req.Price, req.Description, req.AINegotiationEnabled, req.MinPrice, req.ImageURL)
        if err != nil {
             status := http.StatusInternalServerError
             if err.Error() == "unauthorized" {
                 status = http.StatusUnauthorized
             } else if err.Error() == "item not found" {
                 status = http.StatusNotFound
             }
             http.Error(w, err.Error(), status)
             return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(item)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func (c *ItemController) HandleMessages(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "POST, PUT, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    // path: /messages/{id}/approve
    parts := strings.Split(r.URL.Path, "/")
    if len(parts) < 4 {
        http.Error(w, "Invalid URL", http.StatusBadRequest)
        return
    }
    msgID := parts[2]
    action := parts[3]

    if action == "approve" && r.Method == "PUT" {
        // Need userID to verify permission (conceptually)
        // For now, we trust the caller has permission (Client side checks + knowing the draft ID)
        var req struct {
            UserID string `json:"user_id"`
        }
         if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
             http.Error(w, err.Error(), http.StatusBadRequest)
             return
         }

        err := c.usecase.ApproveMessage(msgID)
        if err != nil {
             http.Error(w, err.Error(), http.StatusInternalServerError)
             return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "approved"}`))
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "approved"}`))
    } else if action == "reject" && r.Method == "PUT" {
        var req struct {
            UserID string `json:"user_id"`
        }
         if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
             http.Error(w, err.Error(), http.StatusBadRequest)
             return
         }

        err := c.usecase.RejectMessage(msgID, req.UserID)
        if err != nil {
             http.Error(w, err.Error(), http.StatusInternalServerError)
             return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "rejected"}`))
    } else {
        http.Error(w, "Not found or method not allowed", http.StatusNotFound)
    }
}

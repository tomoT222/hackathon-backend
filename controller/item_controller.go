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
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Description string `json:"description"`
	UserID      string `json:"user_id"`
}

func (c *ItemController) HandleItems(w http.ResponseWriter, r *http.Request) {
    // CORS headers
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
        item, err := c.usecase.CreateItem(req.Name, req.Price, req.Description, req.UserID)
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

func (c *ItemController) HandleItemDetail(w http.ResponseWriter, r *http.Request) {
    // CORS headers
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    // path: /items/{id} or /items/{id}/buy
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
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

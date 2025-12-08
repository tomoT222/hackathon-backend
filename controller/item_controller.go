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
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
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

    // path: /items/{id}
    parts := strings.Split(r.URL.Path, "/")
    if len(parts) < 3 {
        http.Error(w, "Invalid URL", http.StatusBadRequest)
        return
    }
    id := parts[len(parts)-1] 

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

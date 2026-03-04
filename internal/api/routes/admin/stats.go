package admin

import (
	"encoding/json"
	"net/http"

	"github.com/akramboussanni/gocode/internal/api"
)

// HandleGetSystemStats returns overview stats for the admin dashboard
func (ar *AdminRouter) HandleGetSystemStats(w http.ResponseWriter, r *http.Request) {
	stats := make(map[string]interface{})

	// Total users
	totalUsers, err := ar.Repos.User.CountTotalUsers(r.Context())
	if err == nil {
		stats["total_users"] = totalUsers
	}

	// Active users (last 7 days)
	activeUsers, err := ar.Repos.User.CountActiveUsers(r.Context(), 7)
	if err == nil {
		stats["active_users_7d"] = activeUsers
	}

	api.WriteJSON(w, 200, stats)
}

// HandleListDecks returns all decks for admin selection
func (ar *AdminRouter) HandleListDecks(w http.ResponseWriter, r *http.Request) {
	decks, err := ar.Repos.Deck.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Failed to list decks", http.StatusInternalServerError)
		return
	}

	// Enrich with category counts
	type DeckWithCounts struct {
		ID            int64  `json:"id,string"`
		DeckKey       string `json:"deck_key"`
		Title         string `json:"title"`
		Description   string `json:"description"`
		DeckType      string `json:"deck_type"`
		CategoryCount int    `json:"category_count"`
		IsSystem      bool   `json:"is_system"`
	}

	var result []DeckWithCounts
	for _, deck := range decks {
		catCount, _ := ar.Repos.Category.CountByDeckID(r.Context(), deck.ID)
		result = append(result, DeckWithCounts{
			ID:            deck.ID,
			DeckKey:       deck.DeckKey,
			Title:         deck.Title,
			Description:   deck.Description,
			DeckType:      deck.DeckType,
			CategoryCount: catCount,
			IsSystem:      deck.IsSystem,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

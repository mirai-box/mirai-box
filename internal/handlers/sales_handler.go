package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mirai-box/mirai-box/internal/middleware"
)

func (h *SaleHandler) MySales(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserIDFromContext(r.Context())
	sales, err := h.saleService.FindByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to retrieve sales", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(sales)
}


package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

// getClientConsultation returns finalized CR(s) for a visit to the pet owner.
func (a *API) getClientConsultation(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	visitID := chi.URLParam(r, "visitID")
	visit, err := a.store.GetVisit(r.Context(), visitID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	pet, err := a.store.GetPet(r.Context(), visit.PetID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return
	}
	if pet.OwnerUserID != id.UserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
		return
	}
	active, err := a.store.HasActiveEntitlement(r.Context(), pet.ID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if !active {
		writeErr(w, r, http.StatusPaymentRequired, "payment_required", "pet_inactive")
		return
	}

	reports, err := a.store.ListFinalVisitReportsForClient(r.Context(), visitID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if len(reports) == 0 {
		writeErr(w, r, http.StatusNotFound, "not_found", "consultation_not_found")
		return
	}

	practiceName := ""
	if visit.PracticeID != "" {
		if name, err := a.store.GetPracticeName(r.Context(), visit.PracticeID); err == nil {
			practiceName = name
		}
	}

	scheduledAt := ""
	if visit.ScheduledAt != nil {
		scheduledAt = visit.ScheduledAt.UTC().Format(time.RFC3339)
	}

	httpx.WriteData(w, http.StatusOK, map[string]any{
		"visitId":      visit.ID,
		"petId":        pet.ID,
		"petName":      pet.Name,
		"practiceName": practiceName,
		"status":       visit.Status,
		"scheduledAt":  scheduledAt,
		"createdAt":    visit.CreatedAt.UTC().Format(time.RFC3339),
		"reports":      reports,
	})
}

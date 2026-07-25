package handlers

import (
	"net/http"

	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
)

func (a *API) listVetPets(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "pets.read")
	if !ok {
		return
	}
	pets, err := a.store.ListPetsForPractice(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, pets)
}

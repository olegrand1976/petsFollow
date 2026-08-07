package handlers

import (
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
)

// registerPprofRoutes expose /internal/debug/pprof/* pour diagnostiquer une
// instance Cloud Run en place (profil tas, goroutines bloquées, CPU) plutôt que
// de reproduire une fuite en local.
//
// Ces profils exposent des noms de symboles et des piles : ils restent fermés
// tant que PPROF_SECRET est vide, et le header se compare en temps constant
// comme les autres endpoints internes.
func (a *API) registerPprofRoutes(r chi.Router) {
	if a.cfg.PprofSecret == "" {
		return
	}
	r.Route("/internal/debug/pprof", func(pr chi.Router) {
		pr.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if !secretHeaderOK(req, "X-Pprof-Secret", a.cfg.PprofSecret) {
					writeErr(w, req, http.StatusUnauthorized, "unauthorized", "unauthorized")
					return
				}
				next.ServeHTTP(w, req)
			})
		})
		pr.HandleFunc("/", pprof.Index)
		pr.HandleFunc("/cmdline", pprof.Cmdline)
		pr.HandleFunc("/profile", pprof.Profile)
		pr.HandleFunc("/symbol", pprof.Symbol)
		pr.HandleFunc("/trace", pprof.Trace)
		// pprof.Index ne reconnaît un profil nommé que sous /debug/pprof/ :
		// monté ailleurs il renverrait sa page d'index. On réécrit le chemin.
		pr.Get("/{name}", func(w http.ResponseWriter, req *http.Request) {
			req.URL.Path = "/debug/pprof/" + chi.URLParam(req, "name")
			pprof.Index(w, req)
		})
	})
}

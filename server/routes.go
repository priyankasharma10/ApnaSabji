package server

import (
	"github.com/go-chi/chi"
)

// Update InjectRoutes to use the modified srv.register
func (srv *Server) InjectRoutes() *chi.Mux {
	r := chi.NewRouter()

	// r.Get("/health", srv.HealthCheck)
	r.Route("/api", func(api chi.Router) {
		api.Post("/register", srv.register) // Use Post method for POST requests\
		api.Post("/login", srv.loginWithEmailOTP)

		// Other API routes can be added here if needed
	})

	return r
}

package providers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"github.com/vijaygniit/ApnaSabji/models"
)

type PSQLProvider interface {
	// DB returns the database client.
	DB() *sqlx.DB
}

type DBProvider interface {
	// Ping verifies the connection with the database.
	Ping() error
	PSQLProvider
}

type MiddlewareProvider interface {
	Middleware() func(next http.Handler) http.Handler
	UserFromContext(ctx context.Context) *models.UserContextData

	// Default has default middleware written on the top levels of router such as CORS.
	Default() chi.Middlewares
	//SuperAdminCheck() chi.Middlewares
}

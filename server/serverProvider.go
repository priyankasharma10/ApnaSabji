package server

import (
	"context"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/vijaygniit/ApnaSabji/providers"
	dbprovider "github.com/vijaygniit/ApnaSabji/providers/dbProvider"
	"github.com/vijaygniit/ApnaSabji/providers/dbhelperprovider"
	"github.com/vijaygniit/ApnaSabji/providers/middlewareprovider"
)

type Server struct {
	MiddlewareProvider providers.MiddlewareProvider
	DBHelper           providers.DBHelperProvider
	PSQL               providers.PSQLProvider
	httpServer         *http.Server
}

func SrvInit() *Server {

	db := dbprovider.NewPSQLProvider(os.Getenv("ELEPHANTSQL_URL"))

	// database helper functions
	dbHelper := dbhelperprovider.NewDBHepler(db.DB())

	middleware := middlewareprovider.NewMiddleware(dbHelper)

	return &Server{
		PSQL:               db,
		DBHelper:           dbHelper,
		MiddlewareProvider: middleware,
	}
}

func (srv *Server) Start() {
	addr := ":" + os.Getenv("server_port")

	httpSrv := &http.Server{
		Addr:    addr,
		Handler: srv.InjectRoutes(),
	}
	srv.httpServer = httpSrv

	logrus.Info("Server running at PORT ", addr)
	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Fatalf("Start %v", err)
		return
	}
}

func (srv *Server) Stop() {
	logrus.Info("closing Postgres...")
	_ = srv.PSQL.DB().Close()
	//_ = srv.PSQLC.DB().Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logrus.Info("closing server...")
	_ = srv.httpServer.Shutdown(ctx)
	logrus.Info("Done")
}

package dbprovider

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vijaygniit/ApnaSabji/providers"

	//"websocket/database"
	"github.com/sirupsen/logrus"
)

type psqlProvider struct {
	db *sqlx.DB
}

func NewPSQLProvider(connectionString string) providers.DBProvider {
	var (
		db          *sqlx.DB
		err         error
		maxAttempts = 3
	)

	for i := 0; i < maxAttempts; i++ {
		db, err = sqlx.Connect("postgres", connectionString)
		if err != nil {
			logrus.Errorf("unable to connect to postgres PSQL %v, connection string := %v", err, connectionString)
			time.Sleep(3 * time.Second)
			continue
		}
		break
	}

	if err != nil {
		logrus.Fatalf("Failed to initialize PSQL: %v", err)
	} else {
		logrus.Info("connected to postgresql database")
		// database.Migrations(db)  // Commented out the migration call
	}

	return &psqlProvider{
		db: db,
	}
}

func (pp *psqlProvider) Ping() error {
	return pp.db.Ping()
}

func (pp *psqlProvider) DB() *sqlx.DB {
	return pp.db
}

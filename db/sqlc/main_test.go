package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"simplebank/util"
	"testing"
)

var testStore Store

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	connPool, err := pgxpool.New(context.Background(), config.DBUri)
	if err != nil {
		log.Fatal("cannot connect to database:", err)
	}

	testStore = NewStore(connPool)

	os.Exit(m.Run())
}

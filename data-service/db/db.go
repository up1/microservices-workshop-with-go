package db

import (
	"context"
	"database/sql"
	"demo/cmd"
	"demo/model"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"github.com/sirupsen/logrus"
	"github.com/up1/microservices-workshop-with-go/common/tracing"
)

type DbClient interface{
	InquiryAccount(ctx context.Context, accountId string) (model.AccountData, error)
    CreateConnection(addr string)
	Close()
}

type PostgresClient struct{
	db *sql.DB
}

func NewPostgresClient(cfg *cmd.Config) *PostgresClient {
	pc := &PostgresClient{}
    pc.CreateConnection(cfg.PostgresUrl)
	return pc
}

func (pc *PostgresClient) InquiryAccount(ctx context.Context, accountId string) (model.AccountData, error) {
	span := tracing.StartChildSpanFromContext(ctx, "PostgresClient.InquiryAccount")
	defer span.Finish()

	if pc.db == nil {
		return model.AccountData{}, fmt.Errorf("Connection to DB not established!")
	}

	var account model.AccountData
	sql := "SELECT id, name FROM accounts WHERE id = $1"
	err := pc.db.QueryRow(sql, accountId).Scan(&account.ID, &account.Name)
	if err != nil {
		log.Fatal("Failed to execute query: ", err)
	}
	
	return account, nil
}

func (pc *PostgresClient) CreateConnection(addr string) {
	logrus.Infof("Connecting with connection string: '%v'", addr)
    var err error
	pc.db, err = sql.Open("postgres", addr)
	if err != nil {
		log.Fatal("Failed to open a DB connection: ", err)
	}
	err = pc.db.Ping()
    if err != nil {
        log.Fatal("Failed to ping DB: ", err)
    }
	logrus.Info("Successfully connected to DB")
}

func (pc *PostgresClient) Close() {
	pc.db.Close()
}

package repositories

import (
	"context"
	"database/sql"
	"log"

	"gitlab.com/wgroup1/enigma/interfaces"
)

type clientRepo struct{
	DB *sql.DB
}

func NewClientRepository(db *sql.DB) interfaces.ClientRepository {
	return &clientRepo{
		DB: db,
	}
}

func (cli *clientRepo) GetClientID(clientCode string) int {
	db := cli.DB
	ctx := context.Background()
	cID := 0
	sqlQuery := "select id from client where client_code = ? and status = 1"
	err := db.QueryRowContext(ctx, sqlQuery, clientCode).Scan(&cID)
	if err != nil {
		log.Println("error occured when get client:", err.Error())
		return 0
	}

	return cID
}

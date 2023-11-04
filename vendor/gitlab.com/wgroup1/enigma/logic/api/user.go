package api

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"gitlab.com/wgroup1/enigma/common"
	"gitlab.com/wgroup1/enigma/structs"
	"gopkg.in/go-playground/validator.v9"
)

type userLogic struct {
	DB *sql.DB
}

type UserLogic interface {
	CreateUserLogin(w http.ResponseWriter, r *http.Request)
}

func NewUserLogic(db *sql.DB) UserLogic {
	return &userLogic{
		DB: db,
	}
}

//CreateUsers is the func for creating multiple user(tutor, student) at once
func (user*userLogic) CreateUsers(w http.ResponseWriter, r *http.Request) {
	//create user, user role
	w.Header().Set("Content-Type", "application/json")
}

//CreateUserLogin is the func for creating the username/email that will be used for logging in
func (user *userLogic) CreateUserLogin(w http.ResponseWriter, r *http.Request) {
	//create username/email with hashed default password: DOB - YYYYMMDD
	w.Header().Set("Content-Type", "application/json")

	var usr []structs.User
	var errstr structs.ErrorMessage

	_ = json.NewDecoder(r.Body).Decode(&usr)

	db := user.DB
	ctx := context.Background()
	// Generate 16 random bytes
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	sqlQuery := "INSERT INTO users (username, password, client_id, hash_key) values (?,?,?,?)"
	sqlQueryRole := "INSERT INTO user_roles (user_id, role_id) values (?,?)"
	tx, err := db.Begin()
	j := 0
	for range usr {
		if err != nil {
			tx.Rollback()
			// defer db.Close()
			common.JSONError(w, usr[j].Username, err.Error(), http.StatusInternalServerError)
			return
		}

		v := validator.New()
		err := v.Struct(usr[j])
		if err != nil {
			tx.Rollback()
			// defer db.Close()
			common.JSONError(w, structs.Validate, err.Error(), http.StatusInternalServerError)
			return
		}

		res, err := tx.ExecContext(ctx, sqlQuery, &usr[j].Username, hashAndSalt([]byte(usr[j].Password)), &usr[j].ClientID, randomBytes)
		if err != nil {
			tx.Rollback()
			// defer db.Close()
			common.JSONError(w, structs.QueryErr, err.Error(), http.StatusInternalServerError)
			return
		}

		lastInsertedID, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			// defer db.Close()
			common.JSONError(w, structs.LastIDErr, err.Error(), http.StatusInternalServerError)
			return
		}

		if &usr[j].RoleID != nil {
			_, err := tx.ExecContext(ctx, sqlQueryRole, &lastInsertedID, &usr[j].RoleID)
			if err != nil {
				tx.Rollback()
				// defer db.Close()
				common.JSONError(w, structs.QueryErr, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		log.Println("Last inserted ID:", lastInsertedID)

		errstr.Message = structs.Success
		errstr.Code = http.StatusOK
		j++
	}
	tx.Commit()
	// defer db.Close()
	json.NewEncoder(w).Encode(errstr)
}

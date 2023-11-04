package api

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gitlab.com/wgroup1/enigma/common"
	"gitlab.com/wgroup1/enigma/database"
	"gitlab.com/wgroup1/enigma/repositories/mysql"

	"gitlab.com/wgroup1/enigma/structs"
	"golang.org/x/crypto/bcrypt"
)

type authLogic struct{
	DB *sql.DB
	JWTKEY []byte
}

type AuthLogic interface {
	Login(w http.ResponseWriter, r *http.Request)
}

// @todo need to put in env
var expMins time.Duration = 2628000
var myKey = []byte(mysql.ViperEnvVariable("JWT_KEY"))
var SparkAuth string = ""


func NewAuthLogic(db *sql.DB) AuthLogic {
	return &authLogic{
		DB: db,
	}
}

// Login and generate the JWT token
func (auth *authLogic) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var ath structs.Auth
	var rol structs.Role
	var usr structs.User
	var passw string

	_ = json.NewDecoder(r.Body).Decode(&usr)

	db := auth.DB
	ctx := context.Background()

	sqlQuery := "SELECT id, username, password, hash_key FROM users where username = ?"
	err := db.QueryRowContext(ctx, sqlQuery, usr.Username).Scan(&ath.Data.UserID, &ath.Data.Username, &passw, &ath.Data.HashKey)

	if err != nil {
		common.JSONError(w, structs.UserNotFound, err.Error(), http.StatusInternalServerError)
		return
	}
	if comparePasswords(passw, []byte(usr.Password)) {

		//get client data
		sqlQueryClient := "select id, client_name, status from client where id = (select client_id from users u2 where username = ?)"
		err2 := db.QueryRowContext(ctx, sqlQueryClient, usr.Username).Scan(&ath.Client.ClientID, &ath.Client.ClientName,  &ath.Client.Status)

		if err2 != nil {
			common.JSONError(w, structs.QueryErr, err.Error(), http.StatusInternalServerError)
			return
		}

		if ath.Data.HashKey == "" {
			// Generate 16 random bytes
			randomBytes := make([]byte, 16)
			_, err := rand.Read(randomBytes)
			if err != nil {
				common.JSONError(w, structs.QueryErr, err.Error(), http.StatusInternalServerError)
				return
			}

			// Convert random bytes to hex string
			hexString := hex.EncodeToString(randomBytes)

			sqlClient := "UPDATE users SET hash_key = ? WHERE id = ?"
			resp, clientErr := db.ExecContext(ctx, sqlClient, hexString, ath.Data.UserID)

			if clientErr != nil {
				common.JSONError(w, structs.QueryErr, err.Error(), http.StatusInternalServerError)
				return
			}

			rowsAffected, _ := resp.RowsAffected()

			if rowsAffected == 0 {
				common.JSONError(w, structs.QueryErr, err.Error(), http.StatusInternalServerError)
				return
			}

			ath.Data.HashKey = hexString

		}

		tokenString, err := GenerateJWT(ath.Data.HashKey)
		if err != nil {
			common.JSONError(w, structs.GenTokenErr, err.Error(), http.StatusInternalServerError)
			return
		}

		//get user roles
		sqlQueryRole := "select ro.id, ro.name from user_roles ur inner join roles ro on ur.role_id = ro.id and ur.user_id = ?"
		res2, err := db.QueryContext(ctx, sqlQueryRole, ath.Data.UserID)
		defer database.CloseRows(res2)
		if err != nil {
			common.JSONError(w, structs.QueryErr, err.Error(), http.StatusInternalServerError)
			// defer db.Close()
			return
		}
		for res2.Next() {
			res2.Scan(&rol.RoleID, &rol.RoleName)
			ath.Roles = append(ath.Roles, rol)
		}

		ath.Message = structs.Success
		ath.Data.AccessToken = tokenString
		ath.Data.ExpireAt = time.Now().Add(time.Minute * expMins)
		// ath.Data.HashKey = "4f013cd60bb80ca04b4229d360a0e8e4"

		json.NewEncoder(w).Encode(ath)
	} else {
		common.JSONError(w, structs.IncorrectPass, "", http.StatusInternalServerError)
		// defer db.Close()
		return
	}
	// defer db.Close()
}

// IsAuthorized is the func for validating the JWT token
func IsAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Authorization"] != nil {
			token, err := jwt.Parse(r.Header["Authorization"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("there was an error")
				}

				return "", nil
			})


			if err != nil {
				// if client use static token
				if r.Header["Authorization"][0] == mysql.ViperEnvVariable("STATIC_TOKEN") {
					r.Header.Add("X-Hash-Key", r.Header["Authorization"][0])
					endpoint(w, r)
				} else {
					common.JSONError(w, structs.Unauthorized, err.Error(), http.StatusUnauthorized)
					return
				}
				// common.JSONError(w, structs.Unauthorized, err.Error(), http.StatusUnauthorized)
				// return
			} else {
				if token.Valid {
					r.Header.Add("X-Hash-Key", token.Claims.(jwt.MapClaims)["hash_key"].(string))
					endpoint(w, r)
				}
			} 


			// old auth process

			// if err != nil {
			// 	common.JSONError(w, structs.Unauthorized, err.Error(), http.StatusUnauthorized)
			// 	return
			// }

			// r.Header.Add("X-Hash-Key", token.Claims.(jwt.MapClaims)["hash_key"].(string))

			// if token.Valid {
			// 	endpoint(w, r)
			// }

		} else {
			common.JSONError(w, structs.AuthNotFound, "", http.StatusUnauthorized)
			return
		}
	})
}

// IsStaticAuth is the func for validating static token
func IsStaticAuth(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Authorization"] != nil {
			if r.Header["Authorization"][0] == mysql.ViperEnvVariable("STATIC_TOKEN") {
				endpoint(w, r)
			} else {
				common.JSONError(w, structs.AuthNotFound, "", http.StatusUnauthorized)
				return
			}
		} else {
			common.JSONError(w, structs.AuthNotFound, "", http.StatusUnauthorized)
			return
		}
	})
}

// GenerateJWT is func to generate the token
func GenerateJWT(hashKey string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	//claims["user"] = "Erlangga"
	//@todo need to define the expired time
	claims["exp"] = time.Now().Add(time.Minute * expMins).Unix()
	claims["hash_key"] = hashKey

	tokenString, err := token.SignedString(myKey)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return tokenString, nil
}

func hashAndSalt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		fmt.Println(err)
	}

	return string(hash)
}

func comparePasswords(hashedPwd string, plainPwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func BasicAuthMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		//fmt.Println("username: ", user)
		//fmt.Println("password: ", pass)
		if !ok || !checkUsernameAndPassword(user, pass) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Please enter your username and password for this site"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}
		handler(w, r)
	}
}

func checkUsernameAndPassword(username, password string) bool {
	if SparkAuth == "" {
		SparkAuth = mysql.ViperEnvVariable("SPARK_AUTH")

	}

	return (username + " " + password) == SparkAuth
	//return username == "dev.wgroup@gmail.com" && password == "wGroup_1tdev"
}

/*
func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
*/

package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gitlab.com/wgroup1/enigma/common"
	"gitlab.com/wgroup1/enigma/structs"
)

// @todo need to put in env
var expMins time.Duration = 2628000
// var myKey = []byte(GetViperEnvVariable("JWT_KEY"))
var SparkAuth string = ""

// GenerateJWT is func to generate the token
func GenerateJWT(hashKey string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	jwtKey := []byte(GetViperEnvVariable("JWT_KEY"))

	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["exp"] = time.Now().Add(time.Minute * expMins).Unix()
	claims["hash_key"] = hashKey

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return tokenString, nil
}

// IsAuthorized is the func for validating the JWT token
func IsAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Authorization"] != nil {
			jwtKey := []byte(GetViperEnvVariable("JWT_KEY"))
			token, err := jwt.Parse(r.Header["Authorization"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("there was an error")
				}

				return jwtKey, nil
			})


			if err != nil {
				// if client use static token
				if r.Header["Authorization"][0] == GetViperEnvVariable("STATIC_TOKEN") {
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
			if r.Header["Authorization"][0] == GetViperEnvVariable("STATIC_TOKEN") {
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
		SparkAuth = GetViperEnvVariable("SPARK_AUTH")

	}

	return (username + " " + password) == SparkAuth
	//return username == "dev.wgroup@gmail.com" && password == "wGroup_1tdev"
}
package common

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"gitlab.com/wgroup1/enigma/structs"
	"gopkg.in/go-playground/validator.v9"
)

const REDIS_DB int = 0
const OUTBOUND_QUEUE string = "enigma-queue-outbound"
const OUTBOUND_QUEUE_BULK string = "enigma-queue-outbound-bulk"
const FLOW_QUEUE string = "enigma-queue-flow"
const SPARKPOST_QUEUE_REPORT string = "sparkpost-queue-report"
const CONVERSATION_QUEUE string = "enigma-queue-conversation"
const DAMCORP_INBOUND_WA string = "damcorp-queue-inbound"

// JSONError is func to return JSON error format
func JSONError(w http.ResponseWriter, message string, sysMessage string, code int) {
	var errstr structs.ErrorMessage
	errstr.Message = message
	errstr.SysMessage = sysMessage
	errstr.Code = code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errstr)
}

func JSONErr(w http.ResponseWriter, errStr *structs.ErrorMessage) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errStr.Code)
	json.NewEncoder(w).Encode(errStr)
}

func JSONErrs(w http.ResponseWriter, errStr *[]structs.ErrorMessage) {
	w.Header().Set("Content-Type", "application/json")
	code := (*errStr)[0].Code
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errStr)
}

func JSONErrsSparkPost(w http.ResponseWriter, errStr *[]structs.ErrorMessageSparkPost) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(errStr)
}

func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func HitAPI(url string, jsonStr string, method string, strToken string, alias string, timeout time.Duration, DB *sql.DB) (*http.Request, *http.Response, []byte, int, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader([]byte(jsonStr)))

	if err != nil {
		fmt.Println("error when hit URL:", url, "- err:", err.Error())
	} else {
		req.Close = true
		req.Header.Add("Content-Type", "application/json")
	}

	if strToken != "" {
		req.Header.Add("Authorization", strToken) // INI defaultnya mas angga sebelum di ubah
	}

	switch alias {
	case "MTARGET":
		req.Header.Del("Authorization")
	case "JTS":
		//something special for Jatis Header Auth
	case "TEST_VENDOR":
		//something special for Jatis Test Header Auth
	case "MSGBRD":
		//something special for MessageBird Header Auth
	case "SLEEKFLOW":
		req.Header.Del("Authorization")
	}

	tr := &http.Transport{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 500,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	//client := &http.Client{Transport: tr}
	client := &http.Client{Transport: tr, Timeout: timeout * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// time.Sleep(time.Second * 10)
		fmt.Println("Error when hit to API:", err.Error())
		return req, resp, nil, 0, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	respBody := body

	if alias == "DamcorpMedia" {
		respBody = []byte("")
	}


	//commented because of locking table when a lot of traffic comming

	db := DB
	ctx := context.Background()
	tx, err := db.Begin()
	if err != nil {
		log.Println("error when open conn:", err.Error())
		return req, resp, body, 0, nil
	}
	sqlQuery := "insert into api_logs (url, method, request_header, request_body, response_status, response_header, response_body) values (?, ?, ?, ?, ?, ?, ?)"
	res, err2 := tx.ExecContext(ctx, sqlQuery, url, method, req.Header.Get("Authorization"), jsonStr, resp.StatusCode, resp.Header.Get("From"), string(respBody))
	if err2 != nil {
		tx.Rollback()
		// defer db.Close()
		log.Println("error when insert into api_logs:", err2.Error())
	}

	lastID, err3 := res.LastInsertId()
	lastInsID := int(lastID)
	_ = lastInsID
	if err3 != nil {
		tx.Rollback()
		// defer db.Close()
		log.Println("error when get insertID:", err3.Error())
	}

	tx.Commit()
	// defer db.Close()

	return req, resp, body, 0, nil
}

// func HitAPICallBack(url string, jsonStr string, method string, strToken string, alias string, timeout time.Duration) (*http.Request, *http.Response, []byte, int, error) {
func HitAPICallBack(cbauth structs.CallBackAuth, jsonStr string, timeout time.Duration, DB *sql.DB) (*http.Request, *http.Response, []byte, int, error) {
	req, err := http.NewRequest(cbauth.CallBackMethod, cbauth.CallBackUrl, bytes.NewReader([]byte(jsonStr)))

	if err != nil {
		fmt.Println("error when hit URL:", cbauth.CallBackUrl, "- err:", err.Error())
	} else {
		req.Close = true
		req.Header.Add("Content-Type", "application/json")
	}

	if cbauth.AuthHeader != "" {
		if cbauth.AuthHeader != "" {
			req.Header.Add(cbauth.AuthHeader, cbauth.AuthHeaderValue)
		}
	}

	switch strings.ToUpper(cbauth.AuthType) {
	//case "NOAUTH":

	case "APIKEY":
		req.Header.Add(cbauth.ApiKey, cbauth.ApiKeyValue)
		//req.Header.Del("Authorization")
	case "BEARERTOKEN ":
		req.Header.Del("Authorization")
		req.Header.Add("Authorization", "Bearer"+cbauth.AuthToken)
	case "BASICAUTH":
		req.Header.Del("Authorization")
		sbscauth := base64.StdEncoding.EncodeToString([]byte(cbauth.AuthUserName + ":" + cbauth.AuthPassword))
		req.Header.Add("Authorization", "basicauth"+sbscauth)
	case "MSGBRD":
		//something special for MessageBird Header Auth
	}

	tr := &http.Transport{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 500,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	//client := &http.Client{Transport: tr}
	client := &http.Client{Transport: tr, Timeout: timeout * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// time.Sleep(time.Second * 10)
		fmt.Println("Error when hit to API:", err.Error())
		return req, resp, nil, 0, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	//commented because of locking table when a lot of traffic comming

	db := DB
	ctx := context.Background()
	tx, err := db.Begin()
	if err != nil {
		log.Println("error when open conn:", err.Error())
		return req, resp, body, 0, nil
	}
	sqlQuery := "insert into api_logs (url, method, request_header, request_body, response_status, response_header, response_body) values (?, ?, ?, ?, ?, ?, ?)"
	res, err2 := tx.ExecContext(ctx, sqlQuery, cbauth.CallBackUrl, cbauth.CallBackMethod, req.Header.Get("Authorization"), jsonStr, resp.StatusCode, resp.Header.Get("From"), string(body))
	if err2 != nil {
		tx.Rollback()
		// defer db.Close()
		log.Println("error when insert into api_logs:", err2.Error())
	}

	lastID, err3 := res.LastInsertId()
	lastInsID := int(lastID)
	_ = lastInsID
	if err3 != nil {
		tx.Rollback()
		// defer db.Close()
		log.Println("error when get insertID:", err3.Error())
	}

	tx.Commit()
	// defer db.Close()

	return req, resp, body, 0, nil
}

func HitAPIBulk(url string, jsonStr string, method string, strToken string, alias string, timeout time.Duration) (structs.HTTPRequest, error) {
	
	req, err := http.NewRequest(method, url, bytes.NewReader([]byte(jsonStr)))

	if err != nil {
		fmt.Println("error when hit URL:", url, "- err:", err.Error())
	} else {
		req.Close = true
		req.Header.Add("Content-Type", "application/json")
	}

	if strToken != "" {
		req.Header.Add("Authorization", strToken) // INI defaultnya mas angga sebelum di ubah
	}

	switch alias {
	case "MTARGET":
		req.Header.Del("Authorization")
	case "JTS":
		//something special for Jatis Header Auth
	case "TEST_VENDOR":
		//something special for Jatis Test Header Auth
	case "MSGBRD":
		//something special for MessageBird Header Auth
	}

	tr := &http.Transport{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 500,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	//client := &http.Client{Transport: tr}
	client := &http.Client{Transport: tr, Timeout: timeout * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error when hit to API:", err.Error())
		return structs.HTTPRequest{
			Url: url,
			Method: method,
			RequestHeader: req.Header.Get("Authorization"),
			ResponseStatus: http.StatusInternalServerError,
			ResponseHeader: "",
			ResponseBody: []byte(err.Error()),
		}, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)


	response := structs.HTTPRequest{
		Url: url,
		Method: method,
		RequestHeader: req.Header.Get("Authorization"),
		ResponseStatus: resp.StatusCode,
		ResponseHeader: resp.Header.Get("From"),
		ResponseBody: body,
	}

	return response, nil
}

func SetPageLimit(page string, limit string) string {
	var offset int
	if page == "" {
		return " limit 100 offset 0"
	}
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return ""
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return ""
	}

	offset = (pageInt - 1) * limitInt
	ret := fmt.Sprintf(" limit %d offset %d", limitInt, offset)
	return ret

}

func IsEmailAddressFormat(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func RandomString(charExclude string, codeLength int, statictext string, statictext_pos int, allowDuplicateChar int) string {

	//var lettersAll = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	const lettersAll = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	//var lettersUpperOnly = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	//const lettersUpperNumericOnly = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	//var lettersLowerOnly = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	//var lettersLowerNumericOnly = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	letters := lettersAll
	if statictext != "" {
		codeLength = codeLength - len(statictext)
	}

	s := make([]byte, codeLength)
	dedupchar := make(map[string]interface{})

	for iLoopTotalLength := range s {
		if allowDuplicateChar == 1 {
			s[iLoopTotalLength] = letters[seededRand.Intn(len(letters))]
		} else {
			sCharChek := letters[seededRand.Intn(len(letters))]
			for {
				//fmt.Println(string(s))
				if value, exist := dedupchar[string(sCharChek)]; exist {
					//fmt.Println("Key found value is: ", value)
					_ = value
					sCharChek = letters[seededRand.Intn(len(letters))]
				} else {
					//fmt.Println("Key not found")
					dedupchar[string(sCharChek)] = string(sCharChek)
					s[iLoopTotalLength] = sCharChek
					break
				}
			}
		}
	}
	stringS := string(s)
	if statictext != "" {
		if statictext_pos == 2 {
			stringS = stringS + statictext
		} else {
			stringS = statictext + stringS
		}

	}

	return stringS
}

func Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func EncryptAES(text string, secret []byte) (string, error) {
	// MySecret := Encode(secret[:aes.BlockSize])
	MySecret := secret
	iv := []byte(secret[:aes.BlockSize])
	block, err := aes.NewCipher([]byte(MySecret))
	if err != nil {
		return "", err
	}
	plainText := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, iv)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	return Encode(cipherText), nil
}

func DecryptAES(text string, secret []byte) (string, error) {
	// MySecret := Encode(secret[:aes.BlockSize])
	MySecret := secret
	iv := []byte(secret[:aes.BlockSize])
	block, err := aes.NewCipher([]byte(MySecret))
	if err != nil {
		return "", err
	}
	cipherText := Decode(text)
	cfb := cipher.NewCFBDecrypter(block, iv)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	return string(plainText), nil
}

func ValidateStruct(s interface{}) error {
    validate := validator.New()

    if err := validate.Struct(s); err != nil {
        if _, ok := err.(*validator.InvalidValidationError); ok {
            return err
        }

        var validationErrors []string
        for _, err := range err.(validator.ValidationErrors) {
            fieldName := err.Field()
			errType := err.Tag()
			dataType := err.Kind()
			errParams := err.Param()
            validationErrors = append(validationErrors, fmt.Sprintf("Field: '%s' type: %s %s %s", fieldName, dataType, errType, errParams))
        }

		errResult := strings.Join(validationErrors, ", ")

        return fmt.Errorf("Validation errors: [ %v ]", errResult)
    }

    return nil
}


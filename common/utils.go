package common

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	redisMass "github.com/garyburd/redigo/redis"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gitlab.com/wgroup1/enigma-consumer/structs"
	"gitlab.com/wgroup1/enigma/repositories/mysql"
)

var API_IN_ID int = 0

const REDIS_DB int = 0
const OUTBOUND_QUEUE string = "enigma-queue-outbound"
const OUTBOUND_QUEUE_BULK string = "enigma-queue-outbound-bulk"
const FLOW_QUEUE string = "enigma-queue-flow"
const SPARKPOST_QUEUE_REPORT string = "sparkpost-queue-report"
const CONVERSATION_QUEUE string = "enigma-queue-conversation"
const DAMCORP_INBOUND_WA string = "damcorp-queue-inbound"

const ROUTINES = 10

var wg sync.WaitGroup
var pool *redisMass.Pool

var DBUrl string = ""

func GetMD5HashWithSum(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

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
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(errStr)
}

func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func HitAPI(url string, jsonStr []byte, method string, strToken string, timeout time.Duration) (*http.Request, *http.Response, []byte, int, error) {
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

	tr := &http.Transport{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 500,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: timeout * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Second * timeout)
		fmt.Println("error when hit URL:", url, "- err:", err.Error())
		db := mysql.InitializeMySQL()
		tx, err2 := db.Begin()
		if err2 != nil {
			fmt.Println("error when hit URL:", url, "- err:", err2.Error())
			return req, resp, nil, 0, err2
		}
		sqlQuery := "insert into api_logs (url, method, request_header,request_body, response_status, response_header, response_body) values (?, ?, ?, ?, ?, ?)"
		res, err3 := tx.Exec(sqlQuery, url, method, req.Header, jsonStr, 502, "", err.Error())
		if err3 != nil {
			tx.Rollback()
			fmt.Println("error when hit URL:", url, "- err:", err3.Error())
		}

		lastID, err3 := res.LastInsertId()
		lastInsID := int(lastID)
		if err3 != nil {
			tx.Rollback()
			fmt.Println("error when get insertID:", err3.Error())
		}

		tx.Commit()

		defer db.Close()

		return req, resp, nil, lastInsID, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	//log.Println("request Body:", string(body))

	db := mysql.InitializeMySQL()
	tx, err := db.Begin()
	if err != nil {
		log.Println("error when open conn:", err.Error())
		return req, resp, body, 0, nil
	}
	sqlQuery := "insert into api_logs (url, method,request_header, request_body, response_status, response_header, response_body) values (?, ?, ?, ?, ?, ?, ?)"
	rspHeader := ""
	if resp.Header != nil {
		rspHeader = resp.Header.Get("From")
	}
	res, err2 := tx.Exec(sqlQuery, url, method, req.Header.Get("Authorization"), jsonStr, resp.StatusCode, rspHeader, string(body))
	if err2 != nil {
		tx.Rollback()
		log.Println("error when insert into api_logs:", err2.Error())
	}

	lastID, err3 := res.LastInsertId()
	lastInsID := int(lastID)
	if err3 != nil {
		tx.Rollback()
		log.Println("error when get insertID:", err3.Error())
	}

	tx.Commit()

	defer db.Close()

	return req, resp, body, lastInsID, nil
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

func MySqlCustomDateTimeFormat(dateString string, isBegin int) string {
	var dateTimeString string
	if isBegin == 1 {
		dateTimeString = dateString + " 00:00:00.000000"
	} else {
		dateTimeString = dateString + " 23:59:59.000000"
	}
	_, err := time.Parse("2006-01-02 15:04:05", dateTimeString)
	if err != nil {
		dateTimeString = ""
	}
	return dateTimeString
}

// ViperEnvVariable is func to get .env file
func ViperEnvVariable(key string) string {
	switch key {
	case "DB_URL":
		if DBUrl != "" {
			return DBUrl
		}
	case "REDIS_URL":
		if RedisUrl != "" {
			return RedisUrl
		}
	case "REDIS_URL_PRIMARY":
		if RedisUrlPrimary != "" {
			return RedisUrlPrimary
		}
	case "REDIS_URL_REPLICA":
		if RedisUrlReplica != "" {
			return RedisUrlReplica
		}
	}

	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()

	if err != nil {
		log.Fatalf("Error while reading config file %s", err)
	}

	value, ok := viper.Get(key).(string)

	if !ok {
		log.Fatalf("Utils - Invalid type assertion")
	}

	switch key {
	case "DB_URL":
		DBUrl = value
	case "REDIS_URL":
		RedisUrl = value
	case "REDIS_URL_PRIMARY":
		RedisUrlPrimary = value
	case "REDIS_URL_REPLICA":
		RedisUrlReplica = value
	}

	return value
}

// START REDIIIISSSSSSSSSSSSSSSSSSSSS

var ctx = context.Background()
var RedisUrl string = ""
var RedisUrlPrimary string = ""
var RedisUrlReplica string = ""
var envMode string = "on-premmise"

func GetHost(r *http.Request) string {
	if r.URL.IsAbs() {
		host := r.Host
		// Slice off any port information.
		if i := strings.Index(host, ":"); i != -1 {
			host = host[:i]
		}
		return host
	}
	return r.URL.Host
}

func gcpSetRedisHost() string {
	redisHost := os.Getenv("REDISHOST")
	redisPort := os.Getenv("REDISPORT")
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	return redisAddr
}

func InitializeRedis(db int) *redis.Client {
	host := ViperEnvVariable("REDIS_URL_PRIMARY")
	pass := ViperEnvVariable("REDIS_PASSWORD")
	if os.Getenv("ENV_MODE") != "" {
		envMode = os.Getenv("ENV_MODE")
	}

	if envMode == "cloud" {
		host = gcpSetRedisHost()
		if os.Getenv("REDIS_PASSWORD") != "" {
			pass = os.Getenv("REDIS_PASSWORD")
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: pass,
		DB:       db,
	})

	return client
}

func Set(key string, value interface{}, db int) error {
	client := InitializeRedis(db)

	jsonValue, err := json.Marshal(value)

	if err != nil {
		fmt.Println("error when marshalling the value")
		fmt.Println(err.Error())
		return err
	}
	//SetMass(key, jsonValue)

	err2 := client.Set(ctx, key, jsonValue, 0).Err()
	if err2 != nil {
		fmt.Println("error when set key and/or value")
		fmt.Println(err2.Error())
		return err2
	}

	return nil
}

// MASS
func initializeRedisMass() {
	host := ViperEnvVariable("REDIS_URL_PRIMARY")

	pool = &redisMass.Pool{
		MaxIdle:         16,
		MaxActive:       16,
		IdleTimeout:     60 * time.Second,
		MaxConnLifetime: 5 * time.Minute,
		Dial: func() (redisMass.Conn, error) {
			c, err := net.Dial("tcp", host)
			if err != nil {
				fmt.Println("error when dialing:", err.Error())
				return nil, err
			}
			return redisMass.NewConn(c, 60*time.Second, 60*time.Second), nil
		},
		TestOnBorrow: func(c redisMass.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				fmt.Println("error when PING:", err.Error())
			}
			return err
		},
	}
}

func SetMass(valueStruct interface{}, cnt int) {
	runtime.GOMAXPROCS(ROUTINES)
	initializeRedisMass()
	massImport(valueStruct, cnt)
}

func massImport(valueStruct interface{}, cnt int) {
	wg.Add(ROUTINES)
	for i := 0; i < ROUTINES; i++ {
		go importRoutine(i, valueStruct, cnt, pool.Get())
	}
	wg.Wait()
	closePool()
}

func importRoutine(t int, valueStruct interface{}, cnt int, client redisMass.Conn) {
	defer wg.Done()
	defer client.Flush()
	portion := cnt / ROUTINES
	loop := (t + 1) * portion
	mPortion := cnt % ROUTINES
	if t+1 == ROUTINES {
		loop = loop + mPortion
	}

	for i := (t * portion); i < loop; i++ {
		var jsonValue []byte
		var err error
		var key string
		_ = err
		client.Send("SET", key, jsonValue)
	}

}

func closePool() {
	pool.Close()
}

func Get(key string, db int) (string, error) {
	client := InitializeRedis(db)

	//can be used to get all of the match pattern
	itt := client.Keys(ctx, key)
	if len(itt.Val()) == 0 {
		//fmt.Println(key, "does not exist inside redis")
		return "", nil
	}

	//valarr[0] because the result could be more than 1 row
	val, err := client.Get(ctx, itt.Val()[0]).Result()
	if err == redis.Nil {
		//fmt.Println(key, "does not exist inside db:", key)
		return "", nil
	}
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	return val, nil
}

func GetKeyValue(key string, db int) (string, error) {
	client := InitializeRedis(db)

	valarr, _ := client.Scan(ctx, 0, key, 0).Val()

	if len(valarr) == 0 {
		//fmt.Println(key, "does not exist inside redis(len):", key)
		return "", nil
	}

	return valarr[0], nil
}

func Delete(key string, db int) error {
	client := InitializeRedis(db)

	err := client.Del(ctx, key).Err()
	if err != nil {
		fmt.Println("error when deleting the key:", key, "|err:", err.Error())
		return err
	}

	client.Close()
	fmt.Println("deleted:", key)
	return nil
}

// END REDIIIISSSSSSSSSSSSSSSSSSSSS

package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"sync"
	"time"

	redisMass "github.com/gomodule/redigo/redis"

	"github.com/go-redis/redis/v8"
	"gitlab.com/wgroup1/enigma/common"
	"gitlab.com/wgroup1/enigma/middleware"
)

var RedisUrl string = ""
var RedisUrlPrimary string = ""
var RedisUrlReplica string = ""
var envMode string = "on-premmise"
var ctx = context.Background()
var wg sync.WaitGroup
var pool *redisMass.Pool

const routines = 10

func gcpSetRedisHost() string {
	redisHost := os.Getenv("REDISHOST")
    redisPort := os.Getenv("REDISPORT")
    redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	return redisAddr
}

//InitializeRedis is the func to open the connection for MySQL
func initializeRedis(db int) *redis.Client {
	host := middleware.GetViperEnvVariable("REDIS_URL_PRIMARY")
	pass := middleware.GetViperEnvVariable("REDIS_PASSWORD")
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

func initializeRedisReplica(db int) *redis.Client {
	host := middleware.GetViperEnvVariable("REDIS_URL_REPLICA")
	pass := middleware.GetViperEnvVariable("REDIS_PASSWORD")

	if os.Getenv("ENV_MODE") != "" {
		envMode = os.Getenv("ENV_MODE")
	}

	if envMode == "cloud" {
		host = gcpSetRedisHost()
		pass = os.Getenv("REDIS_PASSWORD")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: pass,
		DB:       db,
	})

	return client
}

//InitializeRedis is the func to open the connection for MySQL
func initializeRedisMass() {
	host := middleware.GetViperEnvVariable("REDIS_URL_PRIMARY")

	if os.Getenv("ENV_MODE") != "" {
		envMode = os.Getenv("ENV_MODE")
	}

	if envMode == "cloud" {
		host = gcpSetRedisHost()
	}

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

//Mass set
func SetMass(valueStruct interface{}, cnt int) {
	start := time.Now()
	runtime.GOMAXPROCS(routines)
	initializeRedisMass()
	massImport(valueStruct, cnt)
	elapsed := time.Since(start)
	fmt.Println("Took ", elapsed)
}

func massImport(valueStruct interface{}, cnt int) {
	wg.Add(routines)
	for i := 0; i < routines; i++ {
		go importRoutine(i, valueStruct, cnt, pool.Get())
	}
	wg.Wait()
	time.Sleep(5 * time.Second)
	closePool()
}

func importRoutine(t int, valueStruct interface{}, cnt int, client redisMass.Conn) {
	defer wg.Done()
	defer client.Flush()
	portion := cnt / routines
	loop := (t + 1) * portion
	mPortion := cnt % routines
	if t+1 == routines {
		loop = loop + mPortion
	}

	for i := (t * portion); i < loop; i++ {
		var jsonValue []byte
		var err error
		_ = err
		var key string
		/*
			switch v := valueStruct.(type) {
			case []structs.Customer:
				jsonValue, err = json.Marshal(v[i])
				if err != nil {
					fmt.Println("error when marshalling the value")
					fmt.Println(err.Error())
				}
				client.Send("SET", v[i].RdsKeyID, jsonValue)
				key = v[i].RdsKey
			case []structs.UserBalance:
				jsonValue, err = json.Marshal(v[i])
				if err != nil {
					fmt.Println("error when marshalling the value")
					fmt.Println(err.Error())
				}
				key = v[i].RdsKey
			case []structs.UniqueCodeExisting:
				jsonValue, err = json.Marshal(v[i])
				if err != nil {
					fmt.Println("error when marshalling the value")
					fmt.Println(err.Error())
				}

				key = v[i].RdsKey
			}
		*/

		err2 := client.Send("SET", key, jsonValue)
		if err2 != nil {
			fmt.Println("error when send SET:", err2.Error(), ", when:", key)
		}
	}

}

func closePool() {
	pool.Close()
}

//set is a func to set key and/or value inside redis,
//please note that value must be json format
func Set(key string, value interface{}, db int) error {
	client := initializeRedis(db)

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

func Delete(key string, db int) error {
	client := initializeRedis(db)

	err := client.Del(ctx, key).Err()
	if err != nil {
		fmt.Println("error when deleting the key:", key, "|err:", err.Error())
		return err
	}

	client.Close()
	fmt.Println("deleted:", key)
	return nil
}

//Get is a func to get a key inside particular db
func Get(key string, db int) (string, string, error) {
	client := initializeRedisReplica(db)

	//can be used to get all of the match pattern
	// itt := client.Keys(ctx, key)
	// if len(itt.Val()) == 0 {
	// 	fmt.Println(key, "does not exist inside redis (keys)")
	// 	if itt.Err() != nil {
	// 		fmt.Println("error when get keys on redis:", itt.Err().Error())
	// 	}
	// 	return "", "", nil
	// }

	// valarr, _ := client.Scan(ctx, 0, key, -1).Val()
	// if len(valarr) == 0 {
	// 	fmt.Println(key, "does not exist inside redis")
	// 	return "", "", nil
	// }

	//itt.Val()[0] because the result could be more than 1 row
	//val, err := client.Get(ctx, itt.Val()[0]).Result()
	val, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		fmt.Println(key, "does not exist inside redis (get)")
		return "", "", nil
	}
	if err != nil {
		fmt.Println("error when get keys on redis:", err.Error())
		return "", "", err
	}

	//return itt.Val()[0], val, nil
	return key, val, nil
}

func GetKeyValue(key string, db int) (string, error) {
	client := initializeRedisReplica(db)

	valarr, _ := client.Scan(ctx, 0, key, 0).Val()

	if len(valarr) == 0 {
		fmt.Println(key, "does not exist inside redis(len):", key)
		return "", nil
	}

	return valarr[0], nil
}

func RPush(key string, value []byte, db int) {
	client := initializeRedis(db)
	res := client.RPush(ctx, key, value)
	log.Println("successfully rpush:", res.Args()[1])
}

func FlushAll() bool {
	var err error

	type dts struct {
		Current int `json:"resync"`
	}

	var dt dts

	key := "last-resync"
	currentTimestamp := time.Now().Unix()

	_, ts, err := Get(key, common.REDIS_DB)
	if err != nil {
		fmt.Println(err.Error())
	}

	if ts != "" {
		json.Unmarshal([]byte(ts), &dt)

		if (currentTimestamp - int64(dt.Current)) <= 600 {
			return false
		}
	}

	client := initializeRedis(0)
	client.FlushAll(ctx)

	dt.Current = int(currentTimestamp)
	err2 := Set(key, dt, common.REDIS_DB)
	if err2 != nil {
		fmt.Println(err2.Error())
		return false
	}

	fmt.Println("FlushAll succeeded at", time.Now().String())
	return true
}

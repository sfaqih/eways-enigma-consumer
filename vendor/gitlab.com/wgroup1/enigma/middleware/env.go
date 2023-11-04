package middleware

import (
	"log"

	"github.com/spf13/viper"
)

// ViperEnvVariable is func to get .env file
func NewViperLoad()  {

	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()

	if err != nil {
		log.Println("Error while reading config file:", err)
	}

	
}

func GetViperEnvVariable(key string) string {
	
	value, ok := viper.Get(key).(string)

	if !ok {
		log.Printf("Key %s not found in env file", key)
		return ""
	}

	return value
}
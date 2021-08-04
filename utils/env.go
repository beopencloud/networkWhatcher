package utils

import (
	"encoding/base64"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	CnocdNamespaceLabelKey   = "beopenit.com/network-watching"
	CnocdNamespaceLabelValue = "true"
)



// ++
// +
// Ce fichier contient la definition des variables d'environment utilisés dans les autres packages.
// Chaque variable a une valeur par defaut.
// +
// ++


var (
	API_CONTENT_TYPE         = "application/json; charset=utf8 "
	USERNAME                 = "test"
	PASSWORD                 = "test"
	BASIC_AUTH_CREDENTIALS   = ""
	SERVICE_CREATE_EVENT_URL = "http://localhost:8082/service/post"
	SERVICE_UPDATE_EVENT_URL = "http://localhost:8082/service/put"
	SERVICE_DELETE_EVENT_URL = "http://localhost:8082/service/delete"
	INGRESS_CREATE_EVENT_URL = "http://localhost:8082/ingress/post"
	INGRESS_UPDATE_EVENT_URL = "http://localhost:8082/ingress/put"
	INGRESS_DELETE_EVENT_URL = "http://localhost:8082/ingress/delete"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		_ = godotenv.Load("./../../.env")

	}
	API_CONTENT_TYPE = getStringValue("API_CONTENT_TYPE", API_CONTENT_TYPE)
	USERNAME := getStringValue("USERNAME", USERNAME)
	PASSWORD := getStringValue("PASSWORD", PASSWORD)
	BASIC_AUTH_CREDENTIALS = base64.StdEncoding.EncodeToString([]byte(USERNAME + ":" + PASSWORD))

	SERVICE_CREATE_EVENT_URL = getStringValue("SERVICE_CREATE_EVENT_URL", SERVICE_CREATE_EVENT_URL)
	SERVICE_UPDATE_EVENT_URL = getStringValue("SERVICE_UPDATE_EVENT_URL", SERVICE_UPDATE_EVENT_URL)
	SERVICE_DELETE_EVENT_URL = getStringValue("SERVICE_DELETE_EVENT_URL", SERVICE_DELETE_EVENT_URL)

	INGRESS_CREATE_EVENT_URL = getStringValue("INGRESS_CREATE_EVENT_URL", INGRESS_CREATE_EVENT_URL)
	INGRESS_UPDATE_EVENT_URL = getStringValue("INGRESS_UPDATE_EVENT_URL", INGRESS_UPDATE_EVENT_URL)
	INGRESS_DELETE_EVENT_URL = getStringValue("INGRESS_DELETE_EVENT_URL", INGRESS_DELETE_EVENT_URL)

	log.Println("env loaded")
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func getBrokers(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return strings.Split(value, ",")
}

func getBoolValue(key string, defaultValue bool) bool {

	if len(os.Getenv(key)) == 0 {
		return defaultValue
	}
	val, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		log.Fatal("Error ", key, " must be a boolean. default value ", defaultValue, " is loaded")
		return defaultValue
	}
	return val
}

func getIntValue(key string, defaultValue int) int {
	if len(os.Getenv(key)) == 0 {
		return defaultValue
	}
	val, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		log.Fatal("Error", key, "must be a int. default value ", defaultValue, " is loaded")
		return defaultValue
	}
	return val
}

func getStringValue(key string, defaultValue string) string {
	if len(os.Getenv(key)) == 0 {
		return defaultValue
	}
	return os.Getenv(key)
}

func getUrl(key string, defaultValue string) string {
	result := os.Getenv(key)
	if len(result) == 0 {
		result = defaultValue
	}
	lastChar := result[len(result)-1:]
	if lastChar == "/" {
		result = result[:len(result)-1]
	}
	return result
}

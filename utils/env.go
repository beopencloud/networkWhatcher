package utils

import (
	"encoding/base64"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	CnocdNamespaceLabelKey   = "intrabpce.fr/network-watching"
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
	USERNAME                 = "modou"
	PASSWORD                 = "test"
	BASIC_AUTH_CREDENTIALS   = ""
	SERVICE_CREATE_EVENT_URL = "http://localhost:31015/service/post"
	SERVICE_UPDATE_EVENT_URL = "http://localhost:31015/service/put"
	SERVICE_DELETE_EVENT_URL = "http://localhost:31015/service/delete"
	IN_CLUSTER               = false
	KUBECONFIG               = filepath.Join(homeDir(), ".kube", "config")
)

func init() {
	fmt.Println("ENVOK")
	err := godotenv.Load()
	if err != nil {
		_ = godotenv.Load("./../../.env")

	}
	API_CONTENT_TYPE = getStringValue("API_CONTENT_TYPE", API_CONTENT_TYPE)
	USERNAME = getStringValue("USERNAME", USERNAME)
	PASSWORD = getStringValue("PASSWORD", PASSWORD)
	BASIC_AUTH_CREDENTIALS = base64.StdEncoding.EncodeToString([]byte(USERNAME + ":" + PASSWORD))

	SERVICE_CREATE_EVENT_URL = getStringValue("SERVICE_CREATE_EVENT_URL", SERVICE_CREATE_EVENT_URL)
	SERVICE_UPDATE_EVENT_URL = getStringValue("SERVICE_UPDATE_EVENT_URL", SERVICE_UPDATE_EVENT_URL)
	SERVICE_DELETE_EVENT_URL = getStringValue("SERVICE_DELETE_EVENT_URL", SERVICE_DELETE_EVENT_URL)

	KUBECONFIG = getStringValue("KUBECONFIG", KUBECONFIG)
	fmt.Println("OK INCLUSTER")
	IN_CLUSTER = getBoolValue("IN_CLUSTER", IN_CLUSTER)

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

	fmt.Println("KEY", key, "DEFAY", defaultValue, "OS", os.Getenv(key))
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

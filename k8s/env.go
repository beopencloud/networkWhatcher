package k8s

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
)




// ++
// +
// Ce fichier contient la definition des variables d'environment utilisés dans les autres packages.
// Chaque variable a une valeur par defaut.
// +
// ++


var (
	IN_CLUSTER            = false
    KUBECONFIG            = filepath.Join(homeDir(), ".kube", "config")
)



func init() {
	err := godotenv.Load()
	if err != nil {
		_ = godotenv.Load("./../../.env")

	}

    KUBECONFIG = getStringValue("KUBECONFIG", KUBECONFIG)
	IN_CLUSTER = getBoolValue("IN_CLUSTER", IN_CLUSTER)

	log.Println("env loaded")
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
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


func getStringValue(key string, defaultValue string) string {
	if len(os.Getenv(key)) == 0 {
		return defaultValue
	}
	return os.Getenv(key)
}



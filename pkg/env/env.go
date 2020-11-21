package env

import (
	"fmt"
	"os"

	"github.com/kraem/zhuyi-go/pkg/log"
)

func GetEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func GetEnvOrExit(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		err := fmt.Errorf("env var %s not set", key)
		log.LogError(err)
		os.Exit(1)
	}
	return val
}

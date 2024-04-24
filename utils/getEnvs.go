package utils

import (
	"os"
	"strconv"
)

func GetEnvInt(key string) int {
	v, _ := strconv.Atoi(os.Getenv(key))
	return v
}

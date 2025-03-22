package util

import (
	"os"
	"strings"
	"sync/atomic"
	"time"
)

func Uint64UpdateAndGet(v *int64, f func(int64) int64) int64 {
	var old, next int64
	for {
		old = atomic.LoadInt64(v)
		next = f(old)
		if atomic.CompareAndSwapInt64(v, old, next) {
			break
		}
	}
	return next
}

// IsBlank checks whether the given string is blank.
func IsBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func DateToSecond(date string) uint64 {
	c := date + " 00:00:00"
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation("2006-01-02 15:04:05", c, loc)
	return uint64(theTime.Unix())
}

const (
	ENV_KEY_HOST = "JPAAS_HOST"
	//ENV_KEY_PORT          = "JPAAS_HTTP_PORT"
	//ENV_KEY_PORT_ORIGINAL = "JPAAS_HOST_PORT_8080"
)

func GetHostAndE() (string, int) {
	DOCKER_HOST := os.Getenv(ENV_KEY_HOST)
	//DOCKER_PORT := os.Getenv(ENV_KEY_PORT)
	//if DOCKER_PORT == "" {
	//	DOCKER_PORT = os.Getenv(ENV_KEY_PORT_ORIGINAL)
	//}
	if DOCKER_HOST != "" {
		return DOCKER_HOST, 1
	} else {
		hostname, err := os.Hostname()
		if err != nil {
			return hostname, 2
		}
		return "127.0.0.1", 3
	}
}

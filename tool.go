package unigo

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
	ENV_PHYSICS   = 1
	ENV_CONTAINER = 2
)

func GetHostAndE() (string, int, error) {
	DOCKER_HOST := os.Getenv(ENV_KEY_HOST)
	if DOCKER_HOST != "" {
		return DOCKER_HOST, ENV_CONTAINER, nil
	} else {
		hostname, err := os.Hostname()
		if err != nil {
			return "", -1, err
		}
		return hostname, ENV_PHYSICS, nil
	}
}

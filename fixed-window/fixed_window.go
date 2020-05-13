package main

import (
	"fmt"
	"github.com/orcaman/concurrent-map"
	"guobingithub/rate-limiter/logger"
	"sync/atomic"
	"time"
)

const (
	Default_Threshold = 100

	keyPrefix = "fixed_window"
)

var (
	secondLimiter limiter
	limiterGroup  = make([]*limiter, 0)
)

type limiter struct {
	internal   int64
	threshHold int32
	userData   cmap.ConcurrentMap
}

type counter struct {
	start int64
	count int32
}

func getMillSecond(val int64) int64 {
	return val / int64(time.Millisecond)
}

func getKey(keyPrefix string, hasLogin bool) string {
	if hasLogin {
		return fmt.Sprintf("%s:%s", keyPrefix, "userId001")
	} else {
		return fmt.Sprintf("%s:%s", keyPrefix, "192.168.1.110")
	}
}

func (l *limiter) processLimit(key string) bool {
	var getVal *counter
	now := getMillSecond(int64(time.Now().UnixNano()))
	for {
		get, ok := l.userData.Get(key)
		if ok {
			if realNum, ok := get.(*counter); ok {
				if realNum.start+l.internal < now {
					getVal = &counter{
						start: getMillSecond(int64(time.Now().UnixNano())),
						count: 1,
					}

					l.userData.Set(key, getVal)
					break
				}
				atomic.AddInt32(&realNum.count, 1)
				getVal = realNum
				break
			}
		}

		getVal = &counter{
			start: getMillSecond(int64(time.Now().UnixNano())),
			count: 1,
		}

		ok = l.userData.SetIfAbsent(key, getVal)
		if ok {
			break
		}
	}

	return getVal.count > l.threshHold
}

func init() {
	secondLimiter = limiter{
		internal:   getMillSecond(int64(time.Second)),
		threshHold: Default_Threshold,
		userData:   cmap.New(),
	}

	limiterGroup = append(limiterGroup, &secondLimiter)
}

func main() {
	var key = getKey(keyPrefix, false)

	var count = 1
	for i := 0; i < 150; i++ {
		if secondLimiter.processLimit(key) {
			logger.Error(fmt.Sprintf("第%v次，拒绝访问!!!", count))
		} else {
			logger.Info(fmt.Sprintf("第%v次，可以访问^_^", count))
		}

		count++
		time.Sleep(5 * time.Millisecond)
	}
}

package main

import (
	"fmt"
	"guobin/rate-limiter/logger"
	"sync"
	"time"

	tokenBucket "github.com/juju/ratelimit"
)

var (
	mu           sync.Mutex
	tokenBuckets map[string]*tokenBucket.Bucket
	rate         = float64(0.15)
	capacity     = int64(15)

	keyPrefix = "token_bucket"
)

func init() {
	tokenBuckets = make(map[string]*tokenBucket.Bucket)
}

func rateLimit(key string, rate float64, capacity int64) bool {
	mu.Lock()
	defer mu.Unlock()

	if _, found := tokenBuckets[key]; !found {
		tokenBuckets[key] = tokenBucket.NewBucketWithRate(rate, capacity)
	}
	bucket := tokenBuckets[key]
	count := bucket.TakeAvailable(1)

	return count == 1
}

func getKey(keyPrefix string, hasLogin bool) string {
	if hasLogin {
		return fmt.Sprintf("%s:%s", keyPrefix, "userId001")
	} else {
		return fmt.Sprintf("%s:%s", keyPrefix, "192.168.1.110")
	}
}

func main() {
	var (
		ok  bool
		key string
	)

	key = getKey(keyPrefix, false)

	for i := 0; i < 50; i++ {
		ok = rateLimit(key, 1/rate, capacity)
		if ok {
			logger.Info(fmt.Sprintf("======可以访问，时间:(%v)", time.Now().Unix()))
		} else {
			logger.Error(fmt.Sprintf("======拒绝访问，时间:(%v)", time.Now().Unix()))
		}

		time.Sleep(100 * time.Millisecond)
	}

	//模拟3秒时间内没有访问
	fmt.Println("模拟3s空档期\n\n\n")
	time.Sleep(3 * time.Second)

	wg := new(sync.WaitGroup)
	wg.Add(15)
	for i := 0; i < 15; i++ {
		go func() {
			defer wg.Done()
			ok = rateLimit(key, 1/rate, capacity)
			if ok {
				logger.Info(fmt.Sprintf("======可以访问，时间:(%v)", time.Now().Unix()))
			} else {
				logger.Error(fmt.Sprintf("======拒绝访问，时间:(%v)", time.Now().Unix()))
			}
		}()
	}

	wg.Wait()
	fmt.Println("模拟瞬时突发流量访问\n\n\n")

	for {
		ok = rateLimit(key, 1/rate, capacity)
		if ok {
			logger.Info(fmt.Sprintf("======可以访问，时间:(%v)", time.Now().Unix()))
		} else {
			logger.Error(fmt.Sprintf("======拒绝访问，时间:(%v)", time.Now().Unix()))
		}

		time.Sleep(100 * time.Millisecond)
	}
}

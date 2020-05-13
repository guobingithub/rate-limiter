package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"guobingithub/rate-limiter/db"
	"guobingithub/rate-limiter/logger"
	"time"
)

const (
	RATELIMIT_STRING_KEY = "ratelimit:any"
)

func main() {
	logger.Info(fmt.Sprintf("rate-limiter starting..."))

	redisCli, err := db.NewRedisClient("127.0.0.1:6379", 8, 10, 300)
	if err != nil {
		logger.Error(err)
		return
	}
	defer redisCli.Release()

	logger.Info(fmt.Sprintf("redis conn ok, redisCli:(%v)", redisCli))

	conn := redisCli.Get()
	if err = conn.Err(); err != nil {
		logger.Error(err)
		return
	}
	defer conn.Close()

	var lua = `local times = redis.call('incr', KEYS[1])
	
	if times == 1 then
		redis.call('expire', KEYS[1], ARGV[1])
	end
	
	if times > tonumber(ARGV[2]) then
		return 0
	end
	
	return 1
	`

	script := redis.NewScript(1, lua)
	if script != nil {
		var count = 1
		for {
			sRes, err := script.Do(conn, RATELIMIT_STRING_KEY, 20, 2)
			if err != nil {
				logger.Error(fmt.Sprintf("script.Do error, err:(%v)", err))
				return
			}

			logger.Info(fmt.Sprintf("get script result ok, sRes:(%v)", sRes))
			sResInt, _ := sRes.(int64)
			if sResInt == 1 {
				logger.Info(fmt.Sprintf("第%v次，可以访问^_^", count))
			} else {
				logger.Info(fmt.Sprintf("第%v次，拒绝访问!!!", count))
			}

			count++
			time.Sleep(1 * time.Second)
		}
	} else {
		logger.Fatal("script is nil!!!")
	}
}

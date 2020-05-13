package db

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"guobingithub/rate-limiter/logger"
)

const (
	DialTimeout = 3
)

type ScriptItem struct {
	Script   string
	KeyCount int
}

type RedisClient struct {
	pool    *redis.Pool
	scripts []*redis.Script
}

func NewRedisClient(url string, dbIndex, maxConNum, idleTimeout int) (*RedisClient, error) {
	var (
		c        *RedisClient
		addr     string
		password string
	)
	c = new(RedisClient)
	c.pool = &redis.Pool{
		MaxIdle:     maxConNum,
		IdleTimeout: time.Duration(idleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			var (
				conn redis.Conn
				err  error
			)
			index := strings.LastIndex(url, "@")
			if index < 0 {
				addr = url
			} else {
				password = url[:index]
				addr = url[index+1:] //if url illegal, this will get panic
			}
			conn, err = redis.Dial("tcp", addr, redis.DialPassword(password),
				redis.DialConnectTimeout(DialTimeout*time.Second), redis.DialDatabase(dbIndex))
			return conn, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) (err error) {
			if _, err := c.Do("PING"); err != nil {
				logger.Error(err)
			}
			return err
		},
	}

	return c, nil
}

func (c *RedisClient) Get() redis.Conn {
	return c.pool.Get()
}

// Load load srcipt
func (c *RedisClient) Load(script ...ScriptItem) error {
	conn := c.Get()
	if err := conn.Err(); err != nil {
		logger.Error(err)
		return err
	}
	defer conn.Close()
	l := len(script)
	if l == 0 {
		return errors.New("args nil")
	}
	s := make([]*redis.Script, 0, l)

	for _, v := range script {
		_, err := conn.Do("SCRIPT", "LOAD", v.Script)
		if err != nil {
			logger.Error(err, v.Script)
			return err
		}

		i := redis.NewScript(v.KeyCount, v.Script)
		if i == nil {
			panic(fmt.Sprintf("script illegal, %s, %d", v.Script, v.KeyCount))
		}
		s = append(s, i)
	}
	c.scripts = s
	return nil
}

// GetScript returns a srcipt
func (c *RedisClient) GetScript(index int) *redis.Script {
	if index < len(c.scripts) && index >= 0 {
		return c.scripts[index]
	}
	return nil
}

func (c *RedisClient) Release() error {
	return c.pool.Close()
}

package redis

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

var (
	pool        *redis.Pool
	redisHost   = "127.0.0.1:6379"
	redisPasswd = "123456"
)

func init() {
	pool = newPedisPool()
}

func RedisPool() *redis.Pool {
	return pool
}

// 创建redis连接池
func newPedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,              // 最大连接数
		MaxActive:   30,              // 最多活跃数
		IdleTimeout: 1 * time.Minute, // 超时回收
		Dial: func() (redis.Conn, error) {
			// 1.打开连接
			conn, err := redis.Dial("tcp", redisHost)
			if err != nil {
				fmt.Printf("连接redis失败: %v", err)
				return nil, err
			}
			// 2.访问认证 (我暂时不设置认证密码)
			//if _, err = conn.Do("AUTH", redisPasswd); err != nil {
			//	conn.Close()
			//	return nil, err
			//}
			return conn, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error { // 每拿到一个连接，先检查健康状态
			// t是上次被放回连接池的时间
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err // err!=nil会使此连接关闭
		},
	}
}

package util

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
)

const (
	url      = "localhost:6379"
	password = "123456"
	db       = 0 // 选择的db号

	// 连接池参数
	maxIdle     = 10  // 初始连接数
	maxActive   = 10  // 最大连接数
	idleTimeOut = 300 // 最长空闲时间
)

var pool *redis.Pool

func Init() {
	pool = &redis.Pool{
		MaxIdle:     maxIdle,     //最初的连接数量
		MaxActive:   maxActive,   //连接池最大连接数量,（0表示自动定义），按需分配
		IdleTimeout: idleTimeOut, //连接关闭时间 300秒 （300秒不使用自动关闭）
		// 连接
		Dial: func() (redis.Conn, error) { //要连接的redis数据库
			// 建立tcp连接
			c, err := redis.Dial("tcp", url)
			if err != nil {
				return nil, err
			}
			// 验证密码
			if _, err := c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}

			// 选择库
			if _, err := c.Do("SELECT", db); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
	}
	conn := pool.Get() //从连接池，取一个链接
	defer conn.Close()

	if _, err := conn.Do("PING"); err != nil {
		log.Printf("PING err %s", err)
	}

	log.Println("redis pool init success")

}

func Close() {
	if err := pool.Close(); err != nil {
		fmt.Printf("close redis pool error ： %s", err)
	}
}

func Zadd(key string, score string, value int64) (interface{}, error) {
	conn := pool.Get()
	defer conn.Close()
	return conn.Do("zadd", key, score, value)
}

// zset 删除 key
func Zrem(key string, value int64) (interface{}, error) {
	conn := pool.Get()
	defer conn.Close()
	return conn.Do("zrem", key, value)
}

/*
*
zset 倒序取max-min范围内数据
*/
func ZrevrangeByScore(key string, min string, max string) ([]interface{}, error) {
	conn := pool.Get()
	defer conn.Close()
	return redis.Values(conn.Do("zrevrangebyscore", key, max, min, "withscores"))
}

// zset 倒序取范围内 max-min 的 数据 + 偏移
func ZrevrangeByScoreOffset(key string, min string, max string, offset int, limit int) ([]interface{}, error) {
	conn := pool.Get()
	defer conn.Close()
	return redis.Values(conn.Do("zrevrangebyscore", key, max, min, "withscores", "limit", offset, limit))
}

// zset 高水位 从大到小
func FindTop(key string) ([]interface{}, error) {
	conn := pool.Get()
	defer conn.Close()
	return redis.Values(conn.Do("zrevrangebyscore", key, "+inf", "-inf", "withscores"))
}

// zset 从大到小 只返回value
func FindTopVal(key string) ([]string, error) {
	conn := pool.Get()
	defer conn.Close()
	return redis.Strings(conn.Do("zrevrangebyscore", redis.Args{}.Add(key).Add("+inf").Add("-inf")...))
}

// zset 从大到小 + 偏移
func FindTopOffset(key string, offset int, limit int) ([]interface{}, error) {
	conn := pool.Get()
	defer conn.Close()
	return redis.Values(conn.Do("zrevrangebyscore", key, "+inf", "-inf", "withscores", "limit", offset, limit))
}

// zset 从小到大 + 偏移
func FindLow(key string) ([]interface{}, error) {
	conn := pool.Get()
	defer conn.Close()
	return redis.Values(conn.Do("zrevrangebyscore", key, "+inf", "-inf", "withscores"))
}

// zset  从小到大 + 偏移
func FindLowOffset(key string, offset int, limit int) ([]interface{}, error) {
	conn := pool.Get()
	defer conn.Close()
	return redis.Values(conn.Do("zrevrangebyscore", key, "+inf", "-inf", "withscores", "limit", offset, limit))
}

// zset 元素总数量
func FindZSetCount(key string) (int64, error) {
	conn := pool.Get()
	defer conn.Close()
	return redis.Int64(conn.Do("zcount", redis.Args{}.Add(key).Add("-inf").Add("+inf")...))
}

//// zset 元素总数量
//func FindZSetCountByRange(key string, ) (int64, error) {
//	conn := pool.Get()
//	defer conn.Close()
//	return redis.Int64(conn.Do("zcount", redis.Args{}.Add(key).Add("-inf").Add("+inf")))
//}

// withscore 返回 需要转换
func WithScoreConvert(resp []interface{}) map[string]string {
	var res = make(map[string]string)
	var key, score = "", ""
	for i, v := range resp {
		if i%2 == 0 {
			//json.Unmarshal(v.([]byte), &item.val)
			// todo 不知道有字符集乱码情况没有 目前没发现
			key = string(v.([]byte))
		} else {
			//json.Unmarshal(v.([]byte), &item.score)
			score = string(v.([]byte))
			res[key] = score
		}
	}
	return res
}

/*
*
获取hash的多个字段
*/
func HMGet(key string, fields ...string) ([]interface{}, error) {
	conn := pool.Get()
	defer conn.Close()
	return redis.Values(conn.Do("hmget", fields))
}

/*
*
获取hash的多个字段 field 为 int64
*/
func HMGetFiledI64(key string, fields ...int64) ([]interface{}, error) {
	conn := pool.Get()
	defer conn.Close()
	// int64 不能直接添加到 redis参数 需要转换为 []interface{}
	l := len(fields)
	var is = make([]interface{}, l, l)
	for i, field := range fields {
		is[i] = field
	}
	fmt.Println(redis.Args{}.Add(is...))
	return redis.Values(conn.Do("hmget", redis.Args{}.Add(key).Add(is...)...))
}

/*
*
根据 int64的fields 返回 map<int64, string> field-val
*/
func HMGetI64ReturnMapI64(key string, fields ...int64) (map[int64]string, error) {
	res, err := HMGetFiledI64(key, fields...)
	if err != nil {
		return nil, err
	}
	log.Printf("HMGetFiledI64 end, res:%s", res)
	return ConvertHashFieldI64(fields, res), nil
}

// hmget 返回 需要转换
func ConvertHashFieldI64(fields []int64, resp []interface{}) map[int64]string {
	var resMap = make(map[int64]string)
	idLen := len(fields)
	for i, item := range resp {
		// 避免两集合长度对不上的情况
		if i >= idLen {
			log.Fatal("HashConvertFieldI64 根据userIds查出来的userNames长度对不上")
			continue
		}
		var name = "unknown"
		if item != nil {
			name = string(item.([]byte))
		}
		resMap[fields[i]] = name
	}
	return resMap
}
package util

import (
	"fmt"
	"log"
	"testing"
)

var testKey = "testKey1"
var testVal = 1

func TestZadd(t *testing.T) {
	res, err := Zadd(testKey, GetFollowedTimeStr(), int64(testVal))
	if err != nil {
		fmt.Printf("zadd err ：%s", err)
	}

	fmt.Println(res, "zadd success")

	ans, err := ZrevrangeByScoreOffset("k1", "-inf", "+inf", 0, 10)
	if err != nil {
		fmt.Println("zrevrange err :", err)
		return
	}

	for _, v := range ans {
		fmt.Printf("%s\n", v.([]byte))
	}

}

func TestZrem(t *testing.T) {
	//zrem()
}

func TestWithScoreConvert(t *testing.T) {
	followList, err := FindTop("relation_follow_1")
	if err != nil {
		fmt.Println(err)
	}
	res := WithScoreConvert(followList)
	for k, v := range res {
		fmt.Println(k, v)
	}
}

func TestHMGetI64ReturnMapI64(t *testing.T) {
	key := GetUserNameKey()
	userIds := []int64{1, 2, 3, 4}
	resMap, err := HMGetI64ReturnMapI64(key, userIds...)
	if err != nil {
		log.Fatal("TestHMGetI64ReturnMapI64 exception：", err)
	}
	for k, v := range resMap {
		fmt.Println(k, v)
	}

}

func TestFindZSetCount(t *testing.T) {
	key := GetFollowKey(12)

	count, err := FindZSetCount(key)
	if err != nil {
		log.Println("TestFindZSetCount exception:", err)
	}
	fmt.Println("zsetCount:", count)
}

func TestMain(m *testing.M) {
	//fmt.Println("begin")
	Init()
	m.Run()
	//println(GetFollowedTimeStr())
	Close()
	//fmt.Println("end")
}
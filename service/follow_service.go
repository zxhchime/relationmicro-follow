package service

import (
	"fmt"
	"log"
	redisUtil "relationmicor/util"
	"strconv"
	"time"
)

// 关注者
type FollowUser struct {
	userId       int64
	followedTime time.Time // 关注时间
}

func Init() {
	redisUtil.Init()
}

//var redisUtil = util.

/*
*
关注
myUid: 我的userId
targetUid：关注目标userId
1、targetUid添加到关注集合
2、myUid添加到粉丝集合
3、存入两个用户的name 避免db消耗
#保证原子
*/
func Follow(myUid int64, targetUid int64) error {
	// 校验参数
	if err := CheckFollowParam(myUid, targetUid); err != nil {
		return fmt.Errorf("follow: myUid:%d, targetUid:%d, exception:%s", myUid, targetUid, err)
	}

	// 获取key
	key := redisUtil.GetFollowKey(myUid)
	// 关注 关注时间是now
	// todo 互关加入好友
	// todo 我关注别人的同时 也要我成为别人的粉丝
	// 关注时间 精确到秒级
	// zset 超过17位会精度丢失
	_, err := redisUtil.Zadd(key, redisUtil.GetFollowedTimeStr(), targetUid)
	if err != nil {
		return fmt.Errorf("follow: myUid:%d, targetUid:%d, exception:%s", myUid, targetUid, err)
	}
	return nil
}

// 取关
/**
1、targetUid从my关注集合删除
2、myUid从target粉丝集合删除
# 保证原子
 */
func UnFollow(myUid int64, targetUid int64) error {
	// 校验参数
	if err := CheckFollowParam(myUid, targetUid); err != nil {
		return fmt.Errorf("unFollow: myUid:%d, targetUid:%d, exception:%s", myUid, targetUid, err)
	}
	// 获取key
	key := redisUtil.GetFollowKey(myUid)
	// todo 取关 如果是好友的话 删除好友
	// todo 我取关别人的同时，也要从别人的粉丝中消失
	_, err := redisUtil.Zrem(key, targetUid)
	if err != nil {
		return fmt.Errorf("unFollow: myUid:%d, targetUid:%d, exception:%s", myUid, targetUid, err)
	}
	return nil
}

// 接口文档没给分页接口
// 查询关注集合
func FindFollowList(userId int64) ([]FollowUser, error) {
	var followList = make([]FollowUser, 0, 0)
	var followItem FollowUser

	key := redisUtil.GetFollowKey(userId)
	res, err := redisUtil.FindTopVal(key)
	if err != nil {
		return nil, fmt.Errorf("FindFollowList: userId:%d, exception:%s", userId, err)
	}

	//zset := redisUtil.WithScoreConvert(res)
	// []val -> followList
	for _, val := range res {
		// userId
		if followItem.userId, err = strconv.ParseInt(val, 10, 64); err != nil {
			log.Printf("FindFollowList: userId:%d parseInt exception:%s", userId, err)
			continue
		}
		followList = append(followList, followItem)
	}
	return followList, nil
}

func CheckFollowParam(myUid int64, targetUid int64) error {
	// 不能关注自己 不能取关自己
	if myUid == targetUid {
		return fmt.Errorf("param exception myUid and targetUid the same")
	}
	return nil
}
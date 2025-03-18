package utils

import (
	"strconv"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

func init() {
	// 传入的是节点的编号， 比如当前有三个节点部署了服务，编号1，2，3.
	//  不同节点的服务编号不能相同，否则会出现生成的ID重复的情况
	// 为了方便测试，这里直接写死了，实际开发中应该从配置文件中读取
	n, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	node = n
}

func GenerateSnowflakeID() snowflake.ID {
	return node.Generate()
}

// StringToSnowflakeID 将字符串转换为snowflake.ID
func StringToSnowflakeID(s string) (snowflake.ID, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return snowflake.ID(id), nil
}

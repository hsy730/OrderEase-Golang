package utils

import (
	"errors"
	"log"

	"github.com/sqids/sqids-go"
)

// 全局初始化 sqid 实例
var sqidInstance *sqids.Sqids

func init() {
	var err error
	// 如果需要配置盐值、最小长度等，可使用 NewWithData 方法初始化
	sqidInstance, err = sqids.New() // 或者 sqids.NewWithData(config)
	if err != nil {
		log.Fatalf("failed to initialize sqids: %v", err)
	}
}

// EncryptOrderID 将内部的雪花ID转换为加密后的字符串
func EncryptID(id uint64) (string, error) {
	encoded, err := sqidInstance.Encode([]uint64{id})
	if err != nil {
		return "", err
	}
	return encoded, nil
}

// DecryptOrderID 将加密后的字符串还原为内部的雪花ID
func DecryptID(encryptedID string) (int64, error) {
	ids := sqidInstance.Decode(encryptedID)

	if len(ids) == 0 {
		return 0, errors.New("invalid encrypted ID")
	}
	return int64(ids[0]), nil
}

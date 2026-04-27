package utils

import (
	"github.com/google/uuid"
)

// NewUUID 生成一个新的 UUID v4 并返回其字符串表示
func NewUUID() string {
	return uuid.New().String()
}
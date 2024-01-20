// Package main
// @Author: lzq
// @Email: lizuoqiang@huanjutang.com
// @Date: 2024-01-20 13:48:00
// @Description:
package main

import "strings"

func toCamelCase(input string) string {
	// 分割输入字符串为单词
	words := strings.Split(input, "_")

	// 遍历每个单词并将首字母转换为大写
	for i, word := range words {
		words[i] = strings.Title(word)
	}

	// 合并回驼峰格式的字符串
	return strings.Join(words, "")
}

func isCreateIgnoreField(field string) bool {
	ignoreFields := []string{"Id", "IsDeleted", "CreatedAt", "UpdatedAt"}
	return inSlice(field, ignoreFields)
}

// inSlice
//
//	@Description: 是否在切片内
//	@param val
//	@param slice
//	@return bool
func inSlice(val interface{}, arr interface{}) bool {
	switch arr := arr.(type) {
	case []string:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	case []int:
		for _, a := range arr {
			if a == val {
				return true
			}
		}
	}
	return false
}

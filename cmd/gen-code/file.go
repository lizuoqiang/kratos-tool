// Package main
// @Author: lzq
// @Email: lizuoqiang@huanjutang.com
// @Date: 2024-01-20 12:06:00
// @Description:
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// GetBasePath
//
//	@Description: 获取当前工作目录
//	@return string
func GetBasePath(fileName string) string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir + fileName
}

// GenFile
//
//	@Description: 生成文件
//	@param filePath
//	@param fileContent
//	@param fileMode
//	@return error
func GenFile(filePath, fileContent string, fileMode os.FileMode) error {
	fmt.Println("生成文件：", filePath)
	// Separate path and file name
	pathInfo := path.Dir(filePath)

	// Check and create directory (including parent directories), true means recursive creation
	if err := os.MkdirAll(pathInfo, fileMode); err != nil {
		return err
	}

	// Create file if not exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if _, err := os.Create(filePath); err != nil {
			return err
		}
	}

	err := ioutil.WriteFile(filePath, []byte(fileContent), fileMode)
	if err != nil {
		fmt.Println("生成错误：", err.Error())
	}

	return nil
}

// GetFileContent
//
//	@Description: 获取模版文件
//	@param fileName
//	@return string
func GetFileContent(fileName string) string {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	return string(content)
}

// GetOutputPath
//
//	@Description: 获取输出路径
//	@param fileName
//	@return string
func GetOutputPath(fileName string) string {
	return GetBasePath("/") + fileName
}

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil // 文件存在
	}
	if os.IsNotExist(err) {
		return false, nil // 文件不存在
	}
	return false, err // 其他错误（如权限问题等）
}

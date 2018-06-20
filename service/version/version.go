package version

import (
	"io/ioutil"
	"os"
	"runtime"
)

var (
	appName string
	// Version 应用版本，从当前目录的version.info中提取，应该在build的时候根据当前的代码分支进行生成
	// 比如：echo "${BRANCH_NAME}.${BUILD_NUMBER}\c" > version.info
	Version = "unkown"
)

func loadVersion() {
	f, err := os.Open("version.info")
	if err != nil {
		return
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}

	Version = string(b)
}

// SetName 设置名称
func SetName(n string) {
	appName = n
}

// SetVersion 手动设置版本
func SetVersion(v string) {
	Version = v
}

// GetVersionInfo 获取当前版本
func GetVersionInfo() *VersionResponse {
	if Version == "unkown" {
		loadVersion()
	}

	return &VersionResponse{
		Name:     appName,
		Golang:   runtime.Version(),
		Cpus:     int32(runtime.NumCPU()),
		Routines: int32(runtime.NumGoroutine()),
		Version:  Version,
	}
}

package httpstat

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/restful/code"
)

type StatItem struct {
	Method string
	Path   string
	Count  int
}

// DelayCounter 返回查询延迟统计的单个统计
type DelayCounter struct {
	Level string
	Count int
}

var requestCostLevels = []time.Duration{
	time.Second,          // < 1s
	time.Second * 5,      // 1s  ~ 5s
	time.Second * 10,     // 5s  ~ 10s
	time.Second * 20,     // 10s ~ 20s
	time.Minute * 60,     // 20s ~ 60s
	time.Minute * 3,      // 60s ~ 3m
	time.Minute * 5,      // 3m ~ 5m
	time.Hour * 24 * 365, // > 5m
}

func levelName(level time.Duration) string {
	return level.String()
}

// RequestStat 单请求统计
type RequestStat struct {
	sync.Mutex
	delayLevels map[time.Duration]int
	resultMap   map[string]int
}

// Reset 清空数据
func (stat *RequestStat) Reset() {
	stat.Lock()
	defer stat.Unlock()
	stat.delayLevels = nil
	stat.resultMap = nil
}

// AddStat 添加一次统计数据
func (stat *RequestStat) AddStat(mcode string, delay time.Duration) {
	fmt.Println("AddStat")
	if stat.delayLevels == nil {
		stat.delayLevels = make(map[time.Duration]int, len(requestCostLevels))
	}

	for _, level := range requestCostLevels {
		if delay < level {
			stat.delayLevels[level] = stat.delayLevels[level] + 1
			break
		}
	}

	if stat.resultMap == nil {
		stat.resultMap = make(map[string]int, 15)
	}

	if mcode == "" {
		stat.resultMap["OK"] = stat.resultMap["OK"] + 1
	} else {
		stat.resultMap[mcode] = stat.resultMap[mcode] + 1
	}
}

func (stat *RequestStat) delayItem() interface{} {
	levels := make([]*DelayCounter, 0, len(requestCostLevels))
	for _, level := range requestCostLevels {
		levels = append(levels, &DelayCounter{
			Level: level.String(),
			Count: stat.delayLevels[level],
		})
	}
	return levels
}

func (stat *RequestStat) resultItem() interface{} {
	levels := []interface{}{}
	for name, count := range stat.resultMap {
		levels = append(levels, fmt.Sprintf("%v,%d", name, count))
	}
	return levels
}

// NewRequestStat 构造一种请求的分析统计
func NewRequestStat() *RequestStat {
	return &RequestStat{}
}

// RequestStatMgr 分析统计入口
type RequestStatMgr struct {
	sync.Mutex
	stats map[string]*RequestStat
}

// NewRequestStatMgr 构造一个请求分析统计
func NewRequestStatMgr() *RequestStatMgr {
	return &RequestStatMgr{}
}

// AddStat 统计一次请求
func (statMgr *RequestStatMgr) AddStat(ctx *gin.Context, cerr code.Error, delay time.Duration) {
	statMgr.Lock()
	defer statMgr.Unlock()
	if statMgr.stats == nil {
		statMgr.stats = make(map[string]*RequestStat, 10)
	}

	statName := statMgr.statName(ctx.Request.Method, ctx.Request.URL.Path)
	stat, exist := statMgr.stats[statName]
	if !exist {
		stat = NewRequestStat()
		statMgr.stats[statName] = stat
	}

	mcode := ""
	if cerr != nil {
		mcode = cerr.Mcode()
	}

	stat.AddStat(mcode, delay)
}

func (statMgr *RequestStatMgr) statName(method, path string) string {
	return fmt.Sprintf("%s%s", method, path)
}

func (statMgr *RequestStatMgr) handleStatDelay(ctx *gin.Context) (interface{}, error) {
	requestName := ctx.Query("request")
	statItems := make([]interface{}, 0, 1)
	if requestName != "" {
		func() {
			statMgr.Lock()
			defer statMgr.Unlock()
			stat, exist := statMgr.stats[requestName]
			if exist {
				statItems = append(statItems, map[string]interface{}{
					"name": requestName,
					"stat": stat.delayItem(),
				})
			}
		}()
	} else {
		func() {
			statMgr.Lock()
			defer statMgr.Unlock()
			for name, stat := range statMgr.stats {
				statItems = append(statItems, map[string]interface{}{
					"name": name,
					"stat": stat.delayItem(),
				})
			}
		}()
	}

	return statItems, nil
}

func (statMgr *RequestStatMgr) handleStatResult(ctx *gin.Context) (interface{}, error) {
	requestName := ctx.Query("request")

	statItems := make([]interface{}, 0, 1)
	if requestName != "" {
		// 将+替换为/
		requestName = strings.Replace(requestName, "+", "/", -1)
		func() {
			statMgr.Lock()
			defer statMgr.Unlock()
			stat, exist := statMgr.stats[requestName]
			if exist {
				statItems = append(statItems, map[string]interface{}{
					"name": requestName,
					"stat": stat.resultItem(),
				})
			}
		}()
	} else {
		func() {
			statMgr.Lock()
			defer statMgr.Unlock()
			for name, stat := range statMgr.stats {
				statItems = append(statItems, map[string]interface{}{
					"name": name,
					"stat": stat.resultItem(),
				})
			}
		}()
	}
	return statItems, nil
}

// reset 重置
func (statMgr *RequestStatMgr) reset() {
	statMgr.Lock()
	defer statMgr.Unlock()

	statMgr.stats = nil
}

var statMgr = NewRequestStatMgr()

// Stat 统计函数
// ctx gin.Context对象
// status http结果
// cerr 错误
// delay 延迟
func Stat(ctx *gin.Context, status int, cerr code.Error, delay time.Duration) {
	_ = status
	// TODO(keto@lwork.com):status尚未纳入统计分支
	statMgr.AddStat(ctx, cerr, delay)
}

// Reset 重置统计
func Reset() {
	statMgr.reset()
}

// StatDelay 处理延迟统计请求
func StatDelay(ctx *gin.Context) (interface{}, error) {
	return statMgr.handleStatDelay(ctx)
}

// StatResult 统计处理结果
func StatResult(ctx *gin.Context) (interface{}, error) {
	return statMgr.handleStatResult(ctx)
}

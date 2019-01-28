package httpsrv

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/restful/code"
)

// SnowSlide 抑制过载
type snowSlide struct {
	curTime  int64
	mutex    sync.Mutex
	curCount int32
	LimitCnt int32
	Service  string
}

// Check 检查是否过载
func (snowslide *snowSlide) Check(ctx *gin.Context) code.Error {
	if snowslide.LimitCnt <= 0 {
		return nil
	}

	timeNow := time.Now().Unix()
	snowslide.mutex.Lock()
	defer snowslide.mutex.Unlock()

	if timeNow > snowslide.curTime {
		snowslide.curTime = timeNow
		snowslide.curCount = 1
		return nil
	}

	if snowslide.curCount >= snowslide.LimitCnt {
		return code.NewMcodef("SNOWSLIDE_DENIED", "Check Snow Protect failed,service = %v", snowslide.Service)
	}

	snowslide.curCount++
	return nil
}

// SnowSlide 是防止雪崩的防御对象
type SnowSlide interface {
	Check(*gin.Context) code.Error
}

// NewDefaultSnowSlide 创建一个内置默认
func NewDefaultSnowSlide(cnt int32, service string) SnowSlide {
	return &snowSlide{
		LimitCnt: cnt,
		Service:  service,
	}
}

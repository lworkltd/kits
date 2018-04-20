package grpcsrv

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	context "golang.org/x/net/context"

	"github.com/golang/protobuf/proto"
	"github.com/lworkltd/kits/service/grpcsrv/grpccomm"
	"github.com/lworkltd/kits/service/restful/code"
)

func unexpectError(format string, args ...interface{}) {
	panic(fmt.Errorf("[grpcsrv] "+format, args...))
}

// ErrorRsp 生成GRPC错误返回
func newErrorRsp(mcode string, format string, args ...interface{}) *grpccomm.CommResponse {
	return &grpccomm.CommResponse{
		Result:  false,
		Mcode:   mcode,
		Message: fmt.Sprintf("grpcsrv:"+format, args...),
	}
}

func newRspFromError(err error) *grpccomm.CommResponse {
	cerr, is := err.(code.Error)
	if is {
		// 此错误一般由间接调用参数或一些通用错误残生
		if cerr.Mcode() != "" {
			return newErrorRsp(cerr.Mcode(), cerr.Error())
		}
		// 此类错误一般由服务内部参数，返回了一个数字类型的错误码
		return newErrorRsp(
			fmt.Sprintf("%s%d", mcodePrefix, cerr.Code()),
			cerr.Error())
	}

	// 无法识别的错误
	return newErrorRsp("GRPC_UNKOWN", "grpc unkown error,%v", err.Error())
}

var (
	// CerrCheckSnowProtect 雪崩预警
	CerrCheckSnowProtect = code.New(10011, "Check Snow Protect")
	gCurTime             int64
	checkSnowMutex       sync.Mutex
	gCurCount            int32
)

func checkSnowSlide(showCount int32) error {
	timeNow := time.Now().Unix()
	checkSnowMutex.Lock()
	defer checkSnowMutex.Unlock()

	if timeNow > gCurTime {
		gCurTime = timeNow
		gCurCount = 1
		return nil
	}
	if gCurCount >= showCount {
		return CerrCheckSnowProtect
	}
	gCurCount++

	return nil
}

func createRegInfo(reqBody interface{}, f interface{}) *RegisterInfo {
	if reqBody == nil {
		unexpectError("register body nil")
	}

	var (
		bodyTp  = reflect.TypeOf(reqBody)
		reqName string
	)
	if reflect.TypeOf(reqBody).Kind() == reflect.String {
		reqName = reqBody.(string)
	} else {
		if bodyTp.Kind() == reflect.Ptr {
			bodyTp = bodyTp.Elem()
		}
		reqName = bodyTp.Name()
	}

	regInfo := &RegisterInfo{
		reqName:       reqName,
		ctxInIndex:    -1,
		reqInIndex:    -1,
		headerInIndex: -1,
		errIndex:      -1,
		rspOutIndex:   -1,
		commReqIndex:  -1,
	}

	// 计算不同入参的位置
	// 如果包含Context，则类型必然为Context
	// 如果包含请求参数，则类型名称必然后缀为Req或Request
	// 如果包含头部，则类型名称后缀Header
	ft := reflect.TypeOf(f)
	regInfo.inNum = ft.NumIn()
	for i := 0; i < ft.NumIn(); i++ {
		inType := ft.In(i)
		name := inType.String()
		// 请求透传
		if name == "grpccomm.CommRequest" {
			unexpectError("grpccomm.CommRequest must be a pointer")
		}

		if name == "*grpccomm.CommRequest" {
			regInfo.commReqIndex = i
			continue
		}

		switch {
		case name == "*context.Context":
			unexpectError("context.Context must not be a pointer")
		case name == "context.Context":
			regInfo.ctxInIndex = i
		case strings.HasSuffix(name, "Req") || strings.HasSuffix(name, "Request"):
			if inType.Kind() != reflect.Ptr {
				unexpectError("request struct must be type of pointer,got %s", name)
			}

			if !inType.Implements(interfaceTypeProtoMessage) {
				unexpectError("request struct must implement proto.Message,got %s", name)
			}

			regInfo.newBody = func() reflect.Value {
				return reflect.New(inType.Elem())
			}
			regInfo.reqInIndex = i
		case strings.HasSuffix(name, "Header"):
			if inType.Kind() != reflect.Ptr {
				unexpectError("header struct must be type of pointer")
			}
			if !inType.Implements(interfaceTypeProtoMessage) {
				unexpectError("header struct must implement proto.Message")
			}
			regInfo.newHeader = func() reflect.Value {
				return reflect.New(inType.Elem())
			}
			regInfo.headerInIndex = i
		default:
			unexpectError("unkown input register type %v", inType.String())
		}
	}

	numOut := ft.NumOut()
	// 输出两个参数，第一个为响应结构，第二个为错误
	// 输出一个参数，为错误或者为透传
	// 允许不输出参数
	if numOut == 2 {
		if !ft.Out(0).Implements(interfaceTypeProtoMessage) {
			unexpectError("1st struct must implement proto.Message")
		}
		regInfo.rspOutIndex = 0
		if !ft.Out(1).Implements(interfaceTypeError) {
			unexpectError("2nd output parameter must implement error")
		}
		regInfo.errIndex = 1
	} else if numOut == 1 {
		for {
			if ft.Out(0).Implements(interfaceTypeError) {
				regInfo.errIndex = 0
				break
			}
			if ft.Out(0).Kind() != reflect.Ptr {
				unexpectError(
					"0st output parameter must be error or *grpcomm.CommResponse,got %v",
					ft.Out(0).String(),
				)
			}

			if ft.Out(0).String() == "*grpccomm.CommResponse" {
				regInfo.commRspOutIndex = 0
				break
			}
			unexpectError(
				"0st output parameter must implement error or be type of *grpcomm.CommResponse,got %v",
				ft.Out(0).String(),
			)
		}
	}

	// 调用函数
	fv := reflect.ValueOf(f)
	regInfo.call = fv.Call
	regInfo.callFuncName = fv.String()

	return regInfo
}

func call0(ctx context.Context, headerValue, bodyValue reflect.Value, commReq *grpccomm.CommRequest, regInfo *RegisterInfo) *grpccomm.CommResponse {
	// 调用函数
	var (
		reqVals = make([]reflect.Value, regInfo.inNum)
	)

	// 参数包含Ctx
	if regInfo.ctxInIndex >= 0 {
		reqVals[regInfo.ctxInIndex] = reflect.ValueOf(ctx)
	}

	// 参数包含Header
	if regInfo.headerInIndex >= 0 {
		reqVals[regInfo.headerInIndex] = headerValue
	}

	// 参数包含请求Body
	if regInfo.reqInIndex >= 0 {
		reqVals[regInfo.reqInIndex] = bodyValue
	}

	// 参数包含grpccomm.CommRequest
	if regInfo.commReqIndex >= 0 {
		reqVals[regInfo.commReqIndex] = reflect.ValueOf(commReq)
	}

	var (
		rspVal     reflect.Value
		rspErr     reflect.Value
		rspCommRsp reflect.Value
	)

	//fmt.Println(regInfo.inNum, reqVals, "header", regInfo.headerInIndex, "req", regInfo.reqInIndex, "ctx", regInfo.ctxInIndex, "commReq", regInfo.commReqIndex)

	// 执行调用
	retVals := regInfo.call(reqVals)

	// 不需要返回数据
	if len(retVals) == 0 {
		return &grpccomm.CommResponse{
			Result: true,
		}
	}

	// 错误或commReq
	if len(retVals) == 1 {
		if regInfo.commRspOutIndex == 0 {
			rspCommRsp = retVals[0]
		}
		if regInfo.errIndex == 0 {
			rspErr = retVals[0]
		}
	}

	// 数据和错误
	if len(retVals) == 2 {
		rspVal = retVals[0]
		rspErr = retVals[1]
	}

	// 返回错误
	if rspErr.IsValid() && !rspErr.IsNil() {
		return newRspFromError(rspErr.Interface().(error))
	}

	if rspCommRsp.IsValid() && !rspCommRsp.IsNil() {
		return rspCommRsp.Interface().(*grpccomm.CommResponse)
	}

	// 如果需要，并且返回不为NIL,则返回数据
	rspBody := []byte{}
	if regInfo.rspOutIndex >= 0 && !rspVal.IsNil() {
		rspMsg, yes := rspVal.Interface().(proto.Message)
		if !yes {
			unexpectError("return values[0] not inplement proto.Message")
		}

		b, err := proto.Marshal(rspMsg)
		if err != nil {
			unexpectError("return values[0] can't marshal into bytes,%v", err)
		}
		rspBody = b
	}

	return &grpccomm.CommResponse{
		Result: true,
		Body:   rspBody,
	}
}

var (
	interfaceTypeProtoMessage = reflect.TypeOf(new(proto.Message)).Elem()
	interfaceTypeError        = reflect.TypeOf(new(error)).Elem()
)

func debugRegisterInfo(regInfo *RegisterInfo) {
	//fmt.Println(regInfo.reqName, "func", regInfo.callFuncName, "(", regInfo.inNum, ")(", regInfo.outNum, ")")
}

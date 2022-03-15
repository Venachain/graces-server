package response

import (
	"fmt"
	"net/http"

	"graces/exterr"
	"graces/model"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	DefaultFailMsg     = "fail"
	DefaultFailCode    = http.StatusBadRequest
	DefaultSuccessMsg  = "success"
	DefaultSuccessCode = http.StatusOK
)

// Success 响应成功
func Success(ctx *gin.Context, result model.Result) {
	if result.Code == 0 {
		result.Code = DefaultSuccessCode
	}
	if result.Msg == "" {
		result.Msg = DefaultSuccessMsg
	}
	result.Request = fmt.Sprintf("[%s] %s", ctx.Request.Method, ctx.Request.URL.String())
	ctx.JSON(result.Code, result)
}

// Fail 响应失败
func Fail(ctx *gin.Context, result model.Result) {
	if result.Code == 0 {
		result.Code = DefaultFailCode
	}
	if result.Msg == "" {
		result.Msg = DefaultFailMsg
	}
	result.Request = fmt.Sprintf("[%s] %s", ctx.Request.Method, ctx.Request.URL.String())
	ctx.JSON(result.Code, result)
}

// Response 自定义响应
func Response(ctx *gin.Context, result model.Result) {
	result.Request = fmt.Sprintf("[%s] %s", ctx.Request.Method, ctx.Request.URL.String())
	ctx.JSON(result.Code, result)
}

// ErrorHandler 统一错误处理器
func ErrorHandler(ctx *gin.Context, e error) {
	switch e.(type) {
	case exterr.ExtError:
		handlerErr(ctx, e.(exterr.ExtError))
	case *exterr.ExtError:
		handlerErr(ctx, *e.(*exterr.ExtError))
	default:
		handlerUnknownErr(ctx, e)
	}
}

// 处理已知错误
func handlerErr(ctx *gin.Context, err exterr.ExtError) {
	switch {
	// 10000 ~ 10099 系统通用错误
	case 10000 <= err.Code && err.Code < 10100:
		handlerSysErr(ctx, err)
	// 10100 ~ 10199 用户相关错误
	case 10100 <= err.Code && err.Code < 10200:
		handlerUserErr(ctx, err)
	// 10200 ~ 10299 业务相关的错误
	case 10200 <= err.Code && err.Code < 10300:
		handlerBusinessErr(ctx, err)
	default:
		handlerUnknownErr(ctx, err)
	}
}

// 处理未知错误
func handlerUnknownErr(ctx *gin.Context, err error) {
	logrus.Errorf("unknown error: %+v", err)
	code := http.StatusInternalServerError
	msg := fmt.Sprintf("unknown error: %v", exterr.ErrUnknown.Code)
	if gin.Mode() == gin.DebugMode {
		msg = err.Error()
	}
	result := model.Result{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
	Fail(ctx, result)
	return
}

// 处理系统相关错误
func handlerSysErr(ctx *gin.Context, err exterr.ExtError) {
	logrus.Errorf("system error: %+v", err)
	code := http.StatusInternalServerError
	msg := fmt.Sprintf("Error: %v", err.Code)
	if gin.Mode() == gin.DebugMode {
		msg = err.Error()
	}
	result := model.Result{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
	Fail(ctx, result)
	return
}

// 处理用户相关错误
func handlerUserErr(ctx *gin.Context, err exterr.ExtError) {
	logrus.Debugln(err)
	code := http.StatusBadRequest
	switch err.Code {
	case exterr.ErrCodeUnauthorized:
		code = http.StatusUnauthorized
	case exterr.ErrCodeBadRole | exterr.ErrCodeUserHasNoPermission:
		code = http.StatusForbidden
	default:
		code = http.StatusBadRequest
	}

	msg := fmt.Sprintf("Error: %v", err.Code)
	if gin.Mode() == gin.DebugMode {
		msg = err.Error()
	}
	result := model.Result{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
	Fail(ctx, result)
	return
}

// 处理业务相关错误
func handlerBusinessErr(ctx *gin.Context, err exterr.ExtError) {
	logrus.Debugln(err)
	code := http.StatusBadRequest
	msg := fmt.Sprintf("Error: %v", err.Code)
	if gin.Mode() == gin.DebugMode {
		msg = err.Error()
	}
	result := model.Result{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
	Fail(ctx, result)
	return
}

package middleware

import (
	"graces/exterr"
	"graces/web/util/response"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// PanicHandler gin 统一 panic 处理器
func PanicHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logrus.Errorf("unknown panic：%+v", err)
				ctx.Abort()
				e := exterr.NewError(exterr.ErrCodeUnknown, err)
				response.ErrorHandler(ctx, e)
				return
			}
		}()
		ctx.Next()
	}
}

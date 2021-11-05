package middleware

import (
	"PlatONE-Graces/exterr"
	"PlatONE-Graces/web/util/response"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
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

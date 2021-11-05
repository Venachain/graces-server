package middleware

import (
	"PlatONE-Graces/exterr"
	"PlatONE-Graces/model"
	"PlatONE-Graces/secret"
	"PlatONE-Graces/web/util/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// LoginAuth 登录用户认证
func LoginAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result := model.Result{}
		// 首先在请求头获取 token
		auth := ctx.Request.Header.Get("Authorization")
		if len(auth) == 0 {
			ctx.Abort()
			result.Code = http.StatusUnauthorized
			result.Msg = exterr.ErrUnauthorized.Error()
			response.Fail(ctx, result)
			return
		}
		// 解析 token
		j := secret.NewJWT()
		claims, err := j.ParseToken(auth)
		if err != nil {
			ctx.Abort()
			result.Code = http.StatusUnauthorized
			result.Msg = err.Error()
			response.Fail(ctx, result)
			return
		}
		// 认证通过则刷新 JWT 过期时间
		_, _ = j.RefreshToken(auth)

		// 继续交由下一个路由处理，并将解析出的信息传递下去
		ctx.Set("claims", claims)
		ctx.Next()
	}
}

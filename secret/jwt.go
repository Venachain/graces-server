package secret

import (
	"PlatONE-Graces/config"
	"PlatONE-Graces/exterr"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// CustomClaims 自定义 Claims，继承 jwt.StandardClaims 并添加一些自己需要的信息
type CustomClaims struct {
	UserId   string `json:"userId"`
	Username string `json:"username"`
	jwt.StandardClaims
}

func NewDefaultCustomClaims() *CustomClaims {
	expiresTime := time.Now().Add(config.Config.JWTConf.Expires * time.Second).Unix()
	return &CustomClaims{
		UserId:   "",
		Username: "",
		StandardClaims: jwt.StandardClaims{
			Audience:  "",                           // 受众
			ExpiresAt: expiresTime,                  // 失效时间
			Id:        "",                           // 编号
			IssuedAt:  time.Now().Unix(),            // 签发时间
			Issuer:    config.Config.JWTConf.Issuer, // 签发人
			NotBefore: time.Now().Unix(),            // 生效时间
			Subject:   "",                           // 主题
		},
	}
}

// JWT 结构
type JWT struct {
	// SigningKey 密钥信息
	SigningKey []byte
}

// NewJWT 创建一个 JWT 实例
func NewJWT() *JWT {
	return &JWT{
		[]byte(config.Config.JWTConf.SecKey),
	}
}

// CreateToken 创建 JWT
func (j *JWT) CreateToken(claims CustomClaims) (token string, err error) {
	// 通过 HS256 算法生成 tokenClaims ,这就是我们的 HEADER 部分和 PAYLOAD。
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = tokenClaims.SignedString(j.SigningKey)
	// 返回添加了前缀的 token
	return config.Config.JWTConf.Prefix + token, err
}

// ParseToken 解析 JWT
func (j *JWT) ParseToken(tokenStr string) (*CustomClaims, error) {
	auth := strings.Fields(tokenStr)
	if len(auth) < 1 {
		return nil, exterr.ErrTokenInvalid
	}
	// 解析 token 时去除前缀的影响
	token, err := jwt.ParseWithClaims(auth[1], &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, exterr.ErrTokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return nil, exterr.ErrTokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, exterr.ErrTokenNotValidYet
			} else {
				return nil, exterr.ErrTokenInvalid
			}
		}
	}
	if err == nil && token != nil {
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			return claims, nil
		}
	}
	return nil, exterr.ErrTokenInvalid
}

// RefreshToken 刷新 JWT
func (j *JWT) RefreshToken(tokenStr string) (string, error) {
	auth := strings.Fields(tokenStr)
	if len(auth) < 1 {
		return "", exterr.ErrTokenInvalid
	}
	// 解析 token 时去除前缀的影响
	token, err := jwt.ParseWithClaims(auth[1], &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		claims.StandardClaims.ExpiresAt = time.Now().Add(config.Config.JWTConf.Expires * time.Second).Unix()
		return j.CreateToken(*claims)
	}
	return "", exterr.ErrTokenInvalid
}

// InvalidToken 让 JWT 失效
// 这里实测，并没有产生应有的效果
func (j *JWT) InvalidToken(tokenStr string) error {
	auth := strings.Fields(tokenStr)
	if len(auth) < 1 {
		return exterr.ErrTokenInvalid
	}
	// 解析 token 时去除前缀的影响
	token, err := jwt.ParseWithClaims(auth[1], &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		return err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		// 使这个 JWT 立即失效
		claims.StandardClaims.ExpiresAt = time.Now().Unix()
		return nil
	}
	return exterr.ErrTokenInvalid
}

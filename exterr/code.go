package exterr

//common error
const (
	// 10000 ~ 10099 系统通用错误
	ErrCodeDB      = 10000
	ErrCodeConvert = 10001

	ErrCodeUnknown = 10099

	// 10100 ~ 10199 用户相关错误
	ErrCodeUnauthorized        = 10100
	ErrCodeUserRegistered      = 10101
	ErrCodeUserHasNoPermission = 10102
	ErrCodeBadRole             = 10103
	ErrCodeUserNotExist        = 10104
	ErrCodeUserExisted         = 10105
	ErrCodePasswordWrong       = 10106
	ErrCodeTokenInvalid        = 10107
	ErrCodeTokenMalformed      = 10108
	ErrCodeTokenExpired        = 10109
	ErrCodeTokenNotValidYet    = 10110

	// 10200 ~ 10299 业务相关的错误
	ErrCodeParameterInvalid = 10200
	ErrCodeInsert           = 10201
	ErrCodeUpdate           = 10202
	ErrCodeDelete           = 10203
	ErrCodeFind             = 10204
	ErrCodeObjectIDInvalid  = 10205
	ErrGetStats 			= 10206
	ErrorSysconfigGetting   = 10207
	ErrorSysconfigSetting   = 10208
	ErrorWithContractFirewall = 10209
	ErrorContractWithCNS = 10210
	ErrorContractParams = 10211


	// 10300 ~ 10399 websocket 相关的错误
	ErrCodeWebsocketGroupInvalid     = 10300
	ErrCodeWebsocketGroupNotExist    = 10301
	ErrCodeWebsocketClientNotExist   = 10302
	ErrCodeWebsocketClientNotInGroup = 10303
	ErrCodeWebsocketDial             = 10304
	ErrCodeWebsocketClientSend       = 10305
	ErrCodeWebsocketSubscription     = 10306
	ErrCodeWebsocketSubMsgProcess    = 10307

	// 10400 ~ 10499 链数据同步相关错误
	ErrCodeChainDataSync = 10400
)

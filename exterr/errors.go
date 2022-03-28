package exterr

//common error
var (
	// 10000 ~ 10099 系统通用错误
	ErrDB      = NewError(ErrCodeDB, "system error 01")
	ErrConvert = NewError(ErrCodeConvert, "object convert error")
	ErrUnknown = NewError(ErrCodeUnknown, "system error 02")

	// 10100 ~ 10199 用户相关错误
	ErrUnauthorized        = NewError(ErrCodeUnauthorized, "unauthorized")
	ErrUserHasRegistered   = NewError(ErrCodeUserRegistered, "user registered")
	ErrUserHasNoPermission = NewError(ErrCodeUserHasNoPermission, "no permission")
	ErrBadRole             = NewError(ErrCodeBadRole, "bad role")
	ErrUserNotExist        = NewError(ErrCodeUserNotExist, "user not exist")
	ErrUserExisted         = NewError(ErrCodeUserExisted, "user is existed")
	ErrPasswordWrong       = NewError(ErrCodePasswordWrong, "username or password is incorrect")
	ErrTokenInvalid        = NewError(ErrCodeTokenInvalid, "token invalid error")
	ErrTokenMalformed      = NewError(ErrCodeTokenMalformed, "token malformed error")
	ErrTokenExpired        = NewError(ErrCodeTokenExpired, "token expired error")
	ErrTokenNotValidYet    = NewError(ErrCodeTokenNotValidYet, "token not valid yet error")

	// 10200 ~ 10299 业务相关的错误
	ErrParameterInvalid   = NewError(ErrCodeParameterInvalid, "parameter invalid error")
	ErrInsert             = NewError(ErrCodeInsert, "insert error")
	ErrUpdate             = NewError(ErrCodeUpdate, "update error")
	ErrDelete             = NewError(ErrCodeDelete, "delete error")
	ErrFind               = NewError(ErrCodeFind, "find error")
	ErrObjectIDInvalid    = NewError(ErrCodeObjectIDInvalid, "ObjectID invalid")
	ErrorGetStats         = NewError(ErrGetStats, "get chain stats error")
	ErrorSetSysconfig     = NewError(ErrorSysconfigSetting, "set sysconfig is error")
	ErrorGetSysconfig     = NewError(ErrorSysconfigGetting, "get sysconfig is error")
	ErrorContractFirewall = NewError(ErrorWithContractFirewall, "open or close contract firewall is error")
	ErrrorContractByCns   = NewError(ErrorContractWithCNS, "error find contract with cns name")
	ErrorContractParam    = NewError(ErrorContractParams, "contract param is error")
	ErrChainDataSync      = NewError(ErrCodeChainDataSync, "chain data sync error")
	ErrContractDeploy     = NewError(ErrCodeContractDeploy, "contract deploy error")

	// 10300 ~ 10399 websocket 相关的错误
	ErrWebsocketGroupInvalid     = NewError(ErrCodeWebsocketGroupInvalid, "websocket group invalid")
	ErrWebsocketGroupNotExist    = NewError(ErrCodeWebsocketGroupNotExist, "websocket group not exist")
	ErrWebsocketClientNotExist   = NewError(ErrCodeWebsocketClientNotExist, "websocket client not exist")
	ErrWebsocketClientNotInGroup = NewError(ErrCodeWebsocketClientNotInGroup, "websocket client not in target group")
	ErrWebsocketDial             = NewError(ErrCodeWebsocketDial, "websocket dial error")
	ErrWebsocketClientSend       = NewError(ErrCodeWebsocketClientSend, "ClientSend function call must be a websocket dial connection")
	ErrWebsocketSubscription     = NewError(ErrCodeWebsocketSubscription, "websocket subscription error")
	ErrWebsocketSubMsgProcess    = NewError(ErrCodeWebsocketSubMsgProcess, "websocket subscription message process error")
)

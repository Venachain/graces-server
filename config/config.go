package config

import (
	"fmt"
	"time"

	"graces/util/validate"

	"github.com/BurntSushi/toml"
)

const (
	configFile = "config.toml"
)

var (
	Config *config
)

func init() {
	loadConfigFromFile(configFile)
}

type config struct {
	HttpConf    *httpConf              `toml:"http"`
	LogConf     *logConf               `toml:"log"`
	DBConf      *dbConf                `toml:"db"`
	JWTConf     *jwtConf               `toml:"jwt"`
	WSConf      *wsConf                `toml:"ws"`
	ChainConfig map[string]interface{} `toml:"chain_config"`
	Syncer      *syncer                `toml:"syncer"`
}

type httpConf struct {
	IP   string `toml:"ip" validate:"required"`
	Port string `toml:"port" validate:"required"`
	Mode string `toml:"mode" validate:"required,oneof=debug release test"`
	Cors string `toml:"cors" validate:"required"`
}

func (hc *httpConf) Addr() string {
	return fmt.Sprintf("%s:%s", hc.IP, hc.Port)
}

type dbConf struct {
	IP       string        `toml:"ip" validate:"required"`
	Port     string        `toml:"port" validate:"required"`
	UserName string        `toml:"username" validate:"required"`
	Password string        `toml:"password"`
	DBName   string        `toml:"dbname"`
	Timeout  time.Duration `toml:"timeout"`
}

func (dbc *dbConf) Uri() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s",
		dbc.UserName,
		dbc.Password,
		dbc.IP,
		dbc.Port,
		dbc.DBName,
	)
}

type wsConf struct {
	BuffSize       int64           `toml:"buff_size" validate:"required"`
	Timeout        time.Duration   `toml:"timeout" validate:"required,min=0"`
	RetryInterval  time.Duration   `toml:"retry_interval" validate:"required,min=1"`
	MaxRetryCnt    int64           `toml:"max_retry_cnt" validate:"required,min=0"`
	WsMsgTypesConf *wsMsgTypesConf `toml:"msg_types" validate:"required"`
}

type wsMsgTypesConf struct {
	Sub *subMsgTypeConf `toml:"sub" validate:"required"`
	Pub *pubMsgTypeConf `toml:"pub" validate:"required"`
}

type subMsgTypeConf struct {
	SendType    string `toml:"send_type" validate:"required"`
	ReceiveType string `toml:"receive_type" validate:"required"`
}

type pubMsgTypeConf struct {
	BlockType    string `toml:"block_type" validate:"required"`
	TXType       string `toml:"tx_type" validate:"required"`
	StatsType    string `toml:"stats_type" validate:"required"`
	NodeInfoType string `toml:"node_info_type" validate:"required"`
}

type jwtConf struct {
	// SecKey 服务端秘钥
	SecKey string `toml:"seckey" validate:"required"`
	// Expires JWT 过期时间，单位：秒
	Expires time.Duration `toml:"expires" validate:"required"`
	// Issuer JWT 签发人
	Issuer string `toml:"issuer"`
	// Prefix JWT 生成 token 所添加的前缀
	Prefix string `toml:"prefix"`
}

type Level uint32
type logConf struct {
	//Level      string `toml:"level"`
	//Output     string `toml:"output"`
	//FilePath   string `toml:"filepath"`
	//FileDirAbs string
	//FileName   string
	//Size 	   string `toml:"size"`

	Level Level  `toml:"level"`
	Size  string `toml:"size"`
	Path  string `toml:"path"`
}

type syncer struct {
	Delay        time.Duration `toml:"delay" validate:"required,min=1"`
	IncrInterval time.Duration `toml:"incr_interval" validate:"required,min=1"`
}

// 加载配置信息
func loadConfigFromFile(file string) {
	if _, err := toml.DecodeFile(file, &Config); err != nil {
		panic(err)
	}
	if err := validate.Validate(*Config); err != nil {
		panic(err)
	}
}

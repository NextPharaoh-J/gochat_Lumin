package config

import (
	"github.com/spf13/viper"
	"os"
	"runtime"
	"strings"
	"sync"
)

var (
	once     sync.Once
	realPath string
	Conf     *Config
)

const (
	SuccessReplyCode      = 0
	FailReplyCode         = 1
	SuccessReplyMsg       = "success"
	QueueName             = "gochat_queue"
	RedisBaseValidTime    = 86400
	RedisPrefix           = "gochat_"
	RedisRoomPrefix       = "gochat_room_"
	RedisRoomOnlinePrefix = "gochat_room_online_count_"
	MsgVersion            = 1
	OpSingleSend          = 2 // single user
	OpRoomSend            = 3 // send to room
	OpRoomCountSend       = 4 // get online user count
	OpRoomInfoSend        = 5 // send info to room
	OpBuildTcpConn        = 6 // build tcp conn
)

type Config struct {
	Common  Common
	Connect ConnectConfig
	Logic   LogicConfig
	Task    TaskConfig
	Api     ApiConfig
	Site    SiteConfig
}

type Common struct {
	CommonEtcd  CommonEtcd  `mapstructure:"common-etcd"`
	CommonRedis CommonRedis `mapstructure:"common-redis"`
}
type CommonEtcd struct {
	Host              string `mapstructure:"host"`
	BasePath          string `mapstructure:"basePath"`
	ServerPathLogic   string `mapstructure:"serverPathLogic"`
	ServerPathConnect string `mapstructure:"serverPathConnect"`
	UserName          string `mapstructure:"userName"`
	Password          string `mapstructure:"password"`
	ConnectionTimeout int    `mapstructure:"connectionTimeout"`
}
type CommonRedis struct {
	RedisAddress  string `mapstructure:"redisAddress"`
	RedisPassword string `mapstructure:"redisPassword"`
	DB            int    `mapstructure:"db"`
}

type ConnectConfig struct {
	ConnectBase                ConnectBase                `mapstructure:"connect.toml-base"`
	ConnectBucket              ConnectBucket              `mapstructure:"connect.toml-bucket"`
	ConnectWebsocket           ConnectWebsocket           `mapstructure:"connect.toml-websocket"`
	ConnectTcp                 ConnectTcp                 `mapstructure:"connect.toml-tcp"`
	ConnectRpcAddressWebSockts ConnectRpcAddressWebsockts `mapstructure:"connect.toml-rpcAddress-websockts"`
	ConnectRpcAddressTcp       ConnectRpcAddressTcp       `mapstructure:"connect.toml-rpcAddress-tcp"`
}
type ConnectBase struct {
	CertPath string `mapstructure:"certPath"`
	KeyPath  string `mapstructure:"keyPath"`
}
type ConnectBucket struct {
	CpuNum        int    `mapstructure:"cpuNum"`
	Channel       int    `mapstructure:"channel"`
	Room          int    `mapstructure:"room"`
	SrvProto      int    `mapstructure:"svrProto"`
	RoutineAmount uint64 `mapstructure:"routineAmount"`
	RoutineSize   int    `mapstructure:"routineSize"`
}
type ConnectWebsocket struct {
	ServerId string `mapstructure:"serverId"`
	Bind     string `mapstructure:"bind"`
}
type ConnectTcp struct {
	ServerId      string `mapstructure:"serverId"`
	Bind          string `mapstructure:"bind"`
	SendBuf       int    `mapstructure:"sendbuf"`
	ReceiveBuf    int    `mapstructure:"receivebuf"`
	KeepAlive     bool   `mapstructure:"keepalive"`
	Reader        int    `mapstructure:"reader"`
	ReadBuf       int    `mapstructure:"readBuf"`
	ReadBufSize   int    `mapstructure:"readBufSize"`
	Writer        int    `mapstructure:"writer"`
	WriterBuf     int    `mapstructure:"writerBuf"`
	WriterBufSize int    `mapstructure:"writeBufSize"`
}
type ConnectRpcAddressWebsockts struct {
	Address string `mapstructure:"address"`
}
type ConnectRpcAddressTcp struct {
	Address string `mapstructure:"address"`
}

type LogicConfig struct {
	LogicBase LogicBase `mapstructure:"logic-base"`
}
type LogicBase struct {
	ServerId   string `mapstructure:"serverId"`
	CpuNum     int    `mapstructure:"cpuNum"`
	RpcAddress string `mapstructure:"rpcAddress"`
	CertPath   string `mapstructure:"certPath"`
	KeyPath    string `mapstructure:"keyPath"`
}

type TaskConfig struct {
	TaskBase TaskBase `mapstructure:"task-base"`
}
type TaskBase struct {
	CpuNum        int    `mapstructure:"cpuNum"`
	RedisAddress  string `mapstructure:"redisAddress"`
	RedisPassword string `mapstructure:"redisPassword"`
	RpcAddress    string `mapstructure:"rpcAddress"`
	PushChan      int    `mapstructure:"pushChan"`
	PushChanSize  int    `mapstructure:"pushChanSize"`
}

type ApiConfig struct {
	ApiBase ApiBase `mapstructure:"api-base"`
}
type ApiBase struct {
	ListenPort int `mapstructure:"listenPort"`
}

type SiteConfig struct {
	SiteBase SiteBase `mapstructure:"site-base"`
}
type SiteBase struct {
	ListenPort int `mapstructure:"listenPort"`
}

func GetGinRunMode() string {
	env := os.Getenv("RUN_MODE")
	// gin mode [debug,test,release]
	if env == "prod" {
		return "release"
	}
	return "dev"
}
func getCurrentDIr() string {
	_, filename, _, _ := runtime.Caller(1)
	aPath := strings.Split(filename, "/")
	dir := strings.Join(aPath[0:len(aPath)-1], "/")
	return dir
}

func Init() {
	var err error
	once.Do(func() {
		env := GetGinRunMode()
		realPath = getCurrentDIr() //  thet same : filepath.Abs("./)
		configFilePath := realPath + "/" + env + "/"
		Conf = new(Config)
		viper.SetConfigType("toml")
		viper.AddConfigPath(configFilePath)
		viper.SetConfigName("/connect")
		if err = viper.ReadInConfig(); err != nil {
			panic(err)
		} else {
			err = viper.Unmarshal(&Conf.Connect)
			if err != nil {
				panic(err)
			}
		}
		viper.SetConfigName("/common")
		if err = viper.MergeInConfig(); err != nil {
			panic(err)
		} else {
			err = viper.Unmarshal(&Conf.Common)
			if err != nil {
				panic(err)
			}
		}
		viper.SetConfigName("/task")
		if err = viper.MergeInConfig(); err != nil {
			panic(err)
		} else {
			err = viper.Unmarshal(&Conf.Task)
			if err != nil {
				panic(err)
			}
		}
		viper.SetConfigName("/logic")
		if err = viper.MergeInConfig(); err != nil {
			panic(err)
		} else {
			err = viper.Unmarshal(&Conf.Logic)
			if err != nil {
				panic(err)
			}
		}
		viper.SetConfigName("/api")
		if err = viper.MergeInConfig(); err != nil {
			panic(err)
		} else {
			err = viper.Unmarshal(&Conf.Api)
			if err != nil {
				panic(err)
			}
		}
		viper.SetConfigName("/site")
		if err = viper.MergeInConfig(); err != nil {
			panic(err)
		} else {
			err = viper.Unmarshal(&Conf.Site)
			if err != nil {
				panic(err)
			}
		}
	})
}
func init() {
	Init()
}

//func main() {
//	fmt.Println(Conf)
//}

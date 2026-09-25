package module

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	utl "stargo/sr-utl"
	"strconv"
	"strings"
	"time"
)

const EMPTYSTR = ""

type FlexString string

func (s *FlexString) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = FlexString(str)
		return nil
	}
	var num json.Number
	if err := json.Unmarshal(data, &num); err == nil {
		*s = FlexString(num.String())
		return nil
	}
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		*s = FlexString(strconv.FormatBool(b))
		return nil
	}
	return fmt.Errorf("unsupported config value type: %s", string(data))
}

type ConfStruct struct {
	ClusterInfo struct {
		User       string `json:"user"`
		Version    string `json:"version"`
		CreateDate string `json:"create_date"`
		MetaPath   string `json:"meta_path"`
		PrivateKey string `json:"private_key"`
	} `json:"clusterinfo"`

	Global struct {
		User    string `json:"user"`
		SshPort uint32 `json:"ssh_port"`
	} `json:"global"`

	ServerConfig struct {
		Fe map[string]FlexString `json:"fe"`
		Be map[string]FlexString `json:"be"`
	} `json:"server_configs"`

	FeServers []struct {
		Host             string                `json:"host"`
		SshPort          uint32                `json:"ssh_port"`
		HttpPort         uint32                `json:"http_port"`
		RpcPort          uint32                `json:"rpc_port"`
		QueryPort        uint32                `json:"query_port"`
		EditLogPort      uint32                `json:"edit_log_port"`
		DeployDir        string                `json:"deploy_dir"`
		MetaDir          string                `json:"meta_dir"`
		LogDir           string                `json:"log_dir"`
		PriorityNetworks string                `json:"priority_networks"`
		Config           map[string]FlexString `json:"config"`
	} `json:"fe_servers"`

	BeServers []struct {
		Host                 string                `json:"host"`
		SshPort              uint32                `json:"ssh_port"`
		BePort               uint32                `json:"be_port"`
		WebServerPort        uint32                `json:"webserver_port"`
		HeartbeatServicePort uint32                `json:"heartbeat_service_port"`
		BrpcPort             uint32                `json:"brpc_port"`
		DeployDir            string                `json:"deploy_dir"`
		StorageDir           string                `json:"storage_dir"`
		LogDir               string                `json:"log_dir"`
		PriorityNetworks     string                `json:"priority_networks"`
		Config               map[string]FlexString `json:"config"`
	} `json:"be_servers"`

	PrometheusServer struct {
		Host      string `json:"host"`
		SshPort   uint32 `json:"ssh_port"`
		HttpPort  uint32 `json:"http_port"`
		DeployDir string `json:"deploy_dir"`
		DataDir   string `json:"data_dir"`
		LogDir    string `json:"log_dir"`
	} `json:"prometheus_servers"`

	GrafanaServer struct {
		Host      string `json:"host"`
		SshPort   uint32 `json:"ssh_port"`
		HttpPort  uint32 `json:"http_port"`
		DeployDir string `json:"deploy_dir"`
	} `json:"grafana_servers"`

	AlertManagerServer struct {
		Host        string `json:"host"`
		SshPort     uint32 `json:"ssh_port"`
		WebPort     uint32 `json:"web_port"`
		ClusterPort uint32 `json:"cluster_port"`
		DeployDir   string `json:"deploy_dir"`
		DataDir     string `json:"data_dir"`
		LogDir      string `json:"log_dir"`
	} `json:"alertmanager_servers"`
}

type SettingsStruct struct {

	// 默认配置文件目录
	DefaultSettingDir string

	// 配置文件路径
	SettingsFilePath string

	// SSH私钥文件路径，默认是 ~/.ssh/id_ed25519
	SshPrivateKeyFilePath string `json:"ssh_private_key_file_path"`

	// StarRocks安装包的地址
	Repo string `json:"repo"`

	// StarRocks安装包下载到本地的目录
	DownloadPath string `json:"download_path"`

	// 当前管理的集群名称
	ClusterName string `json:"cluster_name"`

	// 当前集群使用的StarRocks的版本
	SrVersion string `json:"version"`
}

var GlobalSettings SettingsStruct

var GClusterName string
var GConfigInfo *ConfStruct
var GAppendConfigInfo *ConfStruct
var GSshPrivateKeyFilePath string
var GSRCtlRoot string
var GSRVersion string
var GWriteBackMetaPath string
var GJdbcUser string
var GJdbcPasswd string
var GJdbcDb string
var GFeEntryHost string
var GFeEntryQueryPort uint32
var GFeEntryEditLogPort uint32
var GDownloadPath string

// DefaultLogDir 返回默认配置文件目录 ~/.local/share/stargo
func defaultSettingDir() string {
	basedir, err := os.UserHomeDir()
	if err != nil {
		basedir = os.TempDir()
	} else {
		basedir = filepath.Join(basedir, ".local/share")
	}
	return filepath.Join(basedir, "stargo")
}

// InitSettings 初始化配置信息，并从配置文件中读取配置信息
func InitSettings() (err error) {
	// 获取当前用户信息
	osUser, _ := user.Current()
	// 初始化配置文件路径，读取配置文件内容
	GlobalSettings.DefaultSettingDir = defaultSettingDir()
	GlobalSettings.SettingsFilePath = filepath.Join(GlobalSettings.DefaultSettingDir, "settings.json")
	content, err := os.ReadFile(GlobalSettings.SettingsFilePath)
	if err != nil {
		return err
	}
	// 解析配置文件内容
	err = json.Unmarshal(content, &GlobalSettings)
	if err != nil {
		return err
	}
	// 如果没有指定SSH私钥文件路径，默认使用 ~/.ssh/id_ed25519
	if GlobalSettings.SshPrivateKeyFilePath == "" {
		GlobalSettings.SshPrivateKeyFilePath = fmt.Sprintf("%s/.ssh/id_ed25519", osUser.HomeDir)
	}
	// 如果没有指定StarRocks安装包的地址，默认使用 ~/.stargo/download
	if strings.Contains(GlobalSettings.Repo, "file://") {
		GDownloadPath = strings.Replace(GlobalSettings.Repo, "file://", "", -1)
	} else {
		GDownloadPath = GlobalSettings.DefaultSettingDir + "/download"
	}
	return nil
}

func SetSrVersion(newVersion string) {
	GSRVersion = newVersion
}

func GetConf(fileName string) (sr_cluster_config *ConfStruct, err error) {
	jsonFile, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	sr_cluster_config = &ConfStruct{}
	err = json.Unmarshal(jsonFile, sr_cluster_config)
	if err != nil {
		return nil, err
	}
	return sr_cluster_config, nil
}

func InitConf(clusterName string, fileName string) (err error) {
	// get home dir & ssh auth key
	osUser, _ := user.Current()
	GSshPrivateKeyFilePath = fmt.Sprintf("%s/.ssh/id_ed25519", osUser.HomeDir)

	// get sr-ctl root dir
	GSRCtlRoot = os.Getenv("SRCTLROOT")
	if GSRCtlRoot == "" {
		GSRCtlRoot = fmt.Sprintf("%s/.stargo", osUser.HomeDir)
	}

	// get the write back meta path
	GClusterName = clusterName
	GWriteBackMetaPath = fmt.Sprintf("%s/cluster/%s", GSRCtlRoot, GClusterName)

	// get the FE jdbc connection parameters
	GJdbcUser = "root"
	GJdbcPasswd = ""
	GJdbcDb = ""

	// 如果没有指定配置文件路径，默认使用写入路径下的meta.json文件
	if fileName == "" {
		GConfigInfo, err = GetConf(GWriteBackMetaPath + "/meta.json")
		if err != nil {
			return err
		}
	} else {
		GConfigInfo, err = GetConf(fileName)
		if err != nil {
			return err
		}
	}

	return nil
}

func AppendConf(clusterName string) (err error) {
	var metaFile string

	osUser, _ := user.Current()
	GSshPrivateKeyFilePath = fmt.Sprintf("%s/.ssh/id_ed25519", osUser.HomeDir)
	GSRCtlRoot = os.Getenv("SRCTLROOT")
	if GSRCtlRoot == "" {
		GSRCtlRoot = fmt.Sprintf("%s/.stargo", osUser.HomeDir)
	}
	metaFile = fmt.Sprintf("%s/cluster/%s/meta.json", GSRCtlRoot, clusterName)

	GAppendConfigInfo, err = GetConf(metaFile)
	if err != nil {
		return err
	}

	return nil
}

func WriteBackMeta(cc *ConfStruct, metaFilePath string) (err error) {
	var infoMess string
	var metaFileName string
	var metaF *os.File

	// check the metaFile exist, if the file doesn't exist, create a new one.
	metaFileName = metaFilePath + "/meta.json"
	_ = os.MkdirAll(metaFilePath, 0751)
	metaF, err = os.Create(metaFileName)
	if err != nil {
		return err
	}
	defer metaF.Close()

	clusterNameArr := strings.Split(metaFilePath, "/")
	clusterName := clusterNameArr[len(clusterNameArr)-1]
	infoMess = fmt.Sprintf(`You can shoot the trouble as bellowing step:
	    1. check the meta file status [fileName = %s]
		2. check the cluster name you input [clusterName = %s]
		3. check the os env $SRCTLROOT, if you don't set this env variable, please check the ~/.stargo folder`,
		metaFileName,
		clusterName)

	// write back cluster info
	cc.ClusterInfo.User = GConfigInfo.Global.User
	cc.ClusterInfo.CreateDate = time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05")
	cc.ClusterInfo.Version = GSRVersion
	cc.ClusterInfo.MetaPath = GWriteBackMetaPath
	cc.ClusterInfo.PrivateKey = GSshPrivateKeyFilePath

	jsonStr, err := json.MarshalIndent(cc, "", "  ")
	if err != nil {
		infoMess = fmt.Sprintf("Error in marshalling json structure.")
		utl.Logger.Error(infoMess)
	}

	_, err = metaF.WriteString(string(jsonStr))
	if err != nil {
		return err
	}

	return nil
}

func SetFeEntry(feEntryId int) {
	GFeEntryHost = GConfigInfo.FeServers[feEntryId].Host
	GFeEntryQueryPort = GConfigInfo.FeServers[feEntryId].QueryPort
	GFeEntryEditLogPort = GConfigInfo.FeServers[feEntryId].EditLogPort
}

func TestParseYamlConfig(fileName string) {
	yamlConf, err := GetConf(fileName)
	if err != nil {
		panic(err)
	}

	// Print configuration
	fmt.Println("[TEST] >>>>>>>>", yamlConf)
	fmt.Println("[TEST] ######################### GLOBAL #########################")
	fmt.Println("[TEST] Global -> User: %s\n", yamlConf.Global.User)
	fmt.Println("[TEST] Global -> ssh_port: %s\n", yamlConf.Global.SshPort)
	fmt.Println("[TEST] ######################### SERVER CONFIG #########################")
	fmt.Println("[TEST] ServerConfig -> FE -> sys_log_level: ", yamlConf.ServerConfig.Fe["sys_log_level"])
	fmt.Println("[TEST] ServerConfig -> FE -> fe_sys_log_1: ", yamlConf.ServerConfig.Fe["fe_sys_log_1"])
	fmt.Println("[TEST] ServerConfig -> BE -> sys_log_level: ", yamlConf.ServerConfig.Be["sys_log_level"])
	fmt.Println("[TEST] ServerConfig -> BE -> be_sys_log_2: ", yamlConf.ServerConfig.Be["be_sys_log_2"])
	fmt.Println("[TEST] ######################### FE SERVER #########################")
	for i := 0; i < len(yamlConf.FeServers); i++ {
		fmt.Printf("[TEST] FeServer -> [%d] -> host:                             %s\n", i, yamlConf.FeServers[i].Host)
		fmt.Printf("[TEST] FeServer -> [%d] -> ssh_port:                         %d\n", i, yamlConf.FeServers[i].SshPort)
		fmt.Printf("[TEST] FeServer -> [%d] -> http_port:                        %d\n", i, yamlConf.FeServers[i].HttpPort)
		fmt.Printf("[TEST] FeServer -> [%d] -> rpc_port:                         %d\n", i, yamlConf.FeServers[i].RpcPort)
		fmt.Printf("[TEST] FeServer -> [%d] -> query_port:                       %d\n", i, yamlConf.FeServers[i].QueryPort)
		fmt.Printf("[TEST] FeServer -> [%d] -> edit_log_port:                    %d\n", i, yamlConf.FeServers[i].EditLogPort)
		fmt.Printf("[TEST] FeServer -> [%d] -> deploy_dir:                       %s\n", i, yamlConf.FeServers[i].DeployDir)
		fmt.Printf("[TEST] FeServer -> [%d] -> meta_dir:                         %s\n", i, yamlConf.FeServers[i].MetaDir)
		fmt.Printf("[TEST] FeServer -> [%d] -> log_dir:                          %s\n", i, yamlConf.FeServers[i].LogDir)
		fmt.Printf("[TEST] FeServer -> [%d] -> priority_networks:                %s\n", i, yamlConf.FeServers[i].PriorityNetworks)
		fmt.Printf("[TEST] FeServer -> [%d] -> config -> sys_log_level:          %s\n", i, yamlConf.FeServers[i].Config["sys_log_level"])
		fmt.Printf("[TEST] FeServer -> [%d] -> config -> sys_log_delete_age:     %s\n", i, yamlConf.FeServers[i].Config["sys_log_delete_age"])
	}

	fmt.Println("[TEST] ######################### BE SERVER #########################")
	for i := 0; i < len(yamlConf.BeServers); i++ {
		fmt.Printf("[TEST] BeServer -> [%d] -> host:                             %s\n", i, yamlConf.BeServers[i].Host)
		fmt.Printf("[TEST] BeServer -> [%d] -> ssh_port:                         %d\n", i, yamlConf.BeServers[i].SshPort)
		fmt.Printf("[TEST] BeServer -> [%d] -> be_port:                          %d\n", i, yamlConf.BeServers[i].BePort)
		fmt.Printf("[TEST] BeServer -> [%d] -> webserver_port:                   %d\n", i, yamlConf.BeServers[i].WebServerPort)
		fmt.Printf("[TEST] BeServer -> [%d] -> heartbeat_service_port:           %d\n", i, yamlConf.BeServers[i].HeartbeatServicePort)
		fmt.Printf("[TEST] BeServer -> [%d] -> deploy_dir:                       %s\n", i, yamlConf.BeServers[i].DeployDir)
		fmt.Printf("[TEST] BeServer -> [%d] -> storage_dir:                      %s\n", i, yamlConf.BeServers[i].StorageDir)
		fmt.Printf("[TEST] BeServer -> [%d] -> PriorityNetworks                  %s\n", i, yamlConf.BeServers[i].PriorityNetworks)
		fmt.Printf("[TEST] BeServer -> [%d] -> log_dir:                          %s\n", i, yamlConf.BeServers[i].LogDir)
		fmt.Printf("[TEST] BeServer -> [%d] -> config -> sys_log_level:          %s\n", i, yamlConf.BeServers[i].Config["create_tablet_worker_count"])
		fmt.Printf("[TEST] BeServer -> [%d] -> config -> sys_log_delete_age:     %s\n", i, yamlConf.BeServers[i].Config["sys_log_delete_age"])
	}

	fmt.Println("[TEST] ######################### PROMETHEUS SERVER #########################")
	fmt.Println("[TEST] PrometheusServer -> host: ", yamlConf.PrometheusServer.Host)
	fmt.Println("[TEST] PrometheusServer -> ssh_port: ", yamlConf.PrometheusServer.SshPort)
	fmt.Println("[TEST] PrometheusServer -> http_port: ", yamlConf.PrometheusServer.HttpPort)
	fmt.Println("[TEST] PrometheusServer -> deploy_dir: ", yamlConf.PrometheusServer.DeployDir)
	fmt.Println("[TEST] PrometheusServer -> data_dir: ", yamlConf.PrometheusServer.DataDir)
	fmt.Println("[TEST] PrometheusServer -> log_dir: ", yamlConf.PrometheusServer.LogDir)

	fmt.Println("[TEST] ######################### GRAFANA SERVER #########################")
	fmt.Println("[TEST] GrafanaServer -> host: ", yamlConf.GrafanaServer.Host)
	fmt.Println("[TEST] GrafanaServer -> ssh_port: ", yamlConf.GrafanaServer.SshPort)
	fmt.Println("[TEST] GrafanaServer -> http_port: ", yamlConf.GrafanaServer.HttpPort)
	fmt.Println("[TEST] GrafanaServer -> deploy_dir: ", yamlConf.GrafanaServer.DeployDir)

	fmt.Println("[TEST] ######################### ALERTMANAGER SERVER #########################")
	fmt.Println("[TEST] AlertManagerServer -> host: ", yamlConf.AlertManagerServer.Host)
	fmt.Println("[TEST] AlertManagerServer -> ssh_port: ", yamlConf.AlertManagerServer.SshPort)
	fmt.Println("[TEST] AlertManagerServer -> web_port: ", yamlConf.AlertManagerServer.WebPort)
	fmt.Println("[TEST] AlertManagerServer -> cluster_port: ", yamlConf.AlertManagerServer.ClusterPort)
	fmt.Println("[TEST] AlertManagerServer -> deploy_dir: ", yamlConf.AlertManagerServer.DeployDir)
	fmt.Println("[TEST] AlertManagerServer -> data_dir: ", yamlConf.AlertManagerServer.DataDir)
	fmt.Println("[TEST] AlertManagerServer -> log_dir: ", yamlConf.AlertManagerServer.LogDir)
}

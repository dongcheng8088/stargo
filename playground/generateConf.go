package playground

import (
	_ "embed"
	"fmt"
	"os"
	"os/user"
	"stargo/module"
	"time"
)

func InitPlaygroundConf() {

	var tmp module.ConfStruct

	osUser, _ := user.Current()
	module.GSshPrivateKeyFilePath = fmt.Sprintf("%s/.ssh/id_rsa", osUser.HomeDir)
	module.GSRCtlRoot = os.Getenv("SRCTLROOT")
	if module.GSRCtlRoot == "" {
		module.GSRCtlRoot = fmt.Sprintf("%s/.stargo", osUser.HomeDir)
	}
	tmpDeployDir := fmt.Sprintf("%s/playground", module.GSRCtlRoot)

	// ClusterInfo
	tmp.ClusterInfo.User = osUser.Username
	tmp.ClusterInfo.Version = "2.2.0"
	tmp.ClusterInfo.CreateDate = time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05")
	tmp.ClusterInfo.MetaPath = fmt.Sprintf("%s/cluster/sr-playground", module.GSRCtlRoot)
	tmp.ClusterInfo.PrivateKey = module.GSshPrivateKeyFilePath

	tmp.Global.User = osUser.Username
	tmp.Global.SshPort = 22

	tmp.FeServers = append(tmp.FeServers,
		struct {
			Host             string                       `json:"host"`
			SshPort          uint32                       `json:"ssh_port"`
			HttpPort         uint32                       `json:"http_port"`
			RpcPort          uint32                       `json:"rpc_port"`
			QueryPort        uint32                       `json:"query_port"`
			EditLogPort      uint32                       `json:"edit_log_port"`
			DeployDir        string                       `json:"deploy_dir"`
			MetaDir          string                       `json:"meta_dir"`
			LogDir           string                       `json:"log_dir"`
			PriorityNetworks string                       `json:"priority_networks"`
			Config           map[string]module.FlexString `json:"config"`
		}{
			Host:             "127.0.0.1",
			SshPort:          22,
			HttpPort:         8030,
			RpcPort:          9020,
			QueryPort:        9030,
			EditLogPort:      9010,
			DeployDir:        tmpDeployDir + "/fe",
			MetaDir:          tmpDeployDir + "/fe/meta",
			LogDir:           tmpDeployDir + "/fe/log",
			PriorityNetworks: "127.0.0.1/32",
			Config:           nil,
		})

	tmp.BeServers = append(tmp.BeServers,
		struct {
			Host                 string                       `json:"host"`
			SshPort              uint32                       `json:"ssh_port"`
			BePort               uint32                       `json:"be_port"`
			WebServerPort        uint32                       `json:"webserver_port"`
			HeartbeatServicePort uint32                       `json:"heartbeat_service_port"`
			BrpcPort             uint32                       `json:"brpc_port"`
			DeployDir            string                       `json:"deploy_dir"`
			StorageDir           string                       `json:"storage_dir"`
			LogDir               string                       `json:"log_dir"`
			PriorityNetworks     string                       `json:"priority_networks"`
			Config               map[string]module.FlexString `json:"config"`
		}{
			Host:                 "127.0.0.1",
			SshPort:              22,
			BePort:               9060,
			WebServerPort:        8040,
			HeartbeatServicePort: 9050,
			BrpcPort:             8060,
			DeployDir:            tmpDeployDir + "/be",
			StorageDir:           tmpDeployDir + "/be/storage",
			LogDir:               tmpDeployDir + "/be/log",
			PriorityNetworks:     "127.0.0.1/32",
			Config:               nil,
		})

	module.GConfigInfo = &tmp
	module.GSRVersion = "v" + module.GConfigInfo.ClusterInfo.Version
}

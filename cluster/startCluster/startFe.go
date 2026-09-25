package startCluster

import (
	"fmt"
	"stargo/cluster/checkStatus"
	"stargo/module"
	utl "stargo/sr-utl"
	"time"
)

func StartFeCluster() {

	var infoMess string
	//var err error
	//var feStat checkStatus.FeStatusStruct

	// start Fe node one by one
	var tmpUser string
	var tmpKeyRsa string
	var tmpSshHost string
	var tmpSshPort uint32
	var tmpEditLogPort uint32
	var tmpFeDeployDir string

	tmpUser = module.GConfigInfo.Global.User
	tmpKeyRsa = module.GSshPrivateKeyFilePath

	for i := 0; i < len(module.GConfigInfo.FeServers); i++ {
		// for i := 0; i < 1; i++ { ## debug leader node

		tmpSshHost = module.GConfigInfo.FeServers[i].Host
		tmpSshPort = module.GConfigInfo.FeServers[i].SshPort
		tmpEditLogPort = module.GConfigInfo.FeServers[i].EditLogPort
		tmpFeDeployDir = module.GConfigInfo.FeServers[i].DeployDir

		infoMess = fmt.Sprintf("Starting FE node [FeHost = %s, EditLogPort = %d]", tmpSshHost, tmpEditLogPort)
		utl.Logger.Info(infoMess)

		// startFeNode(user string, keyRsa string, sshHost string, sshPort uint32, editLogPort int, feDeployDir string) (err error)
		_ = StartFeNode(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpEditLogPort, tmpFeDeployDir)
		for j := 0; j < 3; j++ {
			portStat, _ := checkStatus.CheckFePortStatus(i)
			if portStat {
				break
				//time.Sleep(10 * time.Second)
			} else {
				time.Sleep(10 * time.Second)
			}
		}
	}
}

func StartFeNode(user string, keyRsa string, sshHost string, sshPort uint32, editLogPort uint32, feDeployDir string) (err error) {

	var infoMess string
	//var isMasterFe bool
	var startFeCmd string

	// check master node
	startFeCmd = fmt.Sprintf("%s/bin/start_fe.sh --daemon", feDeployDir)
	infoMess = fmt.Sprintf("Run starting FE process [host = %s, editLogPort = %d]", sshHost, editLogPort)
	utl.Logger.Debug(infoMess)
	_, err = utl.SshRun(user, keyRsa, sshHost, sshPort, startFeCmd)

	if err != nil {
		infoMess = fmt.Sprintf("Waiting for starting FE node [FeHost = %s]", sshHost)
		utl.Logger.Debug(infoMess)
		return err
	}
	return nil

}

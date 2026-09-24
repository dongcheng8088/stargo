package startCluster

import (
	"fmt"
	"time"

	//    "errors"
	"stargo/cluster/checkStatus"
	"stargo/module"
	utl "stargo/sr-utl"
)

func StartBeCluster() {

	// start Be node one by one
	var infoMess string
	var tmpUser string
	var tmpKeyRsa string
	var tmpSshHost string
	var tmpSshPort uint32
	var tmpHeartbeatServicePort uint32
	var tmpBeDeployDir string

	tmpUser = module.GConfigInfo.Global.User
	tmpKeyRsa = module.GSshPrivateKey

	for i := 0; i < len(module.GConfigInfo.BeServers); i++ {
		// for i := 0; i < 1; i++ { ## debug leader node

		tmpSshHost = module.GConfigInfo.BeServers[i].Host
		tmpSshPort = module.GConfigInfo.BeServers[i].SshPort
		tmpHeartbeatServicePort = module.GConfigInfo.BeServers[i].HeartbeatServicePort
		tmpBeDeployDir = module.GConfigInfo.BeServers[i].DeployDir

		infoMess = fmt.Sprintf("Starting BE node [BeHost = %s, HeartbeatServicePort = %d]", tmpSshHost, tmpHeartbeatServicePort)
		utl.Logger.Info(infoMess)

		_ = StartBeNode(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpHeartbeatServicePort, tmpBeDeployDir)
		for j := 0; j < 3; j++ {
			portStat, _ := checkStatus.CheckBePortStatus(i)
			if portStat {
				break
				//time.Sleep(10 * time.Second)
			} else {
				time.Sleep(10 * time.Second)
			}
		}
	}
}

func StartBeNode(user string, keyRsa string, sshHost string, sshPort uint32, heartbeatServicePort uint32, beDeployDir string) (err error) {
	var infoMess string
	startBeCMD := fmt.Sprintf("%s/bin/start_be.sh --daemon", beDeployDir)
	infoMess = fmt.Sprintf("Starting BE node [host = %s, heartbeatServicePort = %d]", sshHost, heartbeatServicePort)
	utl.Logger.Debug(infoMess)
	_, err = utl.SshRun(user, keyRsa, sshHost, sshPort, startBeCMD)
	if err != nil {
		infoMess = fmt.Sprintf("Waiting for start BE node.[BeHost = %s, Error =  %v", sshHost, err)
		utl.Logger.Debug(infoMess)
		return err
	}
	return nil
}

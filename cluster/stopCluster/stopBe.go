package stopCluster

import (
	"fmt"
	"stargo/cluster/checkStatus"
	"stargo/module"
	utl "stargo/sr-utl"
)

// func startFeNode(user string, keyRsa string, sshHost string, sshPort int, editLogPort int, feDeployDir string) (err error) {

func StopBeNode(user string, keyRsa string, sshHost string, sshPort uint32, beDeployDir string) (err error) {

	var infoMess string
	var stopBeCmd string

	// /opt/starrocks/be/bin/stop_be.sh
	stopBeCmd = fmt.Sprintf("%s/bin/stop_be.sh", beDeployDir)

	infoMess = fmt.Sprintf("Waiting for stoping BE node [BeHost = %s]", sshHost)
	utl.Logger.Info(infoMess)
	_, err = utl.SshRun(user, keyRsa, sshHost, sshPort, stopBeCmd)
	if err != nil {
		infoMess = fmt.Sprintf("Stop BE failed [BeHost = %s, error = %v]", sshHost, err)
		utl.Logger.Debug(infoMess)
		return err
	}
	return nil

}

func StopBeCluster(clusterName string) {

	var infoMess string
	var err error
	var beStat map[string]string

	// Stop BE node one by one
	var tmpUser string
	var tmpKeyRsa string
	var tmpSshHost string
	var tmpSshPort uint32
	var tmpBeDeployDir string
	var tmpHeartbeatServicePort uint32
	//var beStatusList             string

	tmpUser = module.GConfigInfo.Global.User
	tmpKeyRsa = module.GSshPrivateKey

	infoMess = "Stop cluster " + clusterName
	utl.Logger.Info(infoMess)
	for i := 0; i < len(module.GConfigInfo.BeServers); i++ {

		tmpSshHost = module.GConfigInfo.BeServers[i].Host
		tmpSshPort = module.GConfigInfo.BeServers[i].SshPort
		tmpBeDeployDir = module.GConfigInfo.BeServers[i].DeployDir
		tmpHeartbeatServicePort = module.GConfigInfo.BeServers[i].HeartbeatServicePort
		err = StopBeNode(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpBeDeployDir)
		if err != nil {
			infoMess = fmt.Sprintf("Error in stoping BE node [BeHost = %s, HeartbeatServicePort = %d, error = %v]", tmpSshHost, tmpHeartbeatServicePort, err)
			utl.Logger.Debug(infoMess)
		}

		beStat, err = checkStatus.CheckBeStatus(i)

		if err != nil {
			infoMess = fmt.Sprintf("Error in get the Be status [BeHost = %s, HeartbeatServicePort = %d, error = %v]", tmpSshHost, tmpHeartbeatServicePort, err)
			utl.Logger.Debug(infoMess)
		}
		if beStat["Alive"] == "false" {
			infoMess = fmt.Sprintf("The BE node stop succefully [BeHost = %s, HeartbeatServicePort = %d]", tmpSshHost, tmpHeartbeatServicePort)
			utl.Logger.Info(infoMess)
		} else {
			infoMess = fmt.Sprintf("The BE node stop failed [BeHost = %s, HeartbeatServicePort = %d]", tmpSshHost, tmpHeartbeatServicePort)
			utl.Logger.Debug(infoMess)
		}
	}

}

package startCluster

import (
	"fmt"
	"stargo/cluster/checkStatus"
	"stargo/module"
	utl "stargo/sr-utl"
	"strings"
	"time"
)

func InitFeCluster(yamlConf *module.ConfStruct) {
	var infoMess string
	var err error
	var feStat map[string]string

	// start Fe node one by one
	var tmpUser string
	var tmpKeyRsa string
	var tmpSshHost string
	var tmpSshPort uint32
	var tmpEditLogPort uint32
	var tmpQueryPort uint32
	var tmpFeDeployDir string
	var feStatusList string
	var feEntryId int
	tmpUser = yamlConf.Global.User
	tmpKeyRsa = module.GSshPrivateKey

	// get FE entry
	feEntryId, err = checkStatus.GetFeEntry(-1)

	if err != nil || feEntryId == -1 {
		//infoMess = "All FE nodes are down, please start FE node and display the cluster status again."
		//utl.Logger.Warn(infoMess)
		module.SetFeEntry(0)
	} else {
		module.SetFeEntry(feEntryId)
	}

	for i := 0; i < len(module.GConfigInfo.FeServers); i++ {
		// for i := 0; i < 1; i++ { ## debug leader node
		tmpSshHost = module.GConfigInfo.FeServers[i].Host
		tmpSshPort = module.GConfigInfo.FeServers[i].SshPort
		tmpEditLogPort = module.GConfigInfo.FeServers[i].EditLogPort
		tmpQueryPort = module.GConfigInfo.FeServers[i].QueryPort
		tmpFeDeployDir = module.GConfigInfo.FeServers[i].DeployDir

		//infoMess = fmt.Sprintf("Starting FE node [FeHost = %s, FeEditLogPort = %d]", tmpSshHost, tmpEditLogPort)
		//utl.Logger.Info(infoMess)

		for startTimeInd := 0; startTimeInd < 3; startTimeInd++ {
			// initFeNode(user string, keyRsa string, sshHost string, sshPort uint32, editLogPort uint32, feDeployDir string) (err error)
			infoMess = fmt.Sprintf("The %d time to start [%s]", (startTimeInd + 1), tmpSshHost)
			utl.Logger.Debug(infoMess)
			err = InitFeNode(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpEditLogPort, tmpFeDeployDir)
			startWaitTime := time.Duration(20 - startTimeInd*5)
			time.Sleep(startWaitTime * time.Second)

			feStat, err = checkStatus.CheckFeStatus(i)

			if err != nil {
				infoMess = fmt.Sprintf("Error in get the fe status [FeHost = %s, error = %v]", tmpSshHost, err)
				utl.Logger.Debug(infoMess)
			}
			if feStat["Alive"] == "true" {
				infoMess = fmt.Sprintf("The FE node start succefully [host = %s, queryPort = %v]", tmpSshHost, tmpQueryPort)
				utl.Logger.Info(infoMess)
				break
			} else {
				infoMess = fmt.Sprintf("The FE node doesn't start, wait for 10s [FeHost = %s, FeQueryPort = %v, error = %v]", tmpSshHost, tmpQueryPort, err)
				utl.Logger.Warn(infoMess)
			}
		} // FOR-END: 3 time to restart FE node

		if feStat["Alive"] == "false" {
			infoMess = fmt.Sprintf("The FE node start failed [host = %s, queryPort = %v, error = %v]", tmpSshHost, tmpQueryPort, err)
			utl.Logger.Error(infoMess)
		}
		feStatusList = feStatusList + "                                        " + fmt.Sprintf("feHost = %-20sfeQueryPort = %v     feStatus = true\n", tmpSshHost, tmpQueryPort)
	} // FOR-END: list all FE node

	feStatusList = "List all FE status:\n" + feStatusList
	utl.Logger.Info(feStatusList)
}

func InitFeNode(user string, keyRsa string, sshHost string, sshPort uint32, editLogPort uint32, feDeployDir string) (err error) {
	var infoMess string
	//var isMasterFe bool
	var startFeCmd string

	// check master node
	if sshHost == module.GConfigInfo.FeServers[0].Host && editLogPort == module.GConfigInfo.FeServers[0].EditLogPort {
		//isMasterFe = true
		infoMess = fmt.Sprintf("Starting leader FE node [host = %s, editLogPort = %v]",
			module.GConfigInfo.FeServers[0].Host,
			module.GConfigInfo.FeServers[0].EditLogPort)
		utl.Logger.Info(infoMess)
		startFeCmd = fmt.Sprintf("%s/bin/start_fe.sh --daemon", feDeployDir)
		// time.Sleep(30 * time.Second)
	} else {
		infoMess = fmt.Sprintf("Starting follower FE node [host = %s, editLogPort = %v]", sshHost, editLogPort)
		utl.Logger.Info(infoMess)

		startFeCmd = fmt.Sprintf("%s/bin/start_fe.sh --helper %s:%v --daemon", feDeployDir, module.GConfigInfo.FeServers[0].Host, module.GConfigInfo.FeServers[0].EditLogPort)
		// if the start node is follower node, ALTER SYSTEM ADD FOLLOWER "host:editLogPort";
		// func RunSQL(userName string, password string, ip string, port int, dbName string, sqlStat string) (rows *sql.Rows, err error)
		sqlUserName := "root"
		sqlPassword := ""
		sqlIp := module.GConfigInfo.FeServers[0].Host
		sqlPort := module.GFeEntryQueryPort
		sqlDbName := ""
		addFollowerSql := fmt.Sprintf("ALTER SYSTEM ADD FOLLOWER \"%s:%v\"", sshHost, editLogPort)
		_, err := utl.RunSQL(sqlUserName, sqlPassword, sqlIp, sqlPort, sqlDbName, addFollowerSql)
		if err != nil {
			if strings.Contains(err.Error(), "frontend already exists name") {
			} else {
				infoMess = fmt.Sprintf("Error in add follower fe node [FeHost = %s, FeEditLogPort = %v, Error = %v]", sqlIp, editLogPort, err)
				utl.Logger.Error(infoMess)
				return err
			}
		}

	}

	// run feDeploy/bin/start_fe.sh --daemon --helper hsot:edit_log_port
	_, err = utl.SshRun(user, keyRsa, sshHost, sshPort, startFeCmd)
	if err != nil {
		infoMess = fmt.Sprintf("Waiting for starting FE node [FeHost = %s]", sshHost)
		utl.Logger.Debug(infoMess)
		return err
	}
	return nil
}

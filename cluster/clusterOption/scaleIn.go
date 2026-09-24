package clusterOption

import (
	"fmt"
	"time"

	"os"
	"stargo/cluster/checkStatus"
	"stargo/cluster/stopCluster"
	"stargo/module"
	utl "stargo/sr-utl"
)

func ScaleIn(clusterName string, nodeId string) {

	var feEntryId int
	var err error
	var tmpNodeType string
	var nid int
	var dropCmd string
	var infoMess string
	var user string
	var keyRsa string
	var sshHost string
	var sshPort uint32
	var beHeartbeatServicePort uint32
	var feDeployDir string
	var beDeployDir string

	module.InitConf(clusterName, "")

	if checkStatus.CheckClusterName(clusterName) {
		infoMess = "Don't find the Cluster " + clusterName
		utl.Logger.Error(infoMess)
		os.Exit(1)
	}

	tmpNodeType, nid = checkStatus.GetNodeType(nodeId)
	feEntryId, err = checkStatus.GetFeEntry(nid)

	if err != nil || feEntryId == -1 {
		//infoMess = "All FE nodes are down, please start FE node and display the cluster status again."
		//utl.Logger.Warn(infoMess)
		module.SetFeEntry(0)
	} else {
		module.SetFeEntry(feEntryId)
	}

	sqlIp := module.GFeEntryHost
	sqlPort := module.GFeEntryQueryPort
	sqlUserName := "root"
	sqlPassword := ""
	sqlDbName := ""
	user = module.GConfigInfo.Global.User
	keyRsa = module.GSshPrivateKey

	if tmpNodeType == "FE" {

		// stop fe node first
		sshHost = module.GConfigInfo.FeServers[nid].Host
		sshPort = module.GConfigInfo.FeServers[nid].SshPort
		feDeployDir = module.GConfigInfo.FeServers[nid].DeployDir
		err = stopCluster.StopFeNode(user, keyRsa, sshHost, sshPort, feDeployDir)
		if err != nil {
			infoMess = fmt.Sprintf("Error in stop FE node. [nodeId = %s, error = %v]", nodeId, err)
			utl.Logger.Error(infoMess)
			os.Exit(1)
		}

		time.Sleep(time.Duration(10) * time.Second)
		// drop BE node
		dropCmd = fmt.Sprintf("ALTER SYSTEM DROP FOLLOWER '%s'", nodeId)
		_, err := utl.RunSQL(sqlUserName, sqlPassword, sqlIp, sqlPort, sqlDbName, dropCmd)

		if err != nil {
			infoMess = fmt.Sprintf("Error in scale in FE node. [clusterName = %s, nodeId = %s, error = %s]", clusterName, nodeId, err)
			utl.Logger.Error(infoMess)
		}

		// remove FE dir: data, deploy, log
		utl.SshRun(user, keyRsa, sshHost, sshPort, "rm -rf "+module.GConfigInfo.FeServers[nid].LogDir)
		utl.SshRun(user, keyRsa, sshHost, sshPort, "rm -rf "+module.GConfigInfo.FeServers[nid].MetaDir)
		utl.SshRun(user, keyRsa, sshHost, sshPort, "rm -rf "+module.GConfigInfo.FeServers[nid].DeployDir)

		if nid != len(module.GConfigInfo.FeServers)-1 {
			module.GConfigInfo.FeServers = append(module.GConfigInfo.FeServers[:nid], module.GConfigInfo.FeServers[nid+1:]...)
		} else {
			module.GConfigInfo.FeServers = module.GConfigInfo.FeServers[:nid]
		}
		module.WriteBackMeta(module.GConfigInfo, module.GConfigInfo.ClusterInfo.MetaPath)

		infoMess = fmt.Sprintf("Scale in FE node successfully. [clusterName = %s, nodeId = %s]", clusterName, nodeId)
		utl.Logger.Info(infoMess)

	} else if tmpNodeType == "BE" {
		// stop BE node first
		sshHost = module.GConfigInfo.BeServers[nid].Host
		sshPort = module.GConfigInfo.BeServers[nid].SshPort
		sshHost = module.GConfigInfo.BeServers[nid].Host
		beDeployDir = module.GConfigInfo.BeServers[nid].DeployDir
		beHeartbeatServicePort = module.GConfigInfo.BeServers[nid].HeartbeatServicePort
		err = stopCluster.StopBeNode(user, keyRsa, sshHost, sshPort, beDeployDir)

		if err != nil {
			infoMess = fmt.Sprintf("Error in stop BE node. [nodeId = %s, error = %v]", nodeId, err)
			utl.Logger.Error(infoMess)
			os.Exit(1)
		}

		time.Sleep(time.Duration(10) * time.Second)

		// drop BE node
		dropCmd = fmt.Sprintf("ALTER SYSTEM DROP BACKEND '%s:%d'", sshHost, beHeartbeatServicePort)

		_, err := utl.RunSQL(sqlUserName, sqlPassword, sqlIp, sqlPort, sqlDbName, dropCmd)

		if err != nil {
			infoMess = fmt.Sprintf("Error in scale in BE node. [clusterName = %s, nodeId = %s, error = %s]", clusterName, nodeId, err)
			utl.Logger.Error(infoMess)
		}

		// remove dir: data, deploy, log
		utl.SshRun(user, keyRsa, sshHost, sshPort, "rm -rf "+module.GConfigInfo.BeServers[nid].LogDir)
		utl.SshRun(user, keyRsa, sshHost, sshPort, "rm -rf "+module.GConfigInfo.BeServers[nid].StorageDir)
		utl.SshRun(user, keyRsa, sshHost, sshPort, "rm -rf "+module.GConfigInfo.BeServers[nid].DeployDir)

		if nid != len(module.GConfigInfo.BeServers)-1 {
			module.GConfigInfo.BeServers = append(module.GConfigInfo.BeServers[:nid], module.GConfigInfo.BeServers[nid+1:]...)
		} else {
			module.GConfigInfo.BeServers = module.GConfigInfo.BeServers[:nid]
		}
		// module.WriteBackMeta(module.GYamlConf, module.GYamlConf.ClusterInfo.MetaPath)

		module.WriteBackMeta(module.GConfigInfo, module.GConfigInfo.ClusterInfo.MetaPath)
		infoMess = fmt.Sprintf("Scale in BE node successfully. [clusterName = %s, nodeId = %s]", clusterName, nodeId)
		utl.Logger.Info(infoMess)

	} else {
		infoMess = fmt.Sprintf("Error in get Node type. Please check the nodeId. You can use 'sr-ctl-cluster display %s ' to check the node id.[NodeId = %s]", clusterName, nodeId)
		utl.Logger.Error(infoMess)
	}

}

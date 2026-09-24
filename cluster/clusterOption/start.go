package clusterOption

import (
	"fmt"
	"os"
	"stargo/cluster/checkStatus"
	"stargo/cluster/startCluster"
	"stargo/module"
	utl "stargo/sr-utl"
)

func Start(clusterName string, nodeId string, role string) {

	var infoMess string
	var tmpNodeHost string
	var tmpNodePort uint32
	var tmpUser string
	var tmpKeyRsa string
	var tmpSshPort uint32
	var tmpDeployDir string

	module.InitConf(clusterName, "")

	tmpUser = module.GConfigInfo.Global.User
	tmpKeyRsa = module.GSshPrivateKey

	if checkStatus.CheckClusterName(clusterName) {
		infoMess = "Don't find the Cluster " + clusterName
		utl.Logger.Error(infoMess)
		os.Exit(1)
	}
	// start all cluster: sr-ctl-cluster start sr-c1
	// start 1 node: sr-ctl-cluster start sr-c1 --node 192.168.88.33:9010
	// start all FE node: sr-ctl-cluster start sr-c1 --role FE
	// start all BE node: sr-ctl-cluster start sr-c1 --role BE

	// start all cluster: sr-ctl-cluster start sr-c1

	//  -----------------------------------------------------------------------------------------
	//  |  case id   |  node id   |  role     |  option                                         |
	//  -----------------------------------------------------------------------------------------
	//  |  1         | null       |  null     |  start all cluster                              |
	//  |  2         | null       |  !null    |  start FE or BE cluster                         |
	//  |  3         | !null      |  null     |  start the FE/BE node (BE only)                 |
	//  |  4         | !null      |  !null    |  error                                          |
	//  -----------------------------------------------------------------------------------------
	if nodeId == module.NULLSTR && role == module.NULLSTR {
		// case id 1: - start all cluster: sr-ctl-cluster start sr-c1
		startCluster.StartFeCluster()
		startCluster.StartBeCluster()
	} // end of case 1

	if nodeId == module.NULLSTR && role != module.NULLSTR {
		// case id 2: start FE or BE cluster
		if role == "FE" {
			infoMess = "Starting FE cluster ...."
			utl.Logger.Info(infoMess)
			startCluster.StartFeCluster()
		} else if role == "BE" {
			startCluster.StartBeCluster()
			infoMess = "Starting BE cluster ..."
			utl.Logger.Info(infoMess)
		} else {
			infoMess = fmt.Sprintf("Error in get Node type. Please check the nodeId. You can use 'sr-ctl-cluster display %s ' to check the node id.[NodeId = %s]", clusterName, nodeId)
			utl.Logger.Error(infoMess)
		}
	} // end of case 2

	if nodeId != module.NULLSTR && role == module.NULLSTR {
		// case id 3: start the FE/BE node
		// get the node type
		tmpNodeType, i := checkStatus.GetNodeType(nodeId)
		if tmpNodeType == "FE" {
			infoMess = "Please use --role FE to start all the FE node."
			utl.Logger.Error(infoMess)
			os.Exit(1)
			//tmpNodeHost = module.GYamlConf.FeServers[i].Host
			//tmpSshPort = module.GYamlConf.FeServers[i].SshPort
			//tmpNodePort = module.GYamlConf.FeServers[i].EditLogPort
			//tmpDeployDir = module.GYamlConf.FeServers[i].DeployDir
			//startCluster.StartFeNode(tmpUser, tmpKeyRsa, tmpNodeHost, tmpSshPort, tmpNodePort, tmpDeployDir)
		} else if tmpNodeType == "BE" {
			tmpNodeHost = module.GConfigInfo.BeServers[i].Host
			tmpSshPort = module.GConfigInfo.BeServers[i].SshPort
			tmpNodePort = module.GConfigInfo.BeServers[i].HeartbeatServicePort
			tmpDeployDir = module.GConfigInfo.BeServers[i].DeployDir
			infoMess = fmt.Sprintf("Start BE node. [BeHost = %s, HeartbeatServicePort = %d]", tmpNodeHost, tmpNodePort)
			utl.Logger.Info(infoMess)
			startCluster.StartBeNode(tmpUser, tmpKeyRsa, tmpNodeHost, tmpSshPort, tmpNodePort, tmpDeployDir)
		} else {
			infoMess = fmt.Sprintf("Error in get Node type. Please check the nodeId. You can use 'sr-ctl-cluster display %s ' to check the node id.[NodeId = %s]", clusterName, nodeId)
			utl.Logger.Error(infoMess)
		}
	} // end of case 3

	if nodeId != module.NULLSTR && role != module.NULLSTR {
		infoMess = "Detect both --node & --role option."
		utl.Logger.Error(infoMess)
	} // end of case 4

}

/*
func getNodeType(nodeId string) (nodeType string, nodeInd int) {

    // FEID: module.GYamlConf.FeServers[i].EditLogPort, module.GYamlConf.FeServers[i].QueryPort
    // BEID: module.GYamlConf.BeServers[i].Host, module.GYamlConf.BeServers[i].BePort

    tmpNodeId := strings.Split(nodeId, ":")
    fmt.Println("DEBUG>>>>>>>>>>>>>>>>>>>>>>>>>", tmpNodeId)

    // check FE
    for i := 0; i < len(module.GYamlConf.FeServers); i++ {

        if tmpNodeId[0] == module.GYamlConf.FeServers[i].Host &&
	   tmpNodeId[1] == strconv.Itoa(module.GYamlConf.FeServers[i].EditLogPort) {
            nodeType = "FE"
	    //ip = module.GYamlConf.FeServers[i].Host
	    //port = tmpNodeId[1] == module.GYamlConf.FeServers[i].EditLogPort
	    nodeInd = i
	    break
	}
    }

    for i := 0; i < len(module.GYamlConf.BeServers); i++ {

        if tmpNodeId[0] == module.GYamlConf.BeServers[i].Host &&
	   tmpNodeId[1] == strconv.Itoa(module.GYamlConf.BeServers[i].BePort) {
	    nodeType = "BE"
	    //ip = module.GYamlConf.BeServers[i].Host
	    //port = module.GYamlConf.BeServers[i].BePort
	    nodeInd = i
	    break
	}
    }

    return nodeType, nodeInd

}

*/

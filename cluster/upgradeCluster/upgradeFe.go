package upgradeCluster

import (
	"fmt"
	"strings"
	"time"

	//"errors"
	"stargo/cluster/checkStatus"
	"stargo/cluster/startCluster"
	"stargo/cluster/stopCluster"
	"stargo/module"
	utl "stargo/sr-utl"
	//"stargo/cluster/prepareOption"
)

func UpgradeFeCluster() { //(err error){

	var infoMess string
	var err error
	var feEntryId int

	feEntryId, err = checkStatus.GetFeEntry(-1)
	if err != nil || feEntryId == -1 {
		//infoMess = "All FE nodes are down, please start FE node and display the cluster status again."
		//utl.Logger.Warn(infoMess)
		module.SetFeEntry(0)
	} else {
		module.SetFeEntry(feEntryId)
	}

	for i := 0; i < len(module.GConfigInfo.FeServers); i++ {
		infoMess = fmt.Sprintf("Starting upgrade FE node. [feId = %d]", i)
		utl.Logger.Info(infoMess)
		UpgradeFeNode(i)
	}

}

func UpgradeFeNode(feId int) {
	// step 1. backup fe lib
	// step 2. upload new fe lib
	// step 3. stop fe node
	// step 4. start fe node

	var infoMess string
	var user string
	var sourceDir string
	var targetDir string
	var sshHost string
	var sshPort uint32
	var feDeployDir string
	var feEditLogPort uint32
	var keyRsa string
	var feStat map[string]string
	var err error

	user = module.GConfigInfo.Global.User
	keyRsa = module.GSshPrivateKey
	sshHost = module.GConfigInfo.FeServers[feId].Host
	sshPort = module.GConfigInfo.FeServers[feId].SshPort
	feDeployDir = module.GConfigInfo.FeServers[feId].DeployDir
	feEditLogPort = module.GConfigInfo.FeServers[feId].EditLogPort

	// step1. backup fe lib
	sourceDir = fmt.Sprintf("%s/lib", feDeployDir)
	targetDir = fmt.Sprintf("%s/lib.bak-%s", feDeployDir, time.Now().Format("20060102150405"))

	err = utl.RenameDir(user, keyRsa, sshHost, sshPort, sourceDir, targetDir)
	if err != nil {
		infoMess = fmt.Sprintf("Error in rename dir when backup FE lib. [host = %s, sourceDir = %s, targetDir = %s]", sshHost, sourceDir, targetDir)
		utl.Logger.Error(infoMess)
	} else {
		infoMess = fmt.Sprintf("upgrade FE node - backup FE lib. [host = %s, sourceDir = %s, targetDir = %s]", sshHost, sourceDir, targetDir)
		utl.Logger.Info(infoMess)
	}

	// step2. upload new FE lib
	sourceDir = fmt.Sprintf("%s/StarRocks-%s/fe/lib", module.GDownloadPath, strings.Replace(module.GSRVersion, "v", "", -1))
	// sourceDir = fmt.Sprintf("%s/download/StarRocks-%s/fe/lib", module.GSRCtlRoot, strings.Replace(module.GSRVersion, "v", "", -1))
	targetDir = fmt.Sprintf("%s/lib", feDeployDir)
	utl.UploadDir(user, keyRsa, sshHost, sshPort, sourceDir, targetDir)
	infoMess = fmt.Sprintf("upgrade FE node - upload new FE lib. [host = %s, sourceDir = %s, targetDir = %s]", sshHost, sourceDir, targetDir)
	utl.Logger.Info(infoMess)

	// step3. stop FE node
	err = stopCluster.StopFeNode(user, keyRsa, sshHost, sshPort, feDeployDir)
	if err != nil {
		infoMess = fmt.Sprintf("Error in stop FE node when upgrade FE node. [host = %s, feDeployDir = %s]", sshHost, feDeployDir)
		utl.Logger.Error(infoMess)
	} else {
		infoMess = fmt.Sprintf("upgrade FE node - stop FE node. [host = %s, feDeployDir = %s]", sshHost, feDeployDir)
		utl.Logger.Info(infoMess)
	}

	// step4. start FE node
	for j := 0; j < 3; j++ {
		startCluster.StartFeNode(user, keyRsa, sshHost, sshPort, feEditLogPort, feDeployDir)
		infoMess = fmt.Sprintf("upgrade FE node - start FE node. [host = %s, feDeployDir = %s]", sshHost, feDeployDir)
		utl.Logger.Info(infoMess)

		feStat, err = checkStatus.CheckFeStatus(feId)
		if feStat["Alive"] == "true" {
			break
		}
		time.Sleep(10 * time.Second)
	}

	if err != nil {
		infoMess = fmt.Sprintf("Error in get the FE status [feId = %d, error = %v]", feId, err)
		utl.Logger.Debug(infoMess)
	} else if feStat["Alive"] == "false" {
		infoMess = fmt.Sprintf("The FE node upgrade failed. The FE node doesn't work. [feId = %d]\n", feId)
		utl.Logger.Error(infoMess)
	} else if !strings.Contains(feStat["FeVersion"], strings.Replace(module.GSRVersion, "v", "", -1)) {
		infoMess = fmt.Sprintf("The FE node upgrade failed.  [feId = %d, targetVersion = %s, currentVersion = v%s]", feId, module.GSRVersion, feStat["FeVersion"])
		utl.Logger.Error(infoMess)
	} else {
		infoMess = fmt.Sprintf("The Be node upgrade successfully. [feId = %d, currentVersion = v%s]", feId, feStat["FeVersion"])
		utl.Logger.Info(infoMess)
	}

}

package destroyCluster

import (
	"fmt"
	"os"
	"stargo/module"
	utl "stargo/sr-utl"
)

func DestroyCluster(clusterName string) {

	rmFeDir(clusterName)
	rmBeDir(clusterName)
	rmMeta(clusterName)

}

func rmFeDir(clusterName string) {

	var infoMess string
	var tmpFeDeployDir string
	var tmpFeMetaDir string
	var tmpUser string
	var tmpKeyRsa string
	var tmpFeHost string
	var tmpFeSshPort uint32
	var tmpRemoveDeployCmd string
	var tmpRemoveMetaCmd string

	for i := 0; i < len(module.GConfigInfo.FeServers); i++ {

		tmpFeDeployDir = module.GConfigInfo.FeServers[i].DeployDir
		tmpUser = module.GConfigInfo.Global.User
		tmpKeyRsa = module.GSshPrivateKey
		tmpFeMetaDir = module.GConfigInfo.FeServers[i].MetaDir
		tmpFeHost = module.GConfigInfo.FeServers[i].Host
		tmpFeSshPort = module.GConfigInfo.FeServers[i].SshPort
		tmpRemoveDeployCmd = "rm -rf " + tmpFeDeployDir
		tmpRemoveMetaCmd = "rm -rf " + tmpFeMetaDir

		// remove deploy dir
		infoMess = fmt.Sprintf("Waiting for remove FE deploy dir. [FeHost = %s, DeployDir = %s]", tmpFeHost, tmpRemoveDeployCmd)
		utl.Logger.Info(infoMess)
		_, _ = utl.SshRun(tmpUser, tmpKeyRsa, tmpFeHost, tmpFeSshPort, tmpRemoveDeployCmd)

		infoMess = fmt.Sprintf("Waiting for remove FE meta dir. [FeHost = %s, MetaDir = %s]", tmpFeHost, tmpRemoveMetaCmd)
		utl.Logger.Info(infoMess)
		_, _ = utl.SshRun(tmpUser, tmpKeyRsa, tmpFeHost, tmpFeSshPort, tmpRemoveMetaCmd)

		infoMess = fmt.Sprintf("Fe node removed. [FeHost = %s]", tmpFeHost)
		utl.Logger.Info(infoMess)
	}
}

func rmBeDir(clusterName string) {

	var infoMess string
	var tmpBeDeployDir string
	var tmpBeStorageDir string
	var tmpUser string
	var tmpKeyRsa string
	var tmpBeHost string
	var tmpBeSshPort uint32
	var tmpRemoveDeployCmd string
	var tmpRemoveStorageCmd string

	for i := 0; i < len(module.GConfigInfo.BeServers); i++ {

		tmpBeDeployDir = module.GConfigInfo.BeServers[i].DeployDir
		tmpUser = module.GConfigInfo.Global.User
		tmpKeyRsa = module.GSshPrivateKey
		tmpBeStorageDir = module.GConfigInfo.BeServers[i].StorageDir
		tmpBeHost = module.GConfigInfo.BeServers[i].Host
		tmpBeSshPort = module.GConfigInfo.BeServers[i].SshPort
		tmpRemoveDeployCmd = "rm -rf " + tmpBeDeployDir
		tmpRemoveStorageCmd = "rm -rf " + tmpBeStorageDir

		// remove deploy dir
		infoMess = fmt.Sprintf("Waiting for remove BE deploy dir. [BeHost = %s, DeployDir = %s]", tmpBeHost, tmpRemoveDeployCmd)
		utl.Logger.Info(infoMess)
		_, _ = utl.SshRun(tmpUser, tmpKeyRsa, tmpBeHost, tmpBeSshPort, tmpRemoveDeployCmd)

		infoMess = fmt.Sprintf("Waiting for remove BE storage dir. [BeHost = %s, StorageDir = %s]", tmpBeHost, tmpRemoveStorageCmd)
		utl.Logger.Info(infoMess)
		_, _ = utl.SshRun(tmpUser, tmpKeyRsa, tmpBeHost, tmpBeSshPort, tmpRemoveStorageCmd)

		infoMess = fmt.Sprintf("Be node removed. [BeHost = %s]", tmpBeHost)
		utl.Logger.Info(infoMess)
	}
}

func rmMeta(clusterName string) {

	var infoMess string
	var metaFileDir string

	metaFileDir = fmt.Sprintf("%s/cluster/%s", module.GSRCtlRoot, clusterName)
	err := os.RemoveAll(metaFileDir)
	if err != nil {
		infoMess = fmt.Sprintf("Error in remove the meta dir. [Dir = %s]", metaFileDir)
		utl.Logger.Error(infoMess)
	} else {
		infoMess = fmt.Sprintf("Meta Dir removed. [Dir = %s]", metaFileDir)
		utl.Logger.Info(infoMess)
	}

}

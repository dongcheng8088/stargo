package clusterOption

import (
	"os"
	"stargo/cluster/checkStatus"
	"stargo/cluster/destroyCluster"
	"stargo/module"
	utl "stargo/sr-utl"
)

func Destroy(clusterName string) {

	var infoMess string
	module.InitConf(clusterName, "")

	if checkStatus.CheckClusterName(clusterName) {
		infoMess = "Don't find the Cluster " + clusterName
		utl.Logger.Error(infoMess)
		os.Exit(1)
	}

	Stop(clusterName, module.EMPTYSTR, module.EMPTYSTR)
	destroyCluster.DestroyCluster(clusterName)
}

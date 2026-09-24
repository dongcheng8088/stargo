package clusterOption

import (
	"fmt"
	"os"
	"stargo/cluster/checkStatus"
	"stargo/cluster/downgradeCluster"
	"stargo/cluster/prepareOption"
	"stargo/module"
	utl "stargo/sr-utl"
)

func Downgrade(clusterName string, clusterVersion string) {

	var infoMess string
	//var err                error

	module.InitConf(clusterName, "")
	module.SetGlobalVar("GSRVersion", clusterVersion)

	if checkStatus.CheckClusterName(clusterName) {
		infoMess = "Don't find the Cluster " + clusterName
		utl.Logger.Error(infoMess)
		os.Exit(1)
	}

	oldVersion := module.GConfigInfo.ClusterInfo.Version
	newVersion := clusterVersion
	if compareVersions(oldVersion, newVersion) <= 0 {
		infoMess = fmt.Sprintf("OldVersion = %s  NewVersion = %s, the NewVersion is not lower than OldVersion", oldVersion, newVersion)
		utl.Logger.Error(infoMess)
		os.Exit(1)
	} else {
		infoMess = fmt.Sprintf("Downgrade StarRocks Cluster %s, from version %s to version %s", clusterName, oldVersion, newVersion)
		utl.Logger.Info(infoMess)
	}

	prepareOption.PrepareSRPkg()
	downgradeCluster.DowngradeBeCluster()
	downgradeCluster.DowngradeFeCluster()

	module.WriteBackMeta(module.GConfigInfo, module.GConfigInfo.ClusterInfo.MetaPath)

}

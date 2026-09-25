package clusterOption

import (
	"fmt"
	"os"
	"stargo/cluster/checkStatus"
	"stargo/cluster/modifyConfig"
	"stargo/cluster/prepareOption"
	"stargo/cluster/startCluster"
	"stargo/module"
	utl "stargo/sr-utl"
)

func ScaleOut(clusterName string, scaleMetaFile string) {

	var clusterVersion string
	var infoMess string
	// Get the cluster version
	module.AppendConf(clusterName)
	clusterVersion = module.GAppendConfigInfo.ClusterInfo.Version
	module.SetSrVersion(clusterVersion)
	module.InitConf(clusterName, scaleMetaFile)

	if checkStatus.CheckClusterName(clusterName) {
		infoMess = "Don't find the Cluster " + clusterName
		utl.Logger.Error(infoMess)
		os.Exit(1)
	}

	module.GConfigInfo.Global = module.GAppendConfigInfo.Global

	prepareOption.PreCheckSR()
	prepareOption.CreateDir()
	prepareOption.PrepareSRPkg()
	prepareOption.DistributeSrDir()

	tmpYamlConf := module.GAppendConfigInfo
	module.GAppendConfigInfo = module.GConfigInfo
	module.GConfigInfo = tmpYamlConf

	tmpYamlConf.FeServers = append(module.GConfigInfo.FeServers, module.GAppendConfigInfo.FeServers[0:]...)
	tmpYamlConf.BeServers = append(module.GConfigInfo.BeServers, module.GAppendConfigInfo.BeServers[0:]...)
	//fmt.Println("DEBUG >>> tmpYamlConf", tmpYamlConf)
	module.WriteBackMeta(tmpYamlConf, module.GConfigInfo.ClusterInfo.MetaPath)

	//    fmt.Println("DEBUG >>> GYamlConfAppend.FeServers", module.GYamlConfAppend.FeServers)
	//    fmt.Println("################################################")
	//    fmt.Println("DEBUG >>> GYamlConf.FeServers", module.GYamlConf.FeServers)
	//    fmt.Println("################################################")
	//    fmt.Println("DEBUG >>> tmpYamlConf.FeServers", tmpYamlConf.FeServers)

	modifyConfig.ModifyClusterConfig()
	fmt.Println("############################################# SCALE OUT FE CLUSTER #############################################")
	fmt.Println("############################################# SCALE OUT FE CLUSTER #############################################")
	startCluster.InitFeCluster(module.GAppendConfigInfo)
	fmt.Println("############################################# START BE CLUSTER #############################################")
	fmt.Println("############################################# START BE CLUSTER #############################################")
	startCluster.InitBeCluster(module.GAppendConfigInfo)

}

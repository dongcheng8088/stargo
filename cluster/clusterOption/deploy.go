package clusterOption

import (
	"fmt"
	"stargo/cluster/modifyConfig"
	"stargo/cluster/prepareOption"
	"stargo/cluster/startCluster"
	"stargo/module"
	utl "stargo/sr-utl"
)

// sr-ctl-cluster deploy sr-c1 v2.0.1 /tmp/sr-c1.json
func Deploy(clusterName string, clusterVersion string, metaFile string) {
	err := module.InitConf(clusterName, metaFile)
	if err != nil {
		utl.Logger.Error(fmt.Sprintf("初始化元配置文件[%s]失败，错误信息：%s", metaFile, err.Error()))
		return
	}
	module.SetSrVersion(clusterVersion)

	prepareOption.PreCheckSR()
	prepareOption.CreateDir()
	prepareOption.PrepareSRPkg()
	prepareOption.DistributeSrDir()
	module.WriteBackMeta(module.GConfigInfo, module.GWriteBackMetaPath)

	//### recover.sh ##############################")
	modifyConfig.ModifyClusterConfig()

	fmt.Println("############################################# START FE CLUSTER #############################################")
	fmt.Println("############################################# START FE CLUSTER #############################################")

	startCluster.InitFeCluster(module.GConfigInfo)
	fmt.Println("############################################# START BE CLUSTER #############################################")
	fmt.Println("############################################# START BE CLUSTER #############################################")
	startCluster.InitBeCluster(module.GConfigInfo)
	//checkStatus.DeploySuccess()
}

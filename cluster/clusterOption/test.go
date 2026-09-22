package clusterOption

import (
	"stargo/cluster/prepareOption"
	"stargo/module"
)

// sr-ctl-cluster deploy sr-c1 v2.0.1 /tmp/sr-c1.json

func TestOpt() {

	clusterName := "test-sr"
	metaFile := "sr-c1.json"
	module.InitConf(clusterName, metaFile)
	prepareOption.PreCheckSR()

}

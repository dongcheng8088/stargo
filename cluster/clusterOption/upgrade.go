package clusterOption

import (
	"fmt"
	"os"
	"stargo/cluster/checkStatus"
	"stargo/cluster/prepareOption"
	"stargo/cluster/upgradeCluster"
	"stargo/module"
	utl "stargo/sr-utl"
	"strconv"
	"strings"
)

// compareVersions 按数值分段比较两个版本号（如 "v2.1.3" 与 "2.0.1"），
// 自动去除前缀 v/V，按 "." 分割后逐段比较数值。
// 返回 -1 表示 a < b，0 表示 a == b，1 表示 a > b。
func compareVersions(a string, b string) int {

	parse := func(v string) []int {
		v = strings.TrimPrefix(strings.TrimPrefix(v, "v"), "V")
		parts := strings.Split(v, ".")
		nums := make([]int, 0, len(parts))
		for _, p := range parts {
			n, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil {
				n = 0
			}
			nums = append(nums, n)
		}
		return nums
	}

	av := parse(a)
	bv := parse(b)
	n := len(av)
	if len(bv) > n {
		n = len(bv)
	}
	for i := 0; i < n; i++ {
		x, y := 0, 0
		if i < len(av) {
			x = av[i]
		}
		if i < len(bv) {
			y = bv[i]
		}
		if x < y {
			return -1
		}
		if x > y {
			return 1
		}
	}
	return 0
}

func Upgrade(clusterName string, clusterVersion string) {

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
	if compareVersions(oldVersion, newVersion) >= 0 {
		infoMess = fmt.Sprintf("OldVersion = %s  NewVersion = %s, the NewVersion is not higher than OldVersion", oldVersion, newVersion)
		utl.Logger.Error(infoMess)
		os.Exit(1)
	} else {
		infoMess = fmt.Sprintf("Upgrade StarRocks Cluster %s, from version %s to version %s", clusterName, oldVersion, newVersion)
		utl.Logger.Info(infoMess)
	}

	prepareOption.PrepareSRPkg()
	upgradeCluster.UpgradeBeCluster()
	upgradeCluster.UpgradeFeCluster()

	module.WriteBackMeta(module.GConfigInfo, module.GConfigInfo.ClusterInfo.MetaPath)

}

package importCluster

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"stargo/cluster/checkStatus"
	"stargo/module"
	utl "stargo/sr-utl"
	"strconv"
	"strings"
)

func GetFeConf() {

	var infoMess string
	var feHttpUrl string

	for i := 0; i < len(module.GConfigInfo.FeServers); i++ {

		feStat, err := checkStatus.CheckFeStatus(i)
		tmpPort, _ := strconv.Atoi(feStat["HttpPort"])
		module.GConfigInfo.FeServers[i].HttpPort = uint32(tmpPort)
		tmpPort, _ = strconv.Atoi(feStat["RpcPort"])
		module.GConfigInfo.FeServers[i].RpcPort = uint32(tmpPort)
		tmpPort, _ = strconv.Atoi(feStat["EditLogPort"])
		module.GConfigInfo.FeServers[i].EditLogPort = uint32(tmpPort)
		module.GSRVersion = "v" + strings.Split(feStat["Version"], "-")[0]
		rootPasswd := ""

		feHttpUrl = fmt.Sprintf("http://root:%s@%s:%d/variable", rootPasswd, module.GConfigInfo.FeServers[i].Host, module.GConfigInfo.FeServers[i].HttpPort)
		res, err := http.Get(feHttpUrl)
		defer res.Body.Close()
		if err != nil {
			infoMess = fmt.Sprintf("Error in create http get request when get FE conf. [feHttpUrl = %s, error = %v]", feHttpUrl, err)
			utl.Logger.Error(infoMess)
			os.Exit(1)
		}

		robots, err := ioutil.ReadAll(res.Body)
		if err != nil {
			infoMess = fmt.Sprintf("Error in read body.[error = %v]", err)
			utl.Logger.Error(infoMess)
			os.Exit(1)
		}

		//fmt.Println(string(robots))
		// get priority_networks
		r, _ := regexp.Compile("priority_networks=.*")
		module.GConfigInfo.FeServers[i].PriorityNetworks = strings.Replace(r.FindString(string(robots)), "priority_networks=", "", -1)

		// get MetaDir
		r, _ = regexp.Compile("meta_dir=.*")
		module.GConfigInfo.FeServers[i].MetaDir = strings.Replace(r.FindString(string(robots)), "meta_dir=", "", -1)

		// get LogDir
		r, _ = regexp.Compile("sys_log_dir=.*")
		module.GConfigInfo.FeServers[i].LogDir = strings.Replace(r.FindString(string(robots)), "sys_log_dir=", "", -1)
	}

}

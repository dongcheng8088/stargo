package checkStatus

import (
	"errors"
	"fmt"
	"stargo/module"
	utl "stargo/sr-utl"
	"strconv"
	"strings"
)

func GetFeEntry(blackFeNodeId int) (feEntryId int, err error) {

	// get a usable FE host & query port for checking FE/BE status by [show frontends] & [show backends] command

	var infoMess string

	for i := 0; i < len(module.GConfigInfo.FeServers); i++ {
		if i == blackFeNodeId {
			continue
		}

		tmpSshHost := module.GConfigInfo.FeServers[i].Host
		tmpSshPort := module.GConfigInfo.FeServers[i].SshPort
		tmpQueryPort := module.GConfigInfo.FeServers[i].QueryPort
		tmpUser := module.GConfigInfo.Global.User
		tmpKeyRsa := module.GSshPrivateKey
		// check port stat by [netstat -nltp | grep 9030 | grep -v ESTABLISHED]
		cmd := fmt.Sprintf("netstat -an | grep ':%d ' | grep -v ESTABLISHED", tmpQueryPort)

		output, err := utl.SshRun(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, cmd)
		if err != nil {
			infoMess = fmt.Sprintf("Error in get FE entry, checking query port failed. [FeHost = %s, QueryPort = %d, error = %v]", tmpSshHost, tmpQueryPort, err)
			utl.Logger.Debug(infoMess)
		}

		if strings.Contains(string(output), ":"+strconv.FormatUint(uint64(tmpQueryPort), 10)) {
			infoMess = fmt.Sprintf("Get a useable FE entry. [FeID = %d, FeHost = %s, QueryPort = %d]", i, tmpSshHost, tmpQueryPort)
			utl.Logger.Debug(infoMess)
			return i, nil
		}
	}

	err = errors.New("There is no useable FE entry.")
	return -1, err

}

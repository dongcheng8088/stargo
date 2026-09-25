package checkStatus

import (
	"fmt"
	"stargo/module"
	utl "stargo/sr-utl"
	"strconv"
	"strings"
	// "database/sql"
)

// 哪个大聪明在 2.2 版本改了 show backends
// 我真是谢谢你，艹

// var GBeStatArr []BeStatusStruct
func CheckBePortStatus(beId int) (checkPortRes bool, err error) {
	var infoMess string

	tmpUser := module.GConfigInfo.Global.User
	tmpKeyRsa := module.GSshPrivateKeyFilePath
	tmpBeHost := module.GConfigInfo.BeServers[beId].Host
	tmpSshPort := module.GConfigInfo.BeServers[beId].SshPort
	tmpHeartbeatServicePort := module.GConfigInfo.BeServers[beId].HeartbeatServicePort
	checkCMD := fmt.Sprintf("netstat -an | grep ':%d ' | grep -v ESTABLISHED", tmpHeartbeatServicePort)

	output, err := utl.SshRun(tmpUser, tmpKeyRsa, tmpBeHost, tmpSshPort, checkCMD)
	if err != nil {
		infoMess = fmt.Sprintf("Error in run cmd when check BE port status [BeHost = %s, error = %v]", tmpBeHost, err)
		utl.Logger.Debug(infoMess)
		return false, err
	}

	if strings.Contains(string(output), ":"+strconv.FormatUint(uint64(tmpHeartbeatServicePort), 10)) {
		infoMess = fmt.Sprintf("Check the BE query port %s:%d run successfully", tmpBeHost, tmpHeartbeatServicePort)
		utl.Logger.Debug(infoMess)
		return true, nil
	}

	return false, err
}

func GetBeStatJDBC(beId int) (beStatus map[string]string, err error) {
	var infoMess string
	var queryCMD string
	var tmpBeHost string
	var tmpHeartbeatServicePort uint32

	queryCMD = "show backends"
	tmpBeHost = module.GConfigInfo.BeServers[beId].Host
	tmpHeartbeatServicePort = module.GConfigInfo.BeServers[beId].HeartbeatServicePort

	rows, err := utl.RunSQL(module.GJdbcUser, module.GJdbcPasswd,
		module.GConfigInfo.FeServers[0].Host, module.GConfigInfo.FeServers[0].QueryPort,
		module.GJdbcDb, queryCMD)
	if err != nil {
		infoMess = fmt.Sprintf("Error in run sql when check BE status: [BeHost = %s, error = %v]", tmpBeHost, err)
		utl.Logger.Debug(infoMess)
		return beStatus, err
	}

	columns, _ := rows.Columns()
	columnLength := len(columns)
	cache := make([]any, columnLength)

	for index, _ := range cache {
		var tmpVal any
		cache[index] = &tmpVal
	}

	for rows.Next() {
		err = rows.Scan(cache...)
		if err != nil {
			infoMess = fmt.Sprintf("Error in scan sql result [BeHost = %s, error = %v]", tmpBeHost, err)
			utl.Logger.Debug(infoMess)
			return beStatus, err
		}

		beStatus = make(map[string]string)
		for i, data := range cache {
			beStatus[columns[i]] = fmt.Sprintf("%v", data)
		}

		hertbeatPort, _ := strconv.ParseUint(beStatus["HeartbeatPort"], 10, 32)
		if beStatus["IP"] == tmpBeHost && uint32(hertbeatPort) == tmpHeartbeatServicePort {
			return beStatus, err
		}
	}

	return beStatus, err
}

func CheckBeStatus(beId int) (beStat map[string]string, err error) {
	var bePortRun bool
	bePortRun, err = CheckBePortStatus(beId)
	if bePortRun {
		beStat, err = GetBeStatJDBC(beId)
	}
	return beStat, err
}

func TestBeStatus() {
	module.InitConf("sr-c1", "")
	feEntryId, _ := GetFeEntry(-1)
	module.SetFeEntry(feEntryId)
	aaa, _ := CheckBeStatus(0)
	fmt.Println(aaa)
}

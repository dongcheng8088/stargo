package checkStatus

import (
	"fmt"
	"stargo/module"
	utl "stargo/sr-utl"
	"strconv"
	"strings"
)

/*
type FeStatusStruct struct {

    FeName              string
    FeIp                string
    FeEditLogPort       int
    FeHttpPort          int
    FeQueryPort         int
    FeRpcPort           int
    FeRole              string
    FeIsMaster          bool
    FeClusterId         int
    FeJoin              bool
    FeAlive             bool
    FeReplayedJournalId int
    FeLastHeartbeat     sql.NullString
    FeIsHelper          bool
    FeErrMsg            string
    FeStartTime         string
    FeVersion           sql.NullString

}
*/

//var GFeStatusArr []FeStatusStruct

func CheckFePortStatus(feId int) (checkPortRes bool, err error) {

	var infoMess string

	tmpUser := module.GConfigInfo.Global.User
	tmpKeyRsa := module.GSshPrivateKey
	tmpFeHost := module.GConfigInfo.FeServers[feId].Host
	tmpSshPort := module.GConfigInfo.FeServers[feId].SshPort
	tmpQueryPort := module.GConfigInfo.FeServers[feId].QueryPort

	// check Port stat by [netstat -nltp | grep 9030]
	checkCMD := fmt.Sprintf("netstat -an | grep ':%d ' | grep -v ESTABLISHED", tmpQueryPort)
	output, err := utl.SshRun(tmpUser, tmpKeyRsa, tmpFeHost, tmpSshPort, checkCMD)

	if err != nil {
		infoMess = fmt.Sprintf("Error in run cmd when check FE port status [FeHost = %s, error = %v]", tmpFeHost, err)
		utl.Logger.Debug(infoMess)
		return false, err
	}

	if strings.Contains(string(output), ":"+strconv.FormatUint(uint64(tmpQueryPort), 10)) {
		infoMess = fmt.Sprintf("Check the fe query port %s:%d run successfully", tmpFeHost, tmpQueryPort)
		utl.Logger.Debug(infoMess)
		return true, nil
	}

	return false, err
}

func GetFeStatJDBC(feId int) (feStat map[string]string, err error) {

	var infoMess string
	var feStatus map[string]string
	//GJdbcUser = "root"
	//GJdbcPasswd = ""
	//GJdbcDb = ""
	queryCMD := "show frontends"
	tmpFeHost := module.GConfigInfo.FeServers[feId].Host
	tmpQueryPort := module.GConfigInfo.FeServers[feId].QueryPort

	rows, err := utl.RunSQL(module.GJdbcUser, module.GJdbcPasswd, tmpFeHost, tmpQueryPort, module.GJdbcDb, queryCMD)
	if err != nil {
		infoMess = fmt.Sprintf("Error in run sql when check fe status: [FeHost = %s, error = %v]", tmpFeHost, err)
		utl.Logger.Debug(infoMess)
		return feStat, err
	}

	columns, _ := rows.Columns()
	columnLength := len(columns)
	cache := make([]interface{}, columnLength)

	for index, _ := range cache {
		var tmpVal interface{}
		cache[index] = &tmpVal
	}

	for rows.Next() {
		err = rows.Scan(cache...)

		if err != nil {
			infoMess = fmt.Sprintf("Error in scan sql result [FeHost = %s, error = %v]", tmpFeHost, err)
			utl.Logger.Debug(infoMess)
			return feStatus, err
		}

		feStatus = make(map[string]string)
		for i, data := range cache {
			feStatus[columns[i]] = fmt.Sprintf("%s", *data.(*interface{}))
		}

		queryPort, _ := strconv.ParseUint(feStatus["QueryPort"], 10, 32)
		if feStatus["IP"] == tmpFeHost && uint32(queryPort) == tmpQueryPort {
			//GFeStatusArr[feId] = feStat
			return feStatus, nil
		}

	}

	// 循环结束仍未匹配到目标 FE 节点，返回明确错误，避免返回 nil map 被上层误判
	err = fmt.Errorf("FE node not found in 'show frontends' result [FeHost = %s, QueryPort = %d]", tmpFeHost, tmpQueryPort)
	return nil, err
}

func CheckFeStatus(feId int) (feStat map[string]string, err error) {

	//var infoMess    string
	var fePortRun bool
	// CheckFePort
	fePortRun, err = CheckFePortStatus(feId)

	// getFeStat by JDBC
	if fePortRun {
		feStat, err = GetFeStatJDBC(feId)
	}
	return feStat, err
}

func TestFeStatus() {

	module.InitConf("sr-c1", "")
	feEntryId, _ := GetFeEntry(-1)
	module.SetFeEntry(feEntryId)

	aaa, _ := CheckFeStatus(0)
	fmt.Println(aaa)

}

/*
func CheckFeStatus(feId int, user string, keyRsa string, sshHost string, sshPort int, feQueryPort int) (feStat FeStatusStruct, err error) {

    var infoMess string
    var tmpFeStat FeStatusStruct

    // check port stat by [netstat -nltp | grep 9030]
    portStat := CheckFePort(feId)

    cmd := fmt.Sprintf("netstat -an | grep ':%d ' | grep -v ESTABLISHED", feQueryPort)
    output, err := utl.SshRun(user, keyRsa, sshHost, sshPort, cmd)

    if err != nil {
        infoMess = fmt.Sprintf("Error in run cmd when check FE status [FeHost = %s, error = %v]", sshHost, err)
	utl.Logger.Debug(infoMess)
	return feStat, err
    }

    if !strings.Contains(string(output), ":" + strconv.Itoa(feQueryPort)) {
        infoMess = fmt.Sprintf("Check the fe query port %s:%d run failed", sshHost, feQueryPort)
	utl.Logger.Debug(infoMess)
	err = errors.New(infoMess)
	return feStat, err
    }

    // check fe status by jdbc (from the master fe node)
    //RunSQL(userName string, password string, ip string, port int, dbName string, sqlStat string)(rows *sql.Rows, err error)
    feMasterUserName := "root"
    feMasterPassword := ""
    feMasterIP := module.GYamlConf.FeServers[feId].Host
    feMasterQueryPort := module.GYamlConf.FeServers[feId].QueryPort
    feMasterDbName := ""
    sqlStat := "show frontends"
    rows, err := utl.RunSQL(feMasterUserName, feMasterPassword, feMasterIP, feMasterQueryPort, feMasterDbName, sqlStat)
    if err != nil{
        infoMess = fmt.Sprintf(`Error in run sql when check fe status:
                                        feUserName = %s
                                        fePassword = %s
                                        feIP = %s
                                        queryPort = %d
                                        dbName = %s
                                        sqlStat = %s]
                                        error = %v`,
          feMasterUserName, feMasterPassword, feMasterIP, feMasterQueryPort, feMasterDbName, sqlStat, err)
        utl.Logger.Error(infoMess)
	return feStat, err
    }

    for rows.Next(){
	err = rows.Scan(  &tmpFeStat.FeName,
                          &tmpFeStat.FeIp,
                          &tmpFeStat.FeEditLogPort,
                          &tmpFeStat.FeHttpPort,
                          &tmpFeStat.FeQueryPort,
                          &tmpFeStat.FeRpcPort,
                          &tmpFeStat.FeRole,
                          &tmpFeStat.FeIsMaster,
                          &tmpFeStat.FeClusterId,
                          &tmpFeStat.FeJoin,
                          &tmpFeStat.FeAlive,
                          &tmpFeStat.FeReplayedJournalId,
                          &tmpFeStat.FeLastHeartbeat,
                          &tmpFeStat.FeIsHelper,
                          &tmpFeStat.FeErrMsg,
                          &tmpFeStat.FeStartTime,
                          &tmpFeStat.FeVersion)
        if err != nil {
	    infoMess = fmt.Sprintf(`Error in scan sql result:
                                        feUserName = %s
                                        fePassword = %s
                                        feIP = %s
                                        queryPort = %d
                                        dbName = %s
                                        sqlStat = %s]
                                        error = %v`,
                     feMasterUserName, feMasterPassword, feMasterIP, feMasterQueryPort, feMasterDbName, sqlStat, err)
            utl.Logger.Error(infoMess)
	    return feStat, err
	}
        if string(tmpFeStat.FeIp) == sshHost && tmpFeStat.FeQueryPort == feQueryPort {
            feStat = tmpFeStat
            //GFeStatusArr[feId] = feStat
	    return feStat, nil
	}

    }

    return feStat, err

}

*/

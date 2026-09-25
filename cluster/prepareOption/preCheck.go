package prepareOption

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"stargo/module"
	utl "stargo/sr-utl"
	"strconv"
	"strings"
)

// check dir:
//  SRCTLROOT: tmp & download & log - nothing need to precheck
//  FE ssh auth
//  FE user & sudo privileges
//  FE Deploy Dir
//  FE port
//  BE Deploy Dir
//  BE port

const CHECKPASS string = "PASS"
const CHECKFAILED string = "FAILED"

type FePreCheckStruct struct {
	SshAuthRes      string // check ssh auth
	SshAuthInfo     string
	MetaDirRes      string // check meta dir
	MetaDirInfo     string
	DeployDirRes    string // check deploy dir
	DeployDirInfo   string
	HttpPortRes     string // check http port used
	HttpPortInfo    string
	RpcPortRes      string // check rpc port used
	RpcPortInfo     string
	QueryPortRes    string // check query port used
	QueryPortInfo   string
	EditLogPortRes  string // check edit log port used
	EditLogPortInfo string
	OpenFilesRes    string // check open files count
	OpenFilesInfo   string
}

type BePreCheckStruct struct {
	SshAuthRes               string // check ssh auth
	SshAuthInfo              string
	storageDirRes            string // check storageDir
	storageDirInfo           string
	DeployDirRes             string // check deploy dir
	DeployDirInfo            string
	WebServerPortRes         string // check web server port
	WebServerPortInfo        string
	HeartbeatServicePortRes  string // check heartbeat service port
	HeartbeatServicePortInfo string
	BrpcPortRes              string // check brpc port
	BrpcPortInfo             string
	BePortRes                string // check be port
	BePortInfo               string
	OpenFilesRes             string // check open files count
	OpenFilesInfo            string
}

func PreCheckSR() {

	//var preCheckRes bool
	var infoMess string
	var serverId string

	var preCheckFeAdv FePreCheckStruct
	var preCheckBeAdv BePreCheckStruct
	var checkFeMess string
	var checkBeMess string

	tmpMinus := []byte("---------------------------------------")
	fePreCheckStat := preCheckFe()
	bePreCheckStat := preCheckBe()
	infoMess = fmt.Sprintf("PreCheck FE:\n")
	infoMess = infoMess + fmt.Sprintf("%-25s  %-15s  %-30s  %-30s  %-15s  %-15s  %-15s  %-15s  %-15s\n", "server id", "ssh auth", "meta dir", "deploy dir", "http port", "rpc port", "query port", "edit log port", "open files count")
	infoMess = infoMess + fmt.Sprintf("%-25s  %-15s  %-30s  %-30s  %-15s  %-15s  %-15s  %-15s  %-15s\n", tmpMinus[:20], tmpMinus[:15], tmpMinus[:30], tmpMinus[:30], tmpMinus[:15], tmpMinus[:15], tmpMinus[:15], tmpMinus[:15], tmpMinus[:15])
	for i := 0; i < len(fePreCheckStat); i++ {

		serverId = fmt.Sprintf("%s:%d", module.GConfigInfo.FeServers[i].Host, module.GConfigInfo.FeServers[i].EditLogPort)
		infoMess = infoMess + fmt.Sprintf("%-25s  %-15s  %-30s  %-30s  %-15s  %-15s  %-15s  %-15s  %-15s\n",
			serverId,
			fePreCheckStat[i].SshAuthRes,
			fePreCheckStat[i].MetaDirRes,
			fePreCheckStat[i].DeployDirRes,
			fePreCheckStat[i].HttpPortRes,
			fePreCheckStat[i].RpcPortRes,
			fePreCheckStat[i].QueryPortRes,
			fePreCheckStat[i].EditLogPortRes,
			fePreCheckStat[i].OpenFilesRes)

		if fePreCheckStat[i].SshAuthRes != CHECKPASS {
			fmt.Println("DEUBG >>>>>> ssh auth ", i, fePreCheckStat[i].SshAuthRes)
			preCheckFeAdv.SshAuthRes = CHECKFAILED
			preCheckFeAdv.SshAuthInfo = preCheckFeAdv.SshAuthInfo + fePreCheckStat[i].SshAuthInfo + "\n"
		}

		if fePreCheckStat[i].MetaDirRes != CHECKPASS {
			preCheckFeAdv.MetaDirRes = CHECKFAILED
			preCheckFeAdv.MetaDirInfo = preCheckFeAdv.MetaDirInfo + fePreCheckStat[i].MetaDirInfo + "\n"
		}

		if fePreCheckStat[i].DeployDirRes != CHECKPASS {
			preCheckFeAdv.DeployDirRes = CHECKFAILED
			preCheckFeAdv.DeployDirInfo = preCheckFeAdv.DeployDirInfo + fePreCheckStat[i].DeployDirInfo + "\n"
		}

		if fePreCheckStat[i].HttpPortRes != CHECKPASS {
			preCheckFeAdv.HttpPortRes = CHECKFAILED
			preCheckFeAdv.HttpPortInfo = preCheckFeAdv.HttpPortInfo + fePreCheckStat[i].HttpPortInfo + "\n"
		}

		if fePreCheckStat[i].RpcPortRes != CHECKPASS {
			preCheckFeAdv.RpcPortRes = CHECKFAILED
			preCheckFeAdv.RpcPortInfo = preCheckFeAdv.RpcPortInfo + fePreCheckStat[i].RpcPortInfo + "\n"
		}

		if fePreCheckStat[i].QueryPortRes != CHECKPASS {
			preCheckFeAdv.QueryPortRes = CHECKFAILED
			preCheckFeAdv.QueryPortInfo = preCheckFeAdv.QueryPortInfo + fePreCheckStat[i].QueryPortInfo + "\n"
		}

		if fePreCheckStat[i].EditLogPortRes != CHECKPASS {
			preCheckFeAdv.EditLogPortRes = CHECKFAILED
			preCheckFeAdv.EditLogPortInfo = preCheckFeAdv.EditLogPortInfo + fePreCheckStat[i].EditLogPortInfo + "\n"
		}

		if fePreCheckStat[i].OpenFilesRes != CHECKPASS {
			preCheckFeAdv.OpenFilesRes = CHECKFAILED
			preCheckFeAdv.OpenFilesInfo = preCheckFeAdv.OpenFilesInfo + fePreCheckStat[i].OpenFilesInfo + "\n"
		}
	}

	infoMess = infoMess + fmt.Sprintf("\n")
	infoMess = infoMess + fmt.Sprintf("PreCheck BE:\n")

	infoMess = infoMess + fmt.Sprintf("%-25s  %-15s  %-30s  %-30s  %-15s  %-15s  %-15s  %-15s  %-15s\n", "server id", "ssh auth", "storage dir", "deploy dir", "webSer port", "heartbeat port", "brpc port", "be port", "open files count")
	infoMess = infoMess + fmt.Sprintf("%-25s  %-15s  %-30s  %-30s  %-15s  %-15s  %-15s  %-15s  %-15s\n", tmpMinus[:20], tmpMinus[:15], tmpMinus[:30], tmpMinus[:30], tmpMinus[:15], tmpMinus[:15], tmpMinus[:15], tmpMinus[:15], tmpMinus[:15])

	for i := 0; i < len(bePreCheckStat); i++ {
		serverId = fmt.Sprintf("%s:%d", module.GConfigInfo.BeServers[i].Host, module.GConfigInfo.BeServers[i].BePort)
		infoMess = infoMess + fmt.Sprintf("%-25s  %-15s  %-30s  %-30s  %-15s  %-15s  %-15s  %-15s  %-15s\n",
			serverId,
			bePreCheckStat[i].SshAuthRes,
			bePreCheckStat[i].storageDirRes,
			bePreCheckStat[i].DeployDirRes,
			bePreCheckStat[i].WebServerPortRes,
			bePreCheckStat[i].HeartbeatServicePortRes,
			bePreCheckStat[i].BrpcPortRes,
			bePreCheckStat[i].BePortRes,
			bePreCheckStat[i].OpenFilesRes)
		if bePreCheckStat[i].SshAuthRes != CHECKPASS {
			preCheckBeAdv.SshAuthRes = CHECKFAILED
			preCheckBeAdv.SshAuthInfo = preCheckBeAdv.SshAuthInfo + bePreCheckStat[i].SshAuthInfo + "\n"
		}

		if bePreCheckStat[i].storageDirRes != CHECKPASS {
			preCheckBeAdv.storageDirRes = CHECKFAILED
			preCheckBeAdv.storageDirInfo = preCheckBeAdv.storageDirInfo + bePreCheckStat[i].storageDirInfo + "\n"
		}

		if bePreCheckStat[i].DeployDirRes != CHECKPASS {
			preCheckBeAdv.DeployDirRes = CHECKFAILED
			preCheckBeAdv.DeployDirInfo = preCheckBeAdv.DeployDirInfo + bePreCheckStat[i].DeployDirInfo + "\n"
		}

		if bePreCheckStat[i].WebServerPortRes != CHECKPASS {
			preCheckBeAdv.WebServerPortRes = CHECKFAILED
			preCheckBeAdv.WebServerPortInfo = preCheckBeAdv.WebServerPortInfo + bePreCheckStat[i].WebServerPortInfo + "\n"
		}

		if bePreCheckStat[i].HeartbeatServicePortRes != CHECKPASS {
			preCheckBeAdv.HeartbeatServicePortRes = CHECKFAILED
			preCheckBeAdv.HeartbeatServicePortInfo = preCheckBeAdv.HeartbeatServicePortInfo + bePreCheckStat[i].HeartbeatServicePortInfo + "\n"
		}

		if bePreCheckStat[i].BrpcPortRes != CHECKPASS {
			preCheckBeAdv.BrpcPortRes = CHECKFAILED
			preCheckBeAdv.BrpcPortInfo = preCheckBeAdv.BrpcPortInfo + bePreCheckStat[i].BrpcPortInfo + "\n"
		}

		if bePreCheckStat[i].BePortRes != CHECKPASS {
			preCheckBeAdv.BePortRes = CHECKFAILED
			preCheckBeAdv.BePortInfo = preCheckBeAdv.BePortInfo + bePreCheckStat[i].BePortInfo + "\n"
		}

		if bePreCheckStat[i].OpenFilesRes != CHECKPASS {
			preCheckBeAdv.OpenFilesRes = CHECKFAILED
			preCheckBeAdv.OpenFilesInfo = preCheckBeAdv.OpenFilesInfo + bePreCheckStat[i].OpenFilesInfo + "\n"
		}
	}

	infoMess = "PRE CHECK DEPLOY ENV:\n" + infoMess + fmt.Sprintf("\n")
	utl.Logger.Info(infoMess)

	checkFeMess = getFeAdvMess(preCheckFeAdv)
	if checkFeMess != "" {
		checkFeMess = "Please use bellowing promption to fix the issue for FE servers:\n" + checkFeMess
		utl.Logger.Error(checkFeMess)
	}

	checkBeMess = getBeAdvMess(preCheckBeAdv)
	if checkBeMess != "" {
		checkBeMess = "Please use bellowing promption to fix the issue for BE servers:\n" + checkBeMess
		utl.Logger.Error(checkBeMess)
	}

	if strings.Contains(infoMess, CHECKFAILED) {
		infoMess = "PreCheck failed."
		utl.Logger.Error(infoMess)
		os.Exit(1)
	} else {
		infoMess = "PreCheck successfully. RESPECT"
		utl.Logger.Info(infoMess)
	}

}

func getFeAdvMess(preCheckFeAdv FePreCheckStruct) string {

	var checkMess string

	if preCheckFeAdv.SshAuthRes == CHECKFAILED {
		checkMess = checkMess + "Detect no SSH Auth. Use bellowing command to check or fix the issue:\n" + preCheckFeAdv.SshAuthInfo
	}

	if preCheckFeAdv.MetaDirRes == CHECKFAILED {
		checkMess = checkMess + "Detect the FE META FOLDER exist or no privilege. Use bellowing command to check or fix the issue:\n" + preCheckFeAdv.MetaDirInfo
	}

	if preCheckFeAdv.DeployDirRes == CHECKFAILED {
		checkMess = checkMess + "Detect the FE DEPLOY FOLDER exist or no privilege. Use bellowing command to check or fix the issue:\n" + preCheckFeAdv.DeployDirInfo
	}

	if preCheckFeAdv.HttpPortRes == CHECKFAILED {
		checkMess = checkMess + "Detect FE HTTP PORT used. Use bellowing command to check the prot:\n" + preCheckFeAdv.HttpPortInfo
	}
	if preCheckFeAdv.RpcPortRes == CHECKFAILED {
		checkMess = checkMess + "Detect FE RPC PORT used. Use bellowing command to check the prot:\n" + preCheckFeAdv.RpcPortInfo
	}
	if preCheckFeAdv.QueryPortRes == CHECKFAILED {
		checkMess = checkMess + "Detect FE QUERY PORT used. Use bellowing command to check the prot:\n" + preCheckFeAdv.QueryPortInfo
	}

	if preCheckFeAdv.EditLogPortRes == CHECKFAILED {
		checkMess = checkMess + "Detect FE EDIT LOG PORT used. Use bellowing command to check the prot:\n" + preCheckFeAdv.EditLogPortInfo
	}

	if preCheckFeAdv.OpenFilesRes == CHECKFAILED {
		checkMess = checkMess + "Detect FE OPEN FILES COUNT not enought:\n" + preCheckFeAdv.OpenFilesInfo
	}

	return checkMess
}

func getBeAdvMess(preCheckBeAdv BePreCheckStruct) string {

	var checkMess string

	if preCheckBeAdv.SshAuthRes == CHECKFAILED {
		checkMess = checkMess + "Detect no SSH Auth. Use bellowing command to check or fix the issue:\n" + preCheckBeAdv.SshAuthInfo
	}

	if preCheckBeAdv.storageDirRes == CHECKFAILED {
		checkMess = checkMess + "Detect the BE STORAGE FOLDER exist or no privilege. Use bellowing command to check or fix the issue:\n" + preCheckBeAdv.storageDirInfo
	}

	if preCheckBeAdv.DeployDirRes == CHECKFAILED {
		checkMess = checkMess + "Detect the BE DEPLOY FOLDER exist or no privilege. Use bellowing command to check or fix the issue:\n" + preCheckBeAdv.DeployDirInfo
	}

	if preCheckBeAdv.WebServerPortRes == CHECKFAILED {
		checkMess = checkMess + "Detect FE WEB SERVICE PORT used. Use bellowing command to check the prot:\n" + preCheckBeAdv.WebServerPortInfo
	}

	if preCheckBeAdv.HeartbeatServicePortRes == CHECKFAILED {
		checkMess = checkMess + "Detect FE HEARTBEAT SERVICE PORT used. Use bellowing command to check the prot:\n" + preCheckBeAdv.HeartbeatServicePortInfo
	}

	if preCheckBeAdv.BrpcPortRes == CHECKFAILED {
		checkMess = checkMess + "Detect FE BRPC PORT used. Use bellowing command to check the prot:\n" + preCheckBeAdv.BrpcPortInfo
	}

	if preCheckBeAdv.BePortRes == CHECKFAILED {
		checkMess = checkMess + "Detect FE BE PORT used. Use bellowing command to check the prot:\n" + preCheckBeAdv.BePortInfo
	}

	if preCheckBeAdv.OpenFilesRes == CHECKFAILED {
		checkMess = checkMess + "Detect FE OPEN FILES COUNT not enought:\n" + preCheckBeAdv.OpenFilesInfo
	}

	return checkMess
}

func preCheckFe() (fePreCheckRes []FePreCheckStruct) {

	var tmpSshHost string
	var tmpSshPort uint32
	var tmpDetectDir string
	var tmpDetectPort uint32
	var tmpUser string
	var tmpKeyRsa string

	fePreCheckRes = make([]FePreCheckStruct, len(module.GConfigInfo.FeServers))
	tmpUser = module.GConfigInfo.Global.User
	tmpKeyRsa = module.GSshPrivateKey

	for i := 0; i < len(module.GConfigInfo.FeServers); i++ {

		var tmpCheckRes FePreCheckStruct
		tmpSshHost = module.GConfigInfo.FeServers[i].Host
		tmpSshPort = module.GConfigInfo.FeServers[i].SshPort

		// check ssh auth
		tmpCheckRes.SshAuthRes, tmpCheckRes.SshAuthInfo = sshAuth(tmpSshHost, tmpSshPort)

		// check sudo privilege
		// tmpCheckRes.UserSudo = sudoPriv(tmpSshHost, tmpSshPort, tmpUser)

		// check FE deploy folder
		tmpDetectDir = module.GConfigInfo.FeServers[i].DeployDir
		tmpCheckRes.DeployDirRes, tmpCheckRes.DeployDirInfo = dirExist(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpDetectDir, "FE deploy folder")
		//tmpCheckRes.DeployDir = dirPriv(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpDetectDir, "FE deploy folder")

		// check meta folder
		tmpDetectDir = module.GConfigInfo.FeServers[i].MetaDir
		tmpCheckRes.MetaDirRes, tmpCheckRes.MetaDirInfo = dirExist(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpDetectDir, "FE meta folder")
		//tmpCheckRes.MetaDir = dirPriv(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpDetectDir, "FE meta folder")

		// check FE HttpPort
		tmpDetectPort = module.GConfigInfo.FeServers[i].HttpPort
		tmpCheckRes.HttpPortRes, tmpCheckRes.HttpPortInfo = portUsed(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpDetectPort, "FE Http Port")

		// check FE RpcPort
		tmpDetectPort = module.GConfigInfo.FeServers[i].RpcPort
		tmpCheckRes.RpcPortRes, tmpCheckRes.RpcPortInfo = portUsed(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpDetectPort, "FE RPC Port")

		// check FE EditLogPort
		tmpDetectPort = module.GConfigInfo.FeServers[i].EditLogPort
		tmpCheckRes.EditLogPortRes, tmpCheckRes.EditLogPortInfo = portUsed(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpDetectPort, "FE Edit Log Port")

		// check FE QueryPort
		tmpDetectPort = module.GConfigInfo.FeServers[i].QueryPort
		tmpCheckRes.QueryPortRes, tmpCheckRes.QueryPortInfo = portUsed(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpDetectPort, "FE Query Port")

		// check FE open files count
		tmpCheckRes.OpenFilesRes, tmpCheckRes.OpenFilesInfo = openFile(tmpSshHost, tmpSshPort)

		fePreCheckRes[i] = tmpCheckRes
	}

	return fePreCheckRes

}

func preCheckBe() (bePreCheckRes []BePreCheckStruct) {

	var tmpSshHost string
	var tmpSshPort uint32
	var tmpDetectDir string
	var tmpDetectPort uint32
	var tmpUser string
	var tmpKeyRsa string

	bePreCheckRes = make([]BePreCheckStruct, len(module.GConfigInfo.BeServers))
	tmpUser = module.GConfigInfo.Global.User
	tmpKeyRsa = module.GSshPrivateKey

	for i := 0; i < len(module.GConfigInfo.BeServers); i++ {

		var tmpCheckRes BePreCheckStruct
		tmpSshHost = module.GConfigInfo.BeServers[i].Host
		tmpSshPort = module.GConfigInfo.BeServers[i].SshPort

		// check ssh auth
		tmpCheckRes.SshAuthRes, tmpCheckRes.SshAuthInfo = sshAuth(tmpSshHost, tmpSshPort)

		// Check BE deploy user exist
		// tmpCheckRes.UserSudo = sudoPriv(tmpSshHost, tmpSshPort, tmpUser)

		// check BE deploy folder
		tmpDetectDir = module.GConfigInfo.BeServers[i].DeployDir
		tmpCheckRes.DeployDirRes, tmpCheckRes.DeployDirInfo = dirExist(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpDetectDir, "BE deploy folder")

		// check BE storage folder
		tmpDetectDir = module.GConfigInfo.BeServers[i].StorageDir
		tmpCheckRes.storageDirRes, tmpCheckRes.storageDirInfo = dirExist(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpDetectDir, "BE storage folder")

		// check BePort
		tmpDetectPort = module.GConfigInfo.BeServers[i].BePort
		tmpCheckRes.BePortRes, tmpCheckRes.BePortInfo = portUsed(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpDetectPort, "BE Port")

		// check BE WebServerPort
		tmpDetectPort = module.GConfigInfo.BeServers[i].WebServerPort
		tmpCheckRes.WebServerPortRes, tmpCheckRes.WebServerPortInfo = portUsed(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpDetectPort, "BE Web Server Port")

		// check HeartbeatServicePort
		tmpDetectPort = module.GConfigInfo.BeServers[i].HeartbeatServicePort
		tmpCheckRes.HeartbeatServicePortRes, tmpCheckRes.HeartbeatServicePortInfo = portUsed(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpDetectPort, "BE Heartbeat Service Port")

		// check BE brpc port
		tmpDetectPort = module.GConfigInfo.BeServers[i].BrpcPort
		tmpCheckRes.BrpcPortRes, tmpCheckRes.BrpcPortInfo = portUsed(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpDetectPort, "BE brpc Port")

		// check BE open files count
		tmpCheckRes.OpenFilesRes, tmpCheckRes.OpenFilesInfo = openFile(tmpSshHost, tmpSshPort)
		bePreCheckRes[i] = tmpCheckRes

	}

	return bePreCheckRes

}

func dirPriv(user string, keyRsa string, sshHost string, sshPort uint32, dirName string, logStr string) (res string) {
	var infoMess string
	var cmd string
	var dirBase string

	dirBase = path.Dir(dirName)
	// 使用 stat -c '%A %U' 直接输出权限位和属主，两列均不随 locale 变化，
	// 避免原 ls + grep 'd.* .$' 方案依赖单字符目录名和 ls 列顺序的脆弱性。
	cmd = fmt.Sprintf("stat -c '%%A %%U' %s", utl.ShellQuote(dirBase))
	output, _ := utl.SshRun(user, keyRsa, sshHost, sshPort, cmd)
	reg := regexp.MustCompile("\\s+")
	dirStatArr := reg.Split(strings.TrimSpace(string(output)), -1)

	// stat 输出固定两列：index 0 = 权限位（如 drwxr-xr-x），index 1 = 属主用户名。
	// 加长度检查防御：目录不存在或 stat 失败时输出非预期，避免越界 panic。
	if len(dirStatArr) >= 2 && strings.Contains(dirStatArr[0], "drwx") && dirStatArr[1] == user {
		infoMess = fmt.Sprintf("Detect the %-20s don't have create folder privileges [Host = %-20s, Dir = %-30s]\n", logStr, sshHost, dirName)
		utl.Logger.Debug(infoMess)
		return CHECKPASS
	}

	if dirBase != "/" {
		res = dirPriv(user, keyRsa, sshHost, sshPort, dirBase, logStr)
		if res == CHECKPASS {
			return CHECKPASS
		}
	}

	return "Priv failed"
}

func dirExist(user string, keyRsa string, sshHost string, sshPort uint32, dirName string, logStr string) (string, string) {
	var infoMess string
	var cmd string
	var res string = CHECKPASS
	var resExist string = CHECKPASS
	var resPrivs string = CHECKPASS
	var checkMess string

	// 检查目录是否存在
	// 使用 test -d 退出码而非 ls -l 的本地化输出，避免 zh_CN 下 "总计" 导致误判
	cmd = fmt.Sprintf("test -d %s && echo 1 || echo 0", utl.ShellQuote(dirName))
	output, _ := utl.SshRun(user, keyRsa, sshHost, sshPort, cmd)
	if strings.TrimSpace(string(output)) == "1" {
		infoMess = fmt.Sprintf("Detect the %-20s exist [Host = %-20s, Dir = %-30s]\n", logStr, sshHost, dirName)
		utl.Logger.Debug(infoMess)
		resExist = "Dir exist"
	}

	// 检查目录权限是否存在
	resPrivs = dirPriv(user, keyRsa, sshHost, sshPort, dirName, logStr)
	if resPrivs == CHECKPASS && resExist == CHECKPASS {
		res = CHECKPASS
	} else if resPrivs == CHECKPASS && resExist != CHECKPASS {
		// priv ok, dir exist.
		res = fmt.Sprintf("%s: %s", CHECKFAILED, resExist)
		checkMess = fmt.Sprintf("  [Host = %s]  mkdir %s.bak && mv %s/* %s.bak/", sshHost, dirName, dirName, dirName)
	} else if resPrivs != CHECKPASS && resExist == CHECKPASS {
		// dir ok, no priv
		res = fmt.Sprintf("%s: %s", CHECKFAILED, resPrivs)
		checkMess = fmt.Sprintf("  [Host = %s]  chown -R %s %s", sshHost, user, dirName)
	} else if resPrivs != CHECKPASS && resExist != CHECKPASS {
		res = fmt.Sprintf("%s: %s/%s", CHECKFAILED, resExist, resPrivs)
		// dir exist, no priv
		checkMess = fmt.Sprintf("  [Host = %s]  mkdir %s.bak && mv %s/* %s.bak/ && chown -R %s %s", sshHost, dirName, dirName, dirName, user, dirName)
	}

	return res, checkMess
}

func portUsed(user string, keyRsa string, sshHost string, sshPort uint32, detectPort uint32, logStr string) (string, string) {

	var infoMess string
	var checkMess string

	// ss 来自 iproute2，现代 Linux 发行版默认安装；netstat（net-tools）已弃用且常未预装，
	// 原实现因 netstat 不存在导致 output 为空、静默误判端口空闲（false negative）。
	// 用 shell 条件分支输出固定标记，与 locale 无关，且能区分"工具缺失/端口占用/端口空闲"。
	cmd := fmt.Sprintf("if ! command -v ss >/dev/null 2>&1; then echo __NO_SS__; elif ss -tlnp 2>/dev/null | grep -q ':%d '; then echo __USED__; else echo __FREE__; fi", detectPort)
	output, _ := utl.SshRun(user, keyRsa, sshHost, sshPort, cmd)
	switch strings.TrimSpace(string(output)) {
	case "__USED__":
		infoMess = fmt.Sprintf("Detect the %s used [Host = %s, Port = %d]\n", logStr, sshHost, detectPort)
		utl.Logger.Debug(infoMess)
		checkMess = fmt.Sprintf("  [Host = %s]  ss -tlnp | grep ':%d '", sshHost, detectPort)
		return CHECKFAILED, checkMess
	case "__NO_SS__":
		infoMess = fmt.Sprintf("Command 'ss' not found on %s, please install iproute2\n", sshHost)
		utl.Logger.Error(infoMess)
		checkMess = fmt.Sprintf("  [Host = %s]  apt-get install -y iproute2", sshHost)
		return CHECKFAILED, checkMess
	default:
		// __FREE__ 或 SSH 异常导致的空输出，按端口空闲处理
		return CHECKPASS, checkMess
	}
}

func sshAuth(sshHost string, sshPort uint32) (string, string) {

	// check ssh auth
	var infoMess string
	var checkMess string

	keyRsa := module.GSshPrivateKey
	sshUser := module.GConfigInfo.Global.User

	output, _ := utl.SshRun(sshUser, keyRsa, sshHost, sshPort, "date")
	if strings.Contains(string(output), "202") {
		// detect the result has the year 202X, return PASS
		infoMess = fmt.Sprintf("SSH auth check successfully, [host = %s]", sshHost)
		utl.Logger.Debug(infoMess)
		return CHECKPASS, checkMess
	}

	infoMess = fmt.Sprintf("SSH auth check failed, [host = %s]", sshHost)
	utl.Logger.Debug(infoMess)

	checkMess = fmt.Sprintf("  [Host = 127.0.01]  ssh-copy-id %s@%s", sshHost, sshUser)

	return CHECKFAILED, checkMess
}

func openFile(sshHost string, sshPort uint32) (string, string) {

	var infoMess string
	var checkMess string

	keyRsa := module.GSshPrivateKey
	sshUser := module.GConfigInfo.Global.User

	output, err := utl.SshRun(sshUser, keyRsa, sshHost, sshPort, "ulimit -n")
	if err != nil {
		infoMess = fmt.Sprintf("Error in get open file limit. [Host = %s]", sshHost)
		utl.Logger.Error(infoMess)
	}
	openFiles, err := strconv.Atoi(strings.Replace(string(output), "\n", "", -1))

	if err != nil {
		infoMess = fmt.Sprintf("Error in convert the open files count. [host = %s, openFiles = %s, error = %v]", strings.Replace(string(output), "\n", "", -1), err)
		utl.Logger.Debug(infoMess)
		checkMess = fmt.Sprintf("  [Host = %s, User = %s]  Cannot get the open files count. Please use command 'ulimit -n' on user %s to check the copnfiguration.", sshHost, sshUser, sshUser)
		return CHECKFAILED, checkMess
	}

	if openFiles >= 65535 {
		return CHECKPASS, checkMess
	}

	infoMess = fmt.Sprintf("Open files count check failed. Make it more than 65535. [host = %s, openFiles = %d]", sshHost, openFiles)
	utl.Logger.Debug(infoMess)
	checkMess = fmt.Sprintf("  [Host = %s, User = %s]  Please add bellowing line in /etc/security/limits.conf\n%s     soft    nofile          65535\n%s     hard    nofile          65535", sshHost, sshUser, sshUser, sshUser)
	return CHECKFAILED, checkMess

}

func TestPreCheck() {

	module.InitConf("sr-c1", "sr-c1.json")
	module.SetGlobalVar("GSRVersion", "v2.2.0")

	PreCheckSR()
	//aaa := preCheckFe()
	//fmt.Println(aaa)
	//bbb := preCheckBe()
	//fmt.Println(bbb)
}

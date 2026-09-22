package checkStatus

import(
    "strings"
    "fmt"
    "os"
    "strconv"
    "stargo/module"
    "stargo/sr-utl"
)


func CheckClusterName(clusterName string) bool {

    clusterPath := fmt.Sprintf("%s/cluster/%s", module.GSRCtlRoot, clusterName)
    _, err := os.Stat(clusterPath)
    if err != nil {
        // the file doesn't exist, the cluster name can be used
        return true
    }

    return false
}


func GetNodeType(nodeId string) (nodeType string, nodeInd int) {

    // FEID: module.GYamlConf.FeServers[i].EditLogPort, module.GYamlConf.FeServers[i].QueryPort
    // BEID: module.GYamlConf.BeServers[i].Host, module.GYamlConf.BeServers[i].BePort
    var infoMess string
    tmpNodeId := strings.Split(nodeId, ":")

    // nodeId 格式为 host:port，缺少冒号分隔的端口时直接返回，避免数组越界
    if len(tmpNodeId) < 2 {
        infoMess = fmt.Sprintf("Invalid node id, expected format is 'host:port'. [nodeId = %s]", nodeId)
        utl.Logger.Error(infoMess)
        return "", -1
    }

    // check FE
    for i := 0; i < len(module.GYamlConf.FeServers); i++ {

        if tmpNodeId[0] == module.GYamlConf.FeServers[i].Host &&
           tmpNodeId[1] == strconv.Itoa(module.GYamlConf.FeServers[i].EditLogPort) {
            nodeType = "FE"
            //ip = module.GYamlConf.FeServers[i].Host
            //port = tmpNodeId[1] == module.GYamlConf.FeServers[i].EditLogPort
            nodeInd = i
            break
        }
    }

    for i := 0; i < len(module.GYamlConf.BeServers); i++ {

        if tmpNodeId[0] == module.GYamlConf.BeServers[i].Host &&
           tmpNodeId[1] == strconv.Itoa(module.GYamlConf.BeServers[i].BePort) {
            nodeType = "BE"
            //ip = module.GYamlConf.BeServers[i].Host
            //port = module.GYamlConf.BeServers[i].BePort
            nodeInd = i
            break
        }
    }

    infoMess = fmt.Sprintf("Get the node type [nodeid = %s, nodetype = %s, nodeindex = %d]\n", nodeId, nodeType, nodeInd)
    utl.Logger.Debug(infoMess)
    return nodeType, nodeInd

}


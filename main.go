package main

import (
	"flag"
	"fmt"
	"os"
	"stargo/cluster/clusterOption"
	"stargo/playground"
	utl "stargo/sr-utl"
	"strconv"
)

func main() {
	// sr-ctl-cluster deploy    sr-c1   v2.0.1   /tmp/sr-c1.yaml
	// sr-ctl-cluster start     sr-c1
	// sr-ctl-cluster stop      sr-c1
	// sr-ctl-cluster display   sr-c1
	// sr-ctl playground v2.0.1

	var component string
	var command string
	var clusterName string
	var clusterVersion string
	var metaFile string
	var infoMess string
	var node string
	var role string
	var firstArgWithDash int

	// 参数1：组件名称，可选值：playground, cluster
	// 参数2：命令，在参数1为cluster时，可选值：deploy, start, stop, display,
	//      list，destroy，upgrade,downgrade,scale-out, scale-in, import
	// 参数3：集群名称，可选值：sr-c1, sr-c2, ...
	// 参数4：集群版本，就是已部署的集群是StarRocks的哪个版本，例如：v2.0.1, v2.0.2, ...
	// 参数5：元配置文件名称，可以是决定路径，也可以是相对路径，例如：/tmp/sr-c1.yaml, sr-c2.yaml, ...

	if len(os.Args) < 2 {
		utl.Log("ERROR", "参数不足，第一个参数可以是playground、checkport或cluster命令。")
		return
	}

	component = os.Args[1]
	switch component {
	case "playground":
		playground.RunPlayground()

	case "checkport":
		if len(os.Args) < 3 {
			utl.Log("ERROR", "程序启动的第一个参数是checkport时，第二个参数必须给出端口号。")
			return
		}
		// 1. 字符串转 uint32
		u64Val, err := strconv.ParseUint(os.Args[2], 10, 32)
		if err != nil {
			utl.Log("ERROR", "传入参数错误，第二个参数端口号必须是有效的无符号整数。")
			return
		}
		portNo := uint32(u64Val)
		var msg string
		portUsed, err := utl.IsPortUsedTcpOnly(portNo)
		if err != nil {
			msg = fmt.Sprintf("检查端口[%d]是否被占用失败，错误信息：%s", portNo, err.Error())
			utl.Log("ERROR", msg)
		} else {
			if portUsed {
				msg = fmt.Sprintf("端口[%d]已被占用。", portNo)
				utl.Log("INFO", msg)
			} else {
				msg = fmt.Sprintf("端口[%d]未被占用。", portNo)
				utl.Log("INFO", msg)
			}
		}

	case "cluster":
		if len(os.Args) < 3 {
			utl.Log("ERROR", "程序启动的第一个参数是cluster时，第二个参数必须是命令。")
			return
		}
		command = os.Args[2]
		switch command {
		case "deploy":
			if len(os.Args) < 6 {
				utl.Log("ERROR", "deploy命令后的参数依次必须是集群名称、集群版本、元配置文件名称。")
				return
			}
			clusterName = os.Args[3]
			clusterVersion = os.Args[4]
			metaFile = os.Args[5]
			infoMess = fmt.Sprintf("Deploy cluster [clusterName = %s, clusterVersion = %s, metaFile = %s]\n", clusterName, clusterVersion, metaFile)
			utl.Log("OUTPUT", infoMess)
			clusterOption.Deploy(clusterName, clusterVersion, metaFile)
		case "start":
			if len(os.Args) < 4 {
				utl.Log("ERROR", "start命令后的参数必须是集群名称。")
				return
			}
			clusterName = os.Args[3]
			infoMess = fmt.Sprintf("Start cluster [clusterName = %s]", clusterName)
			utl.Log("OUTPUT", infoMess)
			firstArgWithDash = 1
			for i := 1; i < len(os.Args); i++ {
				firstArgWithDash = i
				if len(os.Args[i]) > 0 && os.Args[i][0] == '-' {
					break
				}
			}
			flag.StringVar(&node, "node", "", "The Node ID. Use display command to check the node id.")
			flag.StringVar(&role, "role", "", "The start component type. You can input FE or BE.")
			flag.CommandLine.Parse(os.Args[firstArgWithDash:])
			clusterOption.Start(clusterName, node, role)
		case "stop":
			if len(os.Args) < 4 {
				utl.Log("ERROR", "stop命令后的参数必须是集群名称。")
				return
			}
			clusterName = os.Args[3]
			infoMess = fmt.Sprintf("Stop cluster [clusterName = %s]", clusterName)
			utl.Log("OUTPUT", infoMess)
			firstArgWithDash = 1
			for i := 1; i < len(os.Args); i++ {
				firstArgWithDash = i
				if len(os.Args[i]) > 0 && os.Args[i][0] == '-' {
					break
				}
			}
			flag.StringVar(&node, "node", "", "The Node ID. Use display command to check the node id.")
			flag.StringVar(&role, "role", "", "The start component type. You can input FE or BE.")
			flag.CommandLine.Parse(os.Args[firstArgWithDash:])
			clusterOption.Stop(clusterName, node, role)
		case "display":
			if len(os.Args) < 4 {
				utl.Log("ERROR", "display命令后的参数必须是集群名称。")
				return
			}
			clusterName = os.Args[3]
			infoMess = fmt.Sprintf("Display cluster [clusterName = %s]", clusterName)
			utl.Log("OUTPUT", infoMess)
			clusterOption.Display(clusterName)
		case "list":
			infoMess = fmt.Sprintf("List all clusters")
			utl.Log("OUTPUT", infoMess)
			clusterOption.List()
		case "destroy":
			if len(os.Args) < 4 {
				utl.Log("ERROR", "destroy命令后的参数必须是集群名称。")
				return
			}
			clusterName = os.Args[3]
			infoMess = fmt.Sprintf("Destroy cluster. [ClusterName = %s]", clusterName)
			utl.Log("OUTPUT", infoMess)
			clusterOption.Destroy(clusterName)
		case "upgrade":
			if len(os.Args) < 5 {
				utl.Log("ERROR", "upgrade命令后的参数必须是集群名称、目标版本。")
				return
			}
			clusterName = os.Args[3]
			clusterVersion = os.Args[4]
			infoMess = fmt.Sprintf("Upgrade cluster. [ClusterName = %s, TargetVersion = %s]", clusterName, clusterVersion)
			utl.Log("OUTPUT", infoMess)
			clusterOption.Upgrade(clusterName, clusterVersion)
		case "downgrade":
			if len(os.Args) < 5 {
				utl.Log("ERROR", "downgrade命令后的参数必须是集群名称、目标版本。")
				return
			}
			clusterName = os.Args[3]
			clusterVersion = os.Args[4]
			infoMess = fmt.Sprintf("Downgrade cluster. [ClusterName = %s, TargetVersion = %s]", clusterName, clusterVersion)
			utl.Log("OUTPUT", infoMess)
			clusterOption.Downgrade(clusterName, clusterVersion)
		case "scale-out":
			if len(os.Args) < 5 {
				utl.Log("ERROR", "scale-out命令后的参数必须是集群名称、元配置文件名称。")
				return
			}
			clusterName = os.Args[3]
			metaFile = os.Args[4]
			infoMess = fmt.Sprintf("Scale out cluster. [ClusterName = %s]", clusterName)
			utl.Log("OUTPUT", infoMess)
			clusterOption.ScaleOut(clusterName, metaFile)
		case "scale-in":
			if len(os.Args) < 4 {
				utl.Log("ERROR", "scale-in命令后的参数必须是集群名称。")
				return
			}
			clusterName = os.Args[3]
			firstArgWithDash = -1
			for i := 4; i < len(os.Args); i++ {
				if len(os.Args[i]) > 0 && os.Args[i][0] == '-' {
					firstArgWithDash = i
					break
				}
			}
			if firstArgWithDash == -1 {
				utl.Log("ERROR", "scale-in命令的集群名称参数后面必须是以-开头的参数。")
				return
			}
			flag.StringVar(&node, "node", "", "The Node ID. Use display command to check the node id.")
			flag.CommandLine.Parse(os.Args[firstArgWithDash:])
			infoMess = fmt.Sprintf("Scale in cluster [clusterName = %s, nodeId = %s]", clusterName, node)
			utl.Log("OUTPUT", infoMess)
			clusterOption.ScaleIn(clusterName, node)
		case "import":
			if len(os.Args) < 5 {
				utl.Log("ERROR", "import命令后的参数必须是集群名称、元配置文件名称。")
				return
			}
			clusterName = os.Args[3]
			metaFile = os.Args[4]
			infoMess = fmt.Sprintf("Import the cluster [clusterName = %s, metaFile = %s]", clusterName, metaFile)
			utl.Log("OUTPUT", infoMess)
			clusterOption.ImportCluster(clusterName, metaFile)
		// case "test":
		// utl.Log("OUTPUT", "TEST >>>>>>>>>")
		// checkStatus.TestFeStatus()
		//prepareOption.TestPreCheck()
		//prepareOption.PreCheckSR()
		//playground.DeployPlayground()
		default:
			infoMess = fmt.Sprintf("cluster参数后的命令必须是 %s 之一，命令 %s 无效。",
				"deploy start stop display list destroy upgrade downgrade scale-out scale-in import", command)
			utl.Log("ERROR", infoMess)
		}
	default:
		infoMess = "程序启动时的第一个参数必须是playground或cluster.\n"
		fmt.Printf(infoMess)
		utl.Log("ERROR", infoMess)
	}
}

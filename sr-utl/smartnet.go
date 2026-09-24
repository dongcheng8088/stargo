package utl

import (
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v3/net"
)

// IsPortUsed 检查本机指定端口是否被使用
func IsPortUsed(port uint32) (used bool, err error) {
	// kind = "all" = 所有连接
	// kind = "tcp" = TCP连接
	// kind = "tcp4" = TCP4连接
	// kind = "tcp6" = TCP6连接
	// kind = "udp" = UDP连接
	// kind = "udp4" = UDP4连接
	// kind = "udp6" = UDP6连接
	// kind = "unix" = UNIX连接
	// kind = "inet" = INET连接
	// kind = "inet4" = INET4连接
	// kind = "inet6" = INET6连接
	conns, err := net.Connections("all")
	if err != nil {
		return false, err
	}

	for _, c := range conns {
		if c.Laddr.Port == port {
			return true, nil
		}
	}
	return false, nil
}

func IsPortUsedTcpOnly(port uint32) (used bool, err error) {
	conns, err := net.ConnectionsWithoutUids("tcp")
	if err != nil {
		return false, err
	}

	for _, c := range conns {
		if c.Laddr.Port == port && c.Status == "LISTEN" {
			return true, nil
		}
	}
	return false, nil
}

// IsRemotePortUsed 通过 SSH 检查远程主机指定端口是否被使用。
// 在远程执行 `ss -tun sport = :PORT` 列出源端口为目标端口的所有 TCP/UDP 连接，
// 若输出中存在数据行（非表头）则表示该端口已被占用（监听或已建立连接）。
//
// 参数 sshPort 为远程 SSH 服务端口，targetPort 为待检查的业务端口。
func IsRemotePortUsed(user, keyFile, host string, sshPort uint32, targetPort uint32) (used bool, err error) {
	// ss -tun 列出所有 TCP/UDP 连接（含监听与已建立）；sport = :PORT 过滤源端口
	cmd := fmt.Sprintf("ss -tun sport = :%d", targetPort)
	output, err := SshRun(user, keyFile, host, sshPort, cmd)
	if err != nil {
		return false, err
	}

	// ss 输出首行为表头（Netid State Recv-Q Send-Q Local Address:Port Peer Address:Port）
	// 若仅含表头或输出为空，则端口未被使用
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Netid") {
			continue
		}
		return true, nil
	}
	return false, nil
}

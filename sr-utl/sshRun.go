package utl

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// warnInsecureHostKeyOnce 保证主机密钥校验被禁用的风险警告只输出一次
var warnInsecureHostKeyOnce sync.Once

func NewConfig(keyFile string, user string) (config *ssh.ClientConfig, err error) {
	key, err := os.ReadFile(keyFile)
	if err == nil {
		signer, err := ssh.ParsePrivateKey(key)
		if err == nil {
			config = &ssh.ClientConfig{
				User: user,
				Auth: []ssh.AuthMethod{
					ssh.PublicKeys(signer),
				},
				// TODO: 建议改为基于 ~/.ssh/known_hosts 的主机密钥校验
				// （golang.org/x/crypto/ssh/knownhosts），当前禁用主机密钥校验存在中间人攻击风险。
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}
			warnInsecureHostKeyOnce.Do(func() {
				Logger.Warn("SSH host key verification is disabled (InsecureIgnoreHostKey), the connection is vulnerable to man-in-the-middle attacks.")
			})
			return config, nil
		}
	}
	return nil, err
}

func sshRun(config *ssh.ClientConfig, host string, port uint32, command string) (outPut []byte, err error) {
	client, err := ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", host, port),
		config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	outPut, err = session.CombinedOutput(command)
	if err != nil {
		Logger.Debug(fmt.Sprintf("on [%s:%d] run cmd %s failed: %v, output: %s", host, port, command, err, string(outPut)))
		return outPut, err
	}

	Logger.Debug(fmt.Sprintf("on [%s:%d] run cmd %s result: %s", host, port, command, string(outPut)))
	return outPut, nil
}

func SshRun(user string, keyFile string, host string, port uint32, command string) (outPut []byte, err error) {
	var innerError error

	sshConfig, innerError := NewConfig(keyFile, user)
	if innerError != nil {
		err = fmt.Errorf("Failed to new the ssh config when call SshRun %v", innerError)
		return nil, err
	}

	output, innerError := sshRun(sshConfig, host, port, command)
	if innerError != nil {
		err = fmt.Errorf("Failed to run command on %s:%d, cmd %s, error = %v",
			host, port, command, innerError)
		return nil, err
	}

	return output, nil
}

// sftpConnect 建立 SSH 连接并返回 SFTP 客户端。
// 当 sftp.NewClient 失败时，会关闭已建立的 sshClient，避免连接泄漏。
func sftpConnect(config *ssh.ClientConfig, host string, port uint32) (sftpClient *sftp.Client, err error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}
	sftpClient, err = sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, err
	}
	return sftpClient, nil
}

func uploadFile(sftpClient *sftp.Client, localFileName string, remoteFileName string) (err error) {
	srcFile, err := os.Open(localFileName)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := sftpClient.Create(remoteFileName)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	fileStat, err := os.Stat(localFileName)
	if err != nil {
		return err
	}

	err = sftpClient.Chmod(remoteFileName, fileStat.Mode())
	return err
}

func uploadDirectory(sftpClient *sftp.Client, localPath string, remotePath string) (err error) {
	localFiles, err := os.ReadDir(localPath)
	if err == nil {
		for _, backupDir := range localFiles {
			localFilePath := path.Join(localPath, backupDir.Name())
			remoteFilePath := path.Join(remotePath, backupDir.Name())
			if backupDir.IsDir() {
				// Mkdir 在目录已存在时会返回错误，此时应继续递归上传，而非中断
				mkdirErr := sftpClient.Mkdir(remoteFilePath)
				if mkdirErr != nil && !os.IsExist(mkdirErr) {
					err = mkdirErr
					break
				}
				err = uploadDirectory(sftpClient, localFilePath, remoteFilePath)
			} else {
				localFileName := path.Join(localPath, backupDir.Name())
				remoteFileName := path.Join(remotePath, backupDir.Name())
				err = uploadFile(sftpClient, localFileName, remoteFileName)
			}
			if err != nil {
				break
			}
		}
	}
	return err
}

func UploadFile(user string, keyFile string, host string, port uint32, sourceFile string, targetFile string) (err error) {
	sshConfig, innerError := NewConfig(keyFile, user)
	if innerError != nil {
		err = fmt.Errorf("Failed to new config [keyfile = %s, user = %s]: %s",
			keyFile, user, innerError.Error())
		return err
	}

	sftpClient, innerError := sftpConnect(sshConfig, host, port)
	if innerError != nil {
		err = fmt.Errorf("Failed to connect sftp client [keyfile = %s, user = %s, host = %s, port = %032d]: %s",
			keyFile, user, host, port, innerError.Error())
		return err
	}
	defer sftpClient.Close()

	innerError = uploadFile(sftpClient, sourceFile, targetFile)
	if innerError != nil {
		err = fmt.Errorf("Failed to upload file [user = %s, keyFile = %s, host = %s, port = %032d, sourceFile = %s, targetFile = %s]: %s",
			user, keyFile, host, port, sourceFile, targetFile, innerError.Error())
		return err
	}

	return nil
}

func UploadDir(user string, keyFile string, host string, port uint32, sourceDir string, targetDir string) (err error) {
	var innerError error

	// 确保目标目录存在（mkdir -p 是幂等的，目录已存在时无操作）
	cmd := fmt.Sprintf("mkdir -p %s", shellQuote(targetDir))
	_, innerError = SshRun(user, keyFile, host, port, cmd)
	if innerError != nil {
		infoMess := fmt.Sprintf("Error in create folder [%s] on [%s:%032d]", targetDir, host, port)
		Logger.Error(infoMess)
		err = fmt.Errorf("%s: %v", infoMess, innerError)
		return err
	}

	sshConfig, innerError := NewConfig(keyFile, user)
	if innerError != nil {
		err = fmt.Errorf("Failed to new config [keyfile = %s, user = %s]: %s",
			keyFile, user, innerError.Error())
		return err
	}

	sftpClient, innerError := sftpConnect(sshConfig, host, port)
	if innerError != nil {
		err = fmt.Errorf("Failed to connect sftp client [host = %s, port = %032d, sourceDir = %s, targetDir = %s]: %s",
			host, port, sourceDir, targetDir, innerError.Error())
		return err
	}
	defer sftpClient.Close()

	innerError = uploadDirectory(sftpClient, sourceDir, targetDir)
	if innerError != nil {
		err = fmt.Errorf("Error in uploadDirectory: user = %s, keyFile = %s, host = %s, port = %032d, sourceDir = %s, targetDir = %s, error = %v",
			user, keyFile, host, port, sourceDir, targetDir, innerError)
		return err
	}

	return nil
}

func RenameDir(user string, keyFile string, host string, port uint32, sourceDir string, targetDir string) (err error) {
	var innerError error

	cmd := fmt.Sprintf("ls -- %s", shellQuote(sourceDir))
	_, innerError = SshRun(user, keyFile, host, port, cmd)
	if innerError != nil && !os.IsExist(innerError) {
		err = fmt.Errorf("The source dir [%s] on [%s:%032d] doesn't exist", sourceDir, host, port)
		return err
	}

	sshConfig, innerError := NewConfig(keyFile, user)
	if innerError != nil {
		err = fmt.Errorf("Failed to new config [keyfile = %s, user = %s]: %s",
			keyFile, user, innerError.Error())
		return err
	}

	sftpClient, innerError := sftpConnect(sshConfig, host, port)
	if innerError != nil {
		err = fmt.Errorf("Failed to connect sftp client [host = %s, port = %032d, sourceDir = %s, targetDir = %s]: %s",
			host, sourceDir, targetDir, innerError.Error())
		return err
	}
	defer sftpClient.Close()

	innerError = sftpClient.Rename(sourceDir, targetDir)
	if innerError != nil {
		err = fmt.Errorf("Failed to rename dir [host = %s, sourceDir = %s, targetDir = %s]: %s",
			host, sourceDir, targetDir, innerError.Error())
		return err
	}

	return nil
}

func RemoveDir(user string, keyFile string, host string, port uint32, dirName string) (err error) {
	var innerError error

	// 使用 rm -rf 可删除非空目录，且对不存在的目录返回成功（幂等）
	cmd := fmt.Sprintf("rm -rf %s", shellQuote(dirName))
	_, innerError = SshRun(user, keyFile, host, port, cmd)
	if innerError != nil && !os.IsExist(innerError) {
		err = fmt.Errorf("Failed to remove directory %s on [%s:%032d]: %s", dirName, host, port, innerError.Error())
		return err
	}

	return nil
}

// shellQuote 将字符串转义为可安全传入 shell 单引号上下文的形式，
// 防止路径中的特殊字符（空格、分号、引号等）导致命令注入。
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

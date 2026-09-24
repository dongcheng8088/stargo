package prepareOption

import (
	"fmt"
	"stargo/module"
	utl "stargo/sr-utl"
)

func CreateDir() {

	var infoMess string

	infoMess = "Create the deploy folder ..."
	utl.Logger.Info(infoMess)
	CreateiSrCtlDir()
	CreateFeDir()
	CreateBeDir()

}

func CreateiSrCtlDir() {

	// create SrCtlDir
	// SRCTLROOT/{download,tmp}

	utl.MkDir(module.GSRCtlRoot + "/tmp")
	utl.MkDir(module.GSRCtlRoot + "/download")
	utl.MkDir(module.GSRCtlRoot + "/cluster")

}

func CreateFeDir() {

	var infoMess string
	var errMess string
	var cmd string
	var err error
	//var outPut []byte

	sshUser := module.GConfigInfo.Global.User
	sshKeyRsaFile := module.GSshPrivateKey

	for i := 0; i < len(module.GConfigInfo.FeServers); i++ {
		sshHost := module.GConfigInfo.FeServers[i].Host
		sshPort := module.GConfigInfo.FeServers[i].SshPort

		// create DEPLOY dir for FE nodes
		cmd = fmt.Sprintf("mkdir -p %s", module.GConfigInfo.FeServers[i].DeployDir)
		infoMess = fmt.Sprintf("Create DEPLOY Folder for FE node: %s@%s:%d \"%s\"", sshUser, sshHost, sshPort, cmd)
		utl.Logger.Debug(infoMess)

		_, err = utl.SshRun(sshUser, sshKeyRsaFile, sshHost, sshPort, cmd)
		if err != nil {
			errMess = fmt.Sprintf("ERROR in creating DEPLOY folder for FE node: %s@%s:%d \"%s\"", sshUser, sshHost, sshPort, cmd)
			utl.Logger.Error(errMess)
			panic(err)
		}

		// create META dir for FE nodes
		cmd = fmt.Sprintf("mkdir -p %s", module.GConfigInfo.FeServers[i].MetaDir)
		infoMess = fmt.Sprintf("Create META Folder for FE node: %s@%s:%d \"%s\"", sshUser, sshHost, sshPort, cmd)
		utl.Logger.Debug(infoMess)

		_, err = utl.SshRun(sshUser, sshKeyRsaFile, sshHost, sshPort, cmd)
		if err != nil {
			errMess = fmt.Sprintf("ERROR in creating META folder for FE node: %s@%s:%d \"%s\"", sshUser, sshHost, sshPort, cmd)
			utl.Logger.Error(errMess)
			panic(err)
		}

		if module.GConfigInfo.FeServers[i].DeployDir+"/meta" != module.GConfigInfo.FeServers[i].MetaDir {
			cmd = fmt.Sprintf("ln -s %s %s", module.GConfigInfo.FeServers[i].MetaDir, module.GConfigInfo.FeServers[i].DeployDir+"/meta")
			infoMess = fmt.Sprintf("Detect MetaDir isn't under DeployDir, Create the soft link, CMD %s", cmd)
			utl.Logger.Debug(infoMess)
			_, err := utl.SshRun(sshUser, sshKeyRsaFile, sshHost, sshPort, cmd)
			if err != nil {
				errMess = fmt.Sprintf("Error in create soft link for MetaDir, CMD %s", cmd)
				utl.Logger.Error(errMess)
				panic(err)
			}
		}

		// create LOG dir for FE nodes
		cmd = fmt.Sprintf("mkdir -p %s", module.GConfigInfo.FeServers[i].LogDir)
		infoMess = fmt.Sprintf("Create LOG Folder for FE node: %s@%s:%d \"%s\"", sshUser, sshHost, sshPort, cmd)
		utl.Logger.Debug(infoMess)

		_, err = utl.SshRun(sshUser, sshKeyRsaFile, sshHost, sshPort, cmd)
		if err != nil {
			errMess = fmt.Sprintf("ERROR in creating LOG folder for FE node: %s@%s:%d \"%s\"", sshUser, sshHost, sshPort, cmd)
			utl.Logger.Error(errMess)
			panic(err)
		}
	}
}

func CreateBeDir() {
	var infoMess string
	var errMess string
	var cmd string
	var err error
	//var outPut []byte

	sshUser := module.GConfigInfo.Global.User
	sshKeyRsaFile := module.GSshPrivateKey

	for i := 0; i < len(module.GConfigInfo.BeServers); i++ {
		sshHost := module.GConfigInfo.BeServers[i].Host
		sshPort := module.GConfigInfo.BeServers[i].SshPort

		// create DEPLOY dir for BE nodes
		cmd = fmt.Sprintf("mkdir -p %s", module.GConfigInfo.BeServers[i].DeployDir)
		infoMess = fmt.Sprintf("Create DEPLOY Folder for BE node: %s@%s:%d \"%s\"", sshUser, sshHost, sshPort, cmd)
		utl.Logger.Debug(infoMess)

		_, err = utl.SshRun(sshUser, sshKeyRsaFile, sshHost, sshPort, cmd)
		if err != nil {
			errMess = fmt.Sprintf("ERROR in creating DEPLOY folder for BE node: %s@%s:%d \"%s\"", sshUser, sshHost, sshPort, cmd)
			utl.Logger.Error(errMess)
			panic(err)
		}

		// create STORAGE dir for BE nodes
		cmd = fmt.Sprintf("mkdir -p %s", module.GConfigInfo.BeServers[i].StorageDir)
		infoMess = fmt.Sprintf("Create Storage Folder for BE node: %s@%s:%d \"%s\"", sshUser, sshHost, sshPort, cmd)
		utl.Logger.Debug(infoMess)

		_, err = utl.SshRun(sshUser, sshKeyRsaFile, sshHost, sshPort, cmd)
		if err != nil {
			errMess = fmt.Sprintf("ERROR in creating STORAGE folder for BE node: %s@%s:%d \"%s\"", sshUser, sshHost, sshPort, cmd)
			utl.Logger.Error(errMess)
			panic(err)
		}

		if module.GConfigInfo.BeServers[i].DeployDir+"/storage" != module.GConfigInfo.BeServers[i].StorageDir {
			cmd = fmt.Sprintf("ln -s %s %s", module.GConfigInfo.BeServers[i].StorageDir, module.GConfigInfo.BeServers[i].DeployDir+"/storage")
			infoMess = fmt.Sprintf("Detect StorageDir isn't under DeployDir, Create the soft link, CMD %s", cmd)
			utl.Logger.Debug(infoMess)
			_, err := utl.SshRun(sshUser, sshKeyRsaFile, sshHost, sshPort, cmd)
			if err != nil {
				errMess = fmt.Sprintf("Error in create soft link for StorageDir, CMD %s", cmd)
				utl.Logger.Error(errMess)
				panic(err)
			}
		}

		// create LOG dir for BE nodes
		cmd = fmt.Sprintf("mkdir -p %s", module.GConfigInfo.BeServers[i].LogDir)
		infoMess = fmt.Sprintf("Create LOG Folder for BE node: %s@%s:%d \"%s\"", sshUser, sshHost, sshPort, cmd)
		utl.Logger.Debug(infoMess)

		_, err = utl.SshRun(sshUser, sshKeyRsaFile, sshHost, sshPort, cmd)
		if err != nil {
			errMess = fmt.Sprintf("ERROR in creating LOG folder for BE node: %s@%s:%d \"%s\"", sshUser, sshHost, sshPort, cmd)
			utl.Logger.Error(errMess)
			panic(err)
		}

	}

}

package main

import (
	"os"
	"strings"
	"io/ioutil"
	"path/filepath"
	"fmt"
	"os/exec"
)

var dockerImage = "pbacterio/openstackclient-docker"

var varsPointingFiles = map[string]bool{
	"OS_CACERT": true,
	"OS_CERT":   true,
	"OS_KEY":    true,
}

func main() {
	dockerArgs := []string{"run", "--rm", "-it",}

	varsToFiles := make([]string, 0, len(varsPointingFiles))

	for _, e := range os.Environ() {
		envVar := strings.Split(e, "=")[0]
		if varsPointingFiles[envVar] {
			varsToFiles = append(varsToFiles, envVar)
			continue
		}
		if strings.HasPrefix(envVar, "OS") {
			dockerArgs = append(dockerArgs, "-e", e)
		}
	}

	if len(varsToFiles) > 0 {
		tmpVol, err := ioutil.TempDir("", "openstack-docker")
		if err != nil {
			panic(err)
		}
		dockerArgs = append(dockerArgs, "-v", tmpVol+":/certs_vol")
		for _, varName := range varsToFiles {
			tmpDir, err := ioutil.TempDir(tmpVol, "")
			if err != nil {
				panic(err)
			}
			srcPath := os.Getenv(varName)
			err = copyFile(srcPath, filepath.Join(tmpDir, filepath.Base(srcPath)))
			if err != nil {
				panic(err)
			}
			dockerArgs = append(dockerArgs, "-e",
				fmt.Sprintf("%v=/certs_vol/%v/%v", varName, filepath.Base(tmpDir), filepath.Base(srcPath)))
		}
		defer os.RemoveAll(tmpVol)
	}

	dockerArgs = append(dockerArgs, dockerImage, "openstack")
	dockerArgs = append(dockerArgs, os.Args[1:]...)

	cmd := exec.Command("docker", dockerArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func copyFile(src string, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, data, 0644)
}

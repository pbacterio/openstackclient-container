package main

import (
	"os"
	"strings"
	"path/filepath"
	"fmt"
	//"os/exec"
	"math/rand"
	"time"
	"os/exec"
)

var dockerImage = "pbacterio/openstackclient-docker"

var varsPointingFiles = map[string]bool{
	"OS_CACERT":   true,
	"--os-cacert": true,
	"OS_CERT":     true,
	"--os-cert":   true,
	"OS_KEY":      true,
	"--os-key":    true,
}

func main() {
	files := map[string]string{}
	dockerArgs := []string{}

	// Get enviroment variables
	for _, env := range os.Environ() {
		envParts := strings.SplitN(env, "=", 2)
		name, value := envParts[0], envParts[1]
		if !strings.HasPrefix(name, "OS_") {
			continue
		}
		if !varsPointingFiles[name] {
			dockerArgs = append(dockerArgs, "-e", name+":"+value)
			continue
		}
		mapped_value := files[value]
		if mapped_value == "" {
			mapped_value = "/" + randName() + filepath.Base(value)
			files[value] = mapped_value
			dockerArgs = append([]string{"-v", value + ":" + mapped_value}, dockerArgs...)
		}
		dockerArgs = append(dockerArgs, "-e", name+":"+mapped_value)
	}

	dockerArgs = append(dockerArgs, "pbacterio/openstackclient-docker", "openstack")

	// Get args-vars
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if !varsPointingFiles[arg] || len(os.Args) < i+2 {
			dockerArgs = append(dockerArgs, arg)
			continue
		}
		mapped_value := files[os.Args[i+1]]
		if mapped_value == "" {
			mapped_value = "/" + randName() + filepath.Base(os.Args[i+1])
			files[os.Args[i+1]] = mapped_value
			dockerArgs = append([]string{"-v", os.Args[i+1] + ":" + mapped_value}, dockerArgs...)
		}
		dockerArgs = append(dockerArgs, arg, mapped_value)
		i++
	}

	// Run openstackclient container
	dockerArgs = append([]string{"run", "--rm", "-it"}, dockerArgs...)
	fmt.Println(dockerArgs)
	cmd := exec.Command("docker", dockerArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func randName() string {
	return fmt.Sprintf("%X", rand.Uint64())
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

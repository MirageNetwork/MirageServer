package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	// 获取当前脚本所在的路径
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("获取当前路径出错")
		os.Exit(1)
	}
	dir := filepath.Dir(filename)

	// 切换到项目的根目录
	err := os.Chdir(filepath.Join(dir, ".."))
	if err != nil {
		fmt.Printf("切换工作目录出错: %v\n", err)
		os.Exit(1)
	}

	buildFrontEnd("cockpit_web")
	buildFrontEnd("console_web")

	cmd := exec.Command("go", "run", "tailscale.com/cmd/mkversion")
	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("执行mkversion命令出错: %v\n", stderr.String())
		os.Exit(1)
	}
	lines := strings.Split(out.String(), "\n")

	out.Reset()
	stderr.Reset()

	vars := make(map[string]string)
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			vars[parts[0]] = parts[1]
		}
	}

	ldflags := "-X tailscale.com/version.longStamp=" +
		strings.Trim(vars["VERSION_LONG"], `"`) + " -X tailscale.com/version.shortStamp=" +
		strings.Trim(vars["VERSION_SHORT"], `"`) + " -X tailscale.com/version.gitCommitStamp=" +
		strings.Trim(vars["VERSION_GIT_HASH"], `"`) + " -X tailscale.com/version.extraGitCommitStamp=" +
		strings.Trim(vars["VERSION_EXTRA_HASH"], `"`)

	fmt.Println(ldflags)

	cmd = exec.Command("go", "build", "-ldflags", ldflags, "-o", "dist/dist/MirageServer_"+strings.Trim(vars["VERSION_LONG"], `"`))
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("构建项目出错,错误详情: %s\n", stderr.String())
		os.Exit(1)
	}
}

func buildFrontEnd(frontName string) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前工作目录出错: %v\n", err)
		os.Exit(1)
	}
	err = os.Chdir(filepath.Join(dir, frontName))
	if err != nil {
		fmt.Printf("切换工作目录出错: %v\n", err)
		os.Exit(1)
	}
	cmd := exec.Command("npm", "install")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("error happen during npm install under %s: %s\n", frontName, stderr.String())
		os.Exit(1)
	}
	cmd = exec.Command("npm", "run", "build")
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("error happen during npm runbuild under %s: %s\n", frontName, stderr.String())
		os.Exit(1)
	}
	err = os.Chdir(dir)
	if err != nil {
		fmt.Printf("切换工作目录出错: %v\n", err)
		os.Exit(1)
	}
}

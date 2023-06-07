package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestCommand(t *testing.T) {
	cmd := exec.Command("E:\\Tools\\nginx-1.25.0\\nginx.exe")
	cmd.Dir = "E:\\Tools\\nginx-1.25.0\\"

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()

	if err != nil {
		panic(fmt.Errorf("启动命令时出错: %s", err))
	}
	err = cmd.Wait()
	fmt.Println("结束")
	if err != nil {
		panic(fmt.Errorf("等待命令完成时出错：%s", err))
	}
}

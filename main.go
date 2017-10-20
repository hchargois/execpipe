package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func die(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func main() {
	var cmd1, cmd2 []string
	for i, arg := range os.Args[1:] {
		if arg == "|" {
			cmd1 = os.Args[1 : i+1]
			cmd2 = os.Args[i+2:]
		}
	}
	if len(cmd1) == 0 || len(cmd2) == 0 {
		die("Usage: %s cmd [args...] '|' cmd [args...]\n", os.Args[0])
	}

	cmd1Exec, err := exec.LookPath(cmd1[0])
	if err != nil {
		die(err.Error())
	}
	cmd2Exec, err := exec.LookPath(cmd2[0])
	if err != nil {
		die(err.Error())
	}

	pipe := make([]int, 2)
	err = syscall.Pipe(pipe)
	if err != nil {
		die("Error creating pipe: %v\n", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		die("Error retrieving working directory: %v\n", err)
	}

	env := os.Environ()

	procAttr := &syscall.ProcAttr{
		Dir:   wd,
		Env:   env,
		Files: []uintptr{uintptr(pipe[0]), os.Stdout.Fd(), os.Stderr.Fd()},
	}

	_, err = syscall.ForkExec(cmd2Exec, cmd2, procAttr)
	if err != nil {
		die("Failed to execute command %v: %v\n", cmd2, err)
	}

	syscall.Close(1)
	syscall.Dup2(pipe[1], 1)
	err = syscall.Exec(cmd1Exec, cmd1, env)
	if err != nil {
		die("Failed to execute command %v: %v\n", cmd1, err)
	}
}

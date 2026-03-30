package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Segment struct {
	Cmd string
	Op  string
}

type CommandSpec struct {
	Args    []string
	Input   string
	Output  string
	IsEmpty bool
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	sigintCh := make(chan os.Signal, 1)
	signal.Notify(sigintCh, os.Interrupt)

	var currentPGID int

	go func() {
		for range sigintCh {
			if currentPGID != 0 {
				_ = syscall.Kill(-currentPGID, syscall.SIGINT)
				fmt.Println()
			} else {
				fmt.Println()
				fmt.Print("myshell> ")
			}
		}
	}()

	for {
		fmt.Print("myshell> ")

		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println()
				return
			}
			fmt.Fprintln(os.Stderr, "read error:", err)
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		segments := splitByLogicalOps(line)

		lastSuccess := true
		for i, seg := range segments {
			if i > 0 {
				switch seg.Op {
				case "&&":
					if !lastSuccess {
						continue
					}
				case "||":
					if lastSuccess {
						continue
					}
				}
			}

			ok, pgid := executeSegment(seg.Cmd)
			currentPGID = pgid
			lastSuccess = ok
			currentPGID = 0
		}
	}
}

func splitByLogicalOps(line string) []Segment {
	var result []Segment
	var current strings.Builder
	var lastOp string

	for i := 0; i < len(line); i++ {
		if i+1 < len(line) {
			two := line[i : i+2]
			if two == "&&" || two == "||" {
				part := strings.TrimSpace(current.String())
				if part != "" {
					result = append(result, Segment{
						Cmd: part,
						Op:  lastOp,
					})
				}
				current.Reset()
				lastOp = two
				i++
				continue
			}
		}
		current.WriteByte(line[i])
	}

	part := strings.TrimSpace(current.String())
	if part != "" {
		result = append(result, Segment{
			Cmd: part,
			Op:  lastOp,
		})
	}

	return result
}

func executeSegment(line string) (bool, int) {
	parts := splitPipeline(line)
	if len(parts) == 0 {
		return true, 0
	}

	if len(parts) == 1 {
		spec := parseCommand(parts[0])
		if spec.IsEmpty {
			return true, 0
		}

		// builtin без pipeline
		if isBuiltin(spec.Args[0]) {
			err := runBuiltin(spec)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return false, 0
			}
			return true, 0
		}

		err, pgid := runExternal([]CommandSpec{spec})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return false, pgid
		}
		return true, pgid
	}

	var specs []CommandSpec
	for _, p := range parts {
		spec := parseCommand(p)
		if spec.IsEmpty {
			continue
		}
		if len(spec.Args) == 0 {
			continue
		}
		if isBuiltin(spec.Args[0]) {
			fmt.Fprintf(os.Stderr, "builtin %q inside pipeline is not supported\n", spec.Args[0])
			return false, 0
		}
		specs = append(specs, spec)
	}

	if len(specs) == 0 {
		return true, 0
	}

	err, pgid := runExternal(specs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false, pgid
	}
	return true, pgid
}

func splitPipeline(line string) []string {
	raw := strings.Split(line, "|")
	var result []string
	for _, part := range raw {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func parseCommand(part string) CommandSpec {
	fields := strings.Fields(part)
	if len(fields) == 0 {
		return CommandSpec{IsEmpty: true}
	}

	var args []string
	var input string
	var output string

	for i := 0; i < len(fields); i++ {
		switch fields[i] {
		case "<":
			if i+1 < len(fields) {
				input = expandEnv(fields[i+1])
				i++
			}
		case ">":
			if i+1 < len(fields) {
				output = expandEnv(fields[i+1])
				i++
			}
		default:
			args = append(args, expandEnv(fields[i]))
		}
	}

	if len(args) == 0 {
		return CommandSpec{IsEmpty: true}
	}

	return CommandSpec{
		Args:    args,
		Input:   input,
		Output:  output,
		IsEmpty: false,
	}
}

func expandEnv(s string) string {
	return os.ExpandEnv(s)
}

func isBuiltin(name string) bool {
	switch name {
	case "cd", "pwd", "echo", "kill", "ps", "exit":
		return true
	default:
		return false
	}
}

func runBuiltin(spec CommandSpec) error {
	switch spec.Args[0] {
	case "cd":
		return builtinCD(spec.Args[1:])
	case "pwd":
		return builtinPWD(spec.Output)
	case "echo":
		return builtinEcho(spec.Args[1:], spec.Output)
	case "kill":
		return builtinKill(spec.Args[1:])
	case "ps":
		return builtinPS(spec.Output)
	case "exit":
		os.Exit(0)
	}
	return nil
}

func builtinCD(args []string) error {
	var dir string
	if len(args) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cd: cannot determine home directory: %w", err)
		}
		dir = home
	} else {
		dir = args[0]
	}
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("cd: %w", err)
	}
	return nil
}

func builtinPWD(output string) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("pwd: %w", err)
	}
	return writeOutput(dir+"\n", output)
}

func builtinEcho(args []string, output string) error {
	return writeOutput(strings.Join(args, " ")+"\n", output)
}

func builtinKill(args []string) error {
	if len(args) != 1 {
		return errors.New("kill: usage: kill <pid>")
	}

	pid, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("kill: invalid pid: %w", err)
	}

	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("kill: %w", err)
	}
	return nil
}

func builtinPS(output string) error {
	cmd := exec.Command("ps", "-e", "-o", "pid,ppid,comm")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("ps: %w", err)
	}
	return writeOutput(string(out), output)
}

func writeOutput(text, outputFile string) error {
	if outputFile == "" {
		fmt.Print(text)
		return nil
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(text)
	return err
}

func runExternal(specs []CommandSpec) (error, int) {
	cmds := make([]*exec.Cmd, 0, len(specs))

	for _, spec := range specs {
		if len(spec.Args) == 0 {
			continue
		}
		cmd := exec.Command(spec.Args[0], spec.Args[1:]...)
		cmd.Stderr = os.Stderr
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmds = append(cmds, cmd)
	}

	if len(cmds) == 0 {
		return nil, 0
	}

	for i := 0; i < len(cmds)-1; i++ {
		r, w := io.Pipe()
		cmds[i].Stdout = w
		cmds[i+1].Stdin = r
	}

	if specs[0].Input != "" {
		f, err := os.Open(specs[0].Input)
		if err != nil {
			return fmt.Errorf("input redirect: %w", err), 0
		}
		defer f.Close()
		cmds[0].Stdin = f
	} else if cmds[0].Stdin == nil {
		cmds[0].Stdin = os.Stdin
	}

	last := len(cmds) - 1
	if specs[last].Output != "" {
		f, err := os.Create(specs[last].Output)
		if err != nil {
			return fmt.Errorf("output redirect: %w", err), 0
		}
		defer f.Close()
		cmds[last].Stdout = f
	} else if cmds[last].Stdout == nil {
		cmds[last].Stdout = os.Stdout
	}

	for i := 0; i < len(cmds)-1; i++ {
		if cmds[i].Stdin == nil && i == 0 {
			cmds[i].Stdin = os.Stdin
		}
	}

	var pgid int
	for i, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("start %q: %w", strings.Join(specs[i].Args, " "), err), pgid
		}

		if i == 0 {
			pgid = cmd.Process.Pid
		}
	}

	for i := 1; i < len(cmds); i++ {
		_ = syscall.Setpgid(cmds[i].Process.Pid, pgid)
	}

	for i := 0; i < len(cmds)-1; i++ {
		if w, ok := cmds[i].Stdout.(*io.PipeWriter); ok {
			_ = w.Close()
		}
		if r, ok := cmds[i+1].Stdin.(*io.PipeReader); ok {
			_ = r.Close()
		}
	}

	var finalErr error
	for _, cmd := range cmds {
		err := cmd.Wait()
		if err != nil {
			finalErr = err
		}
	}

	time.Sleep(10 * time.Millisecond)

	return finalErr, pgid
}

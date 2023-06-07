package plugin_manager

import (
	"agent/logger"
	program_service "agent/service"
	"agent/util"
	"errors"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"time"
)

type Process struct {
	program   *Program
	cmd       *exec.Cmd
	startTime time.Time
	stopTime  time.Time
	state     int
	// true if process is starting
	inStart bool
	// true if the process is stopped by user
	stopByUser bool
	retryTimes int
}

const (
	// Stopped the stopped state
	Stopped = 0

	// Starting the starting state
	Starting = 10

	// Running the running state
	Running = 20

	// Backoff the backoff state
	Backoff = 30

	// Stopping the stopping state
	Stopping = 40

	// Exited the Exited state
	Exited = 50

	// Fatal the Fatal state
	Fatal = 60

	// Unknown the unknown state
	Unknown = 70
)

type Program struct {
	Name            string `mapstructure:"name"`
	Directory       string `mapstructure:"directory"`
	Command         string `mapstructure:"command"`
	IsAutoRestart   bool   `mapstructure:"isAutoRestart"`
	IsAutoStart     bool   `mapstructure:"isAutoStart"`
	MaxRestartCount int    `mapstructure:"MaxRestartCount"`
	Process         *Process
}

var currentPrograms []*Program

func Reload(programs []*Program) {
	checkAndRemove(programs)
	addPrograms := computesAddPrograms(programs, currentPrograms)
	removePrograms := computesRemovePrograms(programs, currentPrograms)
	restartPrograms := computesRestartPrograms(programs, currentPrograms)
	currentPrograms = programs
	for _, p := range addPrograms {
		p.start()
	}
	for _, p := range removePrograms {
		p.stop()
	}
	for _, p := range restartPrograms {
		p.stop()
		p.start()
	}
}

func SendProgramChangeMsg() {
	programRss := make([]program_service.ProgramRs, len(currentPrograms), len(currentPrograms))
	for index, program := range currentPrograms {
		programRss[index] = program_service.ProgramRs{Name: program.Name,
			Directory:     program.Directory,
			Command:       program.Command,
			IsAutoStart:   program.IsAutoStart,
			IsAutoRestart: program.IsAutoRestart,
			Pid:           program.Process.cmd.Process.Pid,
			StartTime:     program.Process.startTime,
			StopTime:      program.Process.stopTime,
			State:         program.Process.state,
			StopByUser:    program.Process.stopByUser,
		}
	}
	program_service.SendProgramChangeRequest(programRss)
}

func (p *Program) start() {
	logger.Logger.Info("Start the program.", zap.String("name", p.Name))
	cmd := p.startProcess()
	p.updateProgramToStart(cmd)
}

func (p *Program) occurredStopEvent() {
	logger.Logger.Info("A stop program event occurred", zap.String("name", p.Name))
	p.updateProgramToStop()
	SendProgramChangeMsg()
	if p.IsAutoRestart {
		// 进行重新启动
		logger.Logger.Info("Do a reboot", zap.String("name", p.Name))
		p.start()
		SendProgramChangeMsg()
	}
}

func (p *Program) updateProgramToStart(cmd *exec.Cmd) {
	if p.Process != nil {
		p.Process.startTime = time.Now()
		p.Process.state = Running
		p.Process.inStart = true
		p.Process.retryTimes++
		return
	}
	// 第一次启动，创建Process
	process := new(Process)
	process.cmd = cmd
	process.startTime = time.Now()
	process.stopTime = time.Time{}
	process.state = Running
	process.inStart = true
	process.stopByUser = false
	process.retryTimes = 0
	// 给进程赋值
	p.Process = process
}

func (p *Program) updateProgramToStop() {
	p.Process.cmd = nil
	p.Process.startTime = time.Time{}
	p.Process.stopTime = time.Now()
	p.Process.state = Stopped
	p.Process.inStart = false
	p.Process.stopByUser = false
}

func (p *Program) check() error {
	if len(p.Directory) < 0 {
		return errors.New("the file directory length must be greater than 0")
	}
	if len(p.Command) < 0 {
		return errors.New("command length must be greater than 0")
	}
	return nil
}

func checkAndRemove(programs []*Program) {
	for i := 0; i < len(programs); i++ {
		if err := programs[i].check(); err != nil || programs[i].IsAutoStart {
			programs = append(programs[:i], programs[i+1:]...)
			i--
		}
	}
}

func computesRestartPrograms(new, old []*Program) []*Program {
	var restart []*Program
	for _, n := range new {
		found := false
		for _, o := range old {
			if n.Name == o.Name && (n.Directory != o.Directory || n.Command != o.Command) {
				found = true
				break
			}
		}
		if found {
			restart = append(restart, n)
		}
	}
	return restart
}

func computesAddPrograms(new, old []*Program) []*Program {
	return computesDifference(new, old)
}

func computesRemovePrograms(new, old []*Program) []*Program {
	return computesDifference(old, new)
}

func computesDifference(new, old []*Program) []*Program {
	if new == nil {
		return nil
	} else if old == nil {
		return new
	}
	var diff []*Program
	for _, n := range new {
		found := false
		for _, o := range old {
			if n.Name == o.Name {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, n)
		}
	}
	return diff
}

func (p *Program) stop() {
	err := p.Process.cmd.Process.Kill()
	if err != nil {
		logger.Logger.Error("An error occurred while stopping the program.", zap.Error(err))
	}
}

func (p *Program) startProcess() *exec.Cmd {
	completeCommand := util.AppendPathSeparator(p.Directory) + p.Command
	cmd := exec.Command(completeCommand)
	channel := make(chan *exec.Cmd)
	go func() {
		cmd.Dir = p.Directory
		// 设置标准输出和标准错误输出
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		// 启动命令
		err := cmd.Start()
		if err != nil {
			logger.Logger.Panic("There was an error starting the program.", zap.String("name", p.Name), zap.Error(err))
		}
		channel <- cmd
		err = cmd.Wait()
		if err != nil {
			logger.Logger.Info("Program abort.", zap.String("name", p.Name), zap.Error(err))
		}
		p.occurredStopEvent()
	}()
	return <-channel
}

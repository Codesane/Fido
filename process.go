package main

import (
	"bufio"
	"io"
	"log"
	"os/exec"
	"sync"
)

type OutputSource string

const StdOutSource OutputSource = "StdOut"
const StdErrSource OutputSource = "StdErr"

type ProcessOutput struct {
	Message string
	Source  OutputSource
}

type ExitReason struct {
	ExitCode int
}

type Process struct {
	Output           <-chan ProcessOutput
	ExitReason       *ExitReason // Possibly nil if unknown after exit
	workingDirectory string
	cmd              string
	args             []string

	internalProcOutputCh chan<- ProcessOutput
	command              *exec.Cmd
}

func NewProcess(workingDirectory, cmd string, args []string) *Process {
	processOutputCh := make(chan ProcessOutput, 1024)

	return &Process{
		Output:               processOutputCh,
		ExitReason:           nil,
		workingDirectory:     workingDirectory,
		cmd:                  cmd,
		args:                 args,
		internalProcOutputCh: processOutputCh,
	}
}

func (p *Process) Start() error {
	p.command = exec.Command(p.cmd, p.args...)
	p.command.Dir = p.workingDirectory

	pStdErr, err := p.command.StderrPipe()
	if err != nil {
		close(p.internalProcOutputCh)
		return err
	}

	pStdOut, err := p.command.StdoutPipe()
	if err != nil {
		close(p.internalProcOutputCh)
		return err
	}

	if err := p.command.Start(); err != nil {
		close(p.internalProcOutputCh)
		return err
	}

	stdErr := readProcessOutputAsync(pStdErr, StdErrSource)
	stdOut := readProcessOutputAsync(pStdOut, StdOutSource)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for msg := range stdOut {
			p.internalProcOutputCh <- msg
		}
	}()

	go func() {
		defer wg.Done()
		for errMsg := range stdErr {
			p.internalProcOutputCh <- errMsg
		}
	}()

	go func() {
		if err := p.command.Wait(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				p.ExitReason = &ExitReason{
					ExitCode: exitError.ExitCode(),
				}
			}
		} else {
			p.ExitReason = &ExitReason{
				ExitCode: 0,
			}
		}

		wg.Wait()

		close(p.internalProcOutputCh)
	}()

	return err
}

func (p *Process) Stop() error {
	return p.command.Process.Kill()
}

func readProcessOutputAsync(from io.Reader, source OutputSource) <-chan ProcessOutput {
	out := make(chan ProcessOutput)
	bufferedReader := bufio.NewReader(from)

	go func() {
		for {
			lineBytes, _, err := bufferedReader.ReadLine()
			if err != nil {
				if err.Error() != "EOF" {
					log.Printf("Error reading from process: %v", err.Error())
				}
				close(out)
				break
			}

			out <- ProcessOutput{
				Message: string(lineBytes),
				Source:  source,
			}
		}
	}()

	return out
}

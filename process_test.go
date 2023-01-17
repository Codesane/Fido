package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStartProcess(t *testing.T) {
	p := NewProcess("./testFiles", "python3", []string{"printStdErrStdOut.py"})

	err := p.Start()
	assert.NoError(t, err)

	var outputs []ProcessOutput
	for output := range p.Output {
		outputs = append(outputs, output)
	}
	assert.Len(t, outputs, 2, "Expected len(outputs) == 2 but was '%d'", len(outputs))

	expectedStdErr := ProcessOutput{
		Message: "StdErr Write",
		Source:  StdErrSource,
	}
	assert.Containsf(t, outputs, expectedStdErr, "Expected to find StdErr output '%v' in '%v'", outputs, expectedStdErr)

	expectedStdOut := ProcessOutput{
		Message: "StdOut Write",
		Source:  StdOutSource,
	}
	assert.Containsf(t, outputs, expectedStdOut, "Expected to find StdOut output '%v' in '%v'", outputs, expectedStdOut)
}

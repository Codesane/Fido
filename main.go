package main

import (
	bytes2 "bytes"
	"container/ring"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const TailedOutputBufferSize = 8
const RetryStartApplicationAfter = 3 * time.Second
const SlackWebHookEnvVarName = "FIDO_SLACK_WEBHOOK"

type ProcessExecutionReport struct {
	Uptime              time.Duration
	ExitReason          *ExitReason
	TotalLinesOutputted int
	TailedOutput        [TailedOutputBufferSize]string
}

type SlackMessage struct {
	Text string `json:"text"`
}

func main() {
	args := os.Args[1:]

	webHookUrl, isPresent := os.LookupEnv(SlackWebHookEnvVarName)

	if !isPresent {
		panic(fmt.Sprintf("Missing %s env, cannot report to Slack", SlackWebHookEnvVarName))
	}

	for {
		slackStartupMessage := fmt.Sprintf(
			"%s %s",
			args[0],
			strings.Join(args[1:], " "),
		)
		sendSlackMessage(
			webHookUrl,
			SlackMessage{Text: fmt.Sprintf("*Running Command: %s*", slackStartupMessage)},
		)
		result, err := watchProcess(args[0], args[1:])
		if err != nil {
			log.Printf("Failed to watch process: %v\n", err)
		} else {
			sendReportToSlack(
				webHookUrl,
				result,
				RetryStartApplicationAfter,
			)
		}

		log.Println("Waiting 3 seconds before restarting process")
		time.Sleep(RetryStartApplicationAfter)
	}
}

func watchProcess(cmd string, args []string) (*ProcessExecutionReport, error) {
	outputBuf := ring.New(TailedOutputBufferSize)
	totalLinesOutputted := 0
	startTime := time.Now()

	process := NewProcess(".", cmd, args)

	err := process.Start()
	if err != nil {
		return nil, err
	}

	for output := range process.Output {
		totalLinesOutputted++
		outputBuf = outputBuf.Next()
		outputBuf.Value = output.Message

		// Print to stdout for context
		fmt.Println(output.Message)
	}

	runDuration := time.Since(startTime)

	report := createReport(
		runDuration,
		process.ExitReason,
		totalLinesOutputted,
		outputBuf,
	)

	return report, nil
}

func createReport(
	runDuration time.Duration,
	exitReason *ExitReason,
	totalLinesOutputted int,
	outputBuf *ring.Ring,
) *ProcessExecutionReport {
	// Collect the last couple of lines for the report
	outputTailLength := minInt(totalLinesOutputted, outputBuf.Len())
	var tailedOutput [TailedOutputBufferSize]string
	for i := 0; i < outputTailLength; i++ {
		prevLine := outputBuf.Value

		if prevLine == nil {
			break
		}

		tailedOutput[i] = prevLine.(string)
		outputBuf = outputBuf.Prev()
	}

	return &ProcessExecutionReport{
		Uptime:              runDuration,
		ExitReason:          exitReason,
		TotalLinesOutputted: totalLinesOutputted,
		TailedOutput:        tailedOutput,
	}
}

func sendReportToSlack(slackEndpoint string, results *ProcessExecutionReport, willRetryAfter time.Duration) {
	var buffer bytes2.Buffer

	if results.ExitReason != nil {
		buffer.WriteString(fmt.Sprintf("Process terminated with exit code *%d* after *%s*\n", results.ExitReason.ExitCode, results.Uptime))
	} else {
		buffer.WriteString(fmt.Sprintf("Process terminated with unknown exit code *%s*\n", results.Uptime))
	}

	if results.TotalLinesOutputted != 0 {
		buffer.WriteString("_Application output before termination:_\n")
		buffer.WriteString("```")
		if results.TotalLinesOutputted > TailedOutputBufferSize {
			buffer.WriteString(fmt.Sprintf("... +%d Lines Truncated ...\n", results.TotalLinesOutputted-TailedOutputBufferSize))
		}
		for _, v := range results.TailedOutput {
			if len(v) > 0 {
				buffer.WriteString(v)
				buffer.WriteRune('\n')
			}
		}
		buffer.WriteString("```\n")
	} else {
		buffer.WriteString("Application terminated without any output.\n")
	}

	buffer.WriteString(fmt.Sprintf("I will try to restart the service in *%.0fs*\n", willRetryAfter.Seconds()))

	msg := SlackMessage{
		Text: buffer.String(),
	}

	sendSlackMessage(slackEndpoint, msg)
}

func sendSlackMessage(slackEndpoint string, msg SlackMessage) {
	bytes, _ := json.Marshal(msg)

	resp, _ := http.Post(slackEndpoint, "application/json", bytes2.NewBuffer(bytes))
	defer resp.Body.Close()

	r, _ := ioutil.ReadAll(resp.Body)
	log.Printf("Slack response: %s\n", string(r))
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

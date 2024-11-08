package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"sync"

	"github.com/gen2brain/beeep"
)

type AirbrakeResp struct {
	Count int
	Groups []AirbrakeGroup
}

type AirbrakeGroup struct {
	Id string
	ProjectId int
	Errors []AirbrakeError
	Context AirbrakeContext
	NoticeTotalCount int
}

type AirbrakeError struct {
	Type string
	Message string
}

type AirbrakeContext struct {
	Environment string
	Severity string
}

func startNotifier(c *Config) {
	c.withAirbrakeLogo(func(logo *os.File) {
		airbrakeRespChan := make(chan AirbrakeGroup)
		errChan := make(chan PollError)
		var wg sync.WaitGroup

		for _, project := range c.projects {
			poller := newAirbrakePoller(project, c.apiToken)
			wg.Add(1)
			go poller.startPolling(airbrakeRespChan, errChan, &wg)
		}

		go func() {
			wg.Wait()
			close(airbrakeRespChan)
			close(errChan)
		}()

		for {
			select {
			case data := <-airbrakeRespChan:
				go c.notify(data, logo)
			case pollError, ok := <-errChan:
				if ok {
					c.logger.Errorf("Poller stopped, project: %s, %s", pollError.ProjectId, pollError.Error)
				} else {
					c.logger.Warnf("All pollers stopped. Exiting…")
					return
				}
			}
		}
	})
}

func getSeverityIcon(severity string) string {
	if severity == "notice" {
		return "⚠️"
	}
	return "🚨"
}

func (c *Config) notify(resp AirbrakeGroup, logo *os.File) {
	projectId := fmt.Sprintf("%d", resp.ProjectId)
	groupId := resp.Id
	severity := resp.Context.Severity
	titleIcon := getSeverityIcon(severity)
	environment := resp.Context.Environment

	title := fmt.Sprintf("%s Airbrake %s %s", titleIcon, severity, titleIcon)

	for _, respError := range resp.Errors {
		message := fmt.Sprintf(
			"%s: %s\nhttps://adigi.airbrake.io/projects/%s/groups/%s", environment, respError.Type, projectId, groupId,
		)

		beeep.Notify(title, message, logo.Name())
	}
}

//go:embed assets/airbrakeLogo.png
var airbrakeLogo []byte

func (c *Config) withAirbrakeLogo(callback func(tmpFile *os.File)) error {
	tmpFile, err := os.CreateTemp("", "airbrakeLogo*.png")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpFile.Name())

	// Write the embedded logo content to the temporary file
	if _, err := tmpFile.Write(airbrakeLogo); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	callback(tmpFile)
	return nil
}

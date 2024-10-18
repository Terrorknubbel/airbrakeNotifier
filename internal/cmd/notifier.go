package cmd

import (
	_ "embed"
	"fmt"
	"os"
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

		go func() {
			for resp := range airbrakeRespChan {
				projectId := fmt.Sprintf("%d", resp.ProjectId)
				groupId := resp.Id
				severity := resp.Context.Severity
				environment := resp.Context.Environment

				var titleIcon string

				for _, respError := range resp.Errors {
					switch severity {
					case "notice":
						titleIcon = "⚠️"
					default:
						titleIcon = "🚨"
					}

						beeep.Notify(
							titleIcon + " Airbrake " + severity + " " + titleIcon,
							environment + ": " + respError.Type + "\nhttps://adigi.airbrake.io/projects/" + projectId + "/groups/" + groupId,
							logo.Name(),
						)
				}
			}
		}()

		for _, project := range c.projects {
			go poll(project, c.apiToken, airbrakeRespChan)
		}

		select {}
	})
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

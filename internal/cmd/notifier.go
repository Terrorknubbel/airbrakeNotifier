package cmd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

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
				errorType := resp.Errors[0].Type

				var titleIcon string
				switch severity {
				case "notice":
					titleIcon = "⚠️"
				default:
					titleIcon = "🚨"
				}

					beeep.Notify(
						titleIcon + " Airbrake " + severity + " " + titleIcon,
						environment + ": " + errorType + "\nhttps://adigi.airbrake.io/projects/" + projectId + "/groups/" + groupId,
						logo.Name(),
					)
			}
		}()

		c.pollAirbrake(airbrakeRespChan)
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

func (c *Config) pollAirbrake(airbrakeRespChan chan AirbrakeGroup) {
	ticker := time.NewTicker(time.Second * 1).C

	knownErrors := make(map[string]int)

	go func() {
		for range ticker {
			airbrakeUrl, err := url.Parse("https://api.airbrake.io/api/v4/projects/283571/groups?resolved=false")
			if err != nil {
				panic(err.Error())
			}

			query := airbrakeUrl.Query()
			query.Set("severity", c.projects[0].Severity)
			query.Set("key", c.airbrakeToken)

			airbrakeUrl.RawQuery = query.Encode()

			resp := getJson(airbrakeUrl.String())

			for _, group := range resp.Groups {
				if knownErrors[group.Id] == 0 || knownErrors[group.Id] < group.NoticeTotalCount {
					knownErrors[group.Id] = group.NoticeTotalCount
					airbrakeRespChan <- group
				}
			}
		}
	}()

	select {}
}

func getJson(url string) AirbrakeResp {
	r, err := http.Get(url)
	if err != nil {
			panic(err.Error())
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err.Error())
	}

	var resp AirbrakeResp
	json.Unmarshal(body, &resp)

	return resp
}

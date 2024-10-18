package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

func poll(project Project, token string, airbrakeRespChan chan AirbrakeGroup) {
	knownErrors := make(map[string]int)

	for range time.Tick(time.Second * time.Duration(project.PollingInterval)) {
		airbrakeUrl, err := url.Parse("https://api.airbrake.io/api/v4/projects/" + project.ProjectId + "/groups?resolved=false")
		if err != nil {
			panic(err.Error())
		}

		query := airbrakeUrl.Query()
		query.Set("severity", project.Severity)
		query.Set("key", token)

		airbrakeUrl.RawQuery = query.Encode()

		resp := getJson(airbrakeUrl.String())
		for _, group := range resp.Groups {
			if knownErrors[group.Id] == 0 || knownErrors[group.Id] < group.NoticeTotalCount {
				knownErrors[group.Id] = group.NoticeTotalCount
				airbrakeRespChan <- group
			}
		}
	}
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

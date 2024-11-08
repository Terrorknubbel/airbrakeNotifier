package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type airbrakePoller struct {
	client    *http.Client
	project   Project
	token     string
	baseURL   string
	knownErrs map[string]int
}

type PollError struct {
	ProjectId string
	Error string
}

func newAirbrakePoller(project Project, token string) *airbrakePoller {
	return &airbrakePoller{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		project:   project,
		token:     token,
		baseURL:   "https://api.airbrake.io/api/v4",
		knownErrs: make(map[string]int),
	}
}

func (ap *airbrakePoller) startPolling(resultChan chan<- AirbrakeGroup, errChan chan<- PollError, wg *sync.WaitGroup) {
	ticker := time.NewTicker(time.Duration(ap.project.PollingInterval) * time.Second)
	defer ticker.Stop()
	defer wg.Done()

	for range ticker.C {
		if err := ap.pollOnce(resultChan); err != nil {
			errChan <- PollError{ap.project.ProjectId, err.Error()}
			return
		}
	}
}

func (ap *airbrakePoller) pollOnce(resultChan chan<- AirbrakeGroup) error {
	url, err := ap.buildURL()

	if err != nil {
		return fmt.Errorf("failed to build URL: %w", err)
	}

	resp, err := ap.fetchAirbrakeData(url)
	if err != nil {
		return err
	}

	ap.processGroups(resp.Groups, resultChan)
	return nil
}

func (ap *airbrakePoller) buildURL() (string, error) {
	endpoint := fmt.Sprintf("%s/projects/%s/groups", ap.baseURL, ap.project.ProjectId)

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	query := u.Query()
	query.Set("resolved", strconv.FormatBool(ap.project.Resolved))
	query.Set("severity", ap.project.Severity)
	query.Set("key", ap.token)
	u.RawQuery = query.Encode()

	return u.String(), nil
}

func (ap *airbrakePoller) fetchAirbrakeData(url string) (*AirbrakeResp, error) {
	resp, err := ap.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result AirbrakeResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (ap *airbrakePoller) processGroups(groups []AirbrakeGroup, resultChan chan<- AirbrakeGroup) {
	for _, group := range groups {
		if ap.knownErrs[group.Id] == 0 || ap.knownErrs[group.Id] < group.NoticeTotalCount {
			ap.knownErrs[group.Id] = group.NoticeTotalCount
			resultChan <- group
		}
	}
}

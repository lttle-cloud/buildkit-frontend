package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/charmbracelet/log"
)

type ReportRequest struct {
	BuildId             string
	SerializedPlan      string
	SerializedBuildInfo string
}

func (request *ReportRequest) Encode() (string, error) {
	json, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	return string(json), nil
}

type ReportClient struct {
	baseUrl   string
	authToken string
	waitGroup sync.WaitGroup
}

func NewReportClient(baseUrl string, authToken string) *ReportClient {
	return &ReportClient{
		baseUrl:   baseUrl,
		authToken: authToken,
		waitGroup: sync.WaitGroup{},
	}
}

func (client *ReportClient) IsConfigured() bool {
	return client.baseUrl != "" && client.authToken != ""
}

func (client *ReportClient) Report(request *ReportRequest) error {
	url := fmt.Sprintf("%s/reports", client.baseUrl)
	json, err := request.Encode()
	if err != nil {
		return err
	}

	body := bytes.NewBuffer([]byte(json))

	httpRequest, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.authToken))
	response, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to report build: %s %s %s", response.Status, url, json)
	}

	return nil
}

func (client *ReportClient) ReportAsync(request *ReportRequest) error {
	client.waitGroup.Add(1)
	go func() {
		defer client.waitGroup.Done()

		err := client.Report(request)
		if err != nil {
			log.Errorf("failed to report build: %s", err)
		}
	}()
	return nil
}

func (client *ReportClient) WaitForAllAsyncReports() {
	client.waitGroup.Wait()
}

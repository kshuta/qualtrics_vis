package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	// req, err :=
}

const baseUrl = "https://%s.qualtrics.com/API/v3/surveys/%s/export-responses/"

var logger = log.New(os.Stderr, "logger: ", log.LstdFlags|log.Lshortfile)

func exportSurvey(apiToken, surveyId, dataCenter, fileFormat string) error {
	baseUrl := fmt.Sprintf(baseUrl, dataCenter, surveyId)
	reqBody := []byte(`{"format": "csv"}`)

	req, err := http.NewRequest(http.MethodPost, baseUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("x-api-token", apiToken)

	client := http.Client{}

	res, err := client.Do(req)

	progResp, err := parseRequestId(res.Body)
	defer res.Body.Close()
	if err != nil {
		return err
	}

	progressId := progResp.Result.ProgressId
	progressStatus := progResp.Result.Status

	for progressStatus != "complete" && progressStatus != "failed" {

	}

	// progressId := res.Body.

	return nil
}

type progressResponse struct {
	Result struct {
		ProgressId string `json:"progressId"`
		Status     string `json:"status"`
	} `json:"result"`
	Meta struct {
		RequestId  string `json:"requestId"`
		HttpStatus string `json:"httpStatus"`
	} `json:"meta"`
}

func parseRequestId(body io.Reader) (progressResponse, error) {
	var progResp progressResponse
	decoder := json.NewDecoder(body)
	err := decoder.Decode(&progResp)
	return progResp, err
}

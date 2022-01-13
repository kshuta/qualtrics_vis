package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	apiToken := os.Getenv("API_TOKEN")
	surveyId := os.Getenv("SURVEY_ID")
	dataCenter := os.Getenv("DATA_CENTER_ID")
	fileFormat := os.Getenv("FILE_FORMAT")

	fmt.Println(exportSurvey(apiToken, surveyId, dataCenter, fileFormat))
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
	if err != nil {
		return err
	}

	progResp, err := parseRequestId(res.Body)
	if err != nil {
		return err
	}

	res.Body.Close()

	progressId := progResp.Result.ProgressId
	progressStatus := progResp.Result.Status

	for progressStatus != "complete" && progressStatus != "failed" {
		time.Sleep(time.Second * 5)
		logger.Println("Status: ", progressStatus)
		requestCheckUrl := baseUrl + progressId
		req, err = http.NewRequest(http.MethodGet, requestCheckUrl, nil)
		if err != nil {
			panic(err)
			// return err
		}
		req.Header.Add("content-type", "application/json")
		req.Header.Add("x-api-token", apiToken)

		res, err = client.Do(req)
		if err != nil {
			// return err
			panic(err)
		}

		progResp, err = parseRequestId(res.Body)
		if err != nil {
			// return err
			panic(err)
		}

		res.Body.Close()

		progressStatus = progResp.Result.Status
	}

	if progressStatus == "failed" {
		return errors.New("failed to request survey response export")
	}

	fileId := progResp.Result.FileId

	requestDownloadUrl := baseUrl + fileId + "/file"
	req, err = http.NewRequest(http.MethodGet, requestDownloadUrl, nil)
	if err != nil {
		// return err
		panic(err)
	}

	res, err = client.Do(req)
	if err != nil {
		// return err
		panic(err)
	}

	return nil
}

type progressResponse struct {
	Result struct {
		ProgressId string `json:"progressId"`
		Status     string `json:"status"`
		FileId     string `json:"fileId"`
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

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const baseUrl = "https://%s.qualtrics.com/API/v3/surveys/%s/export-responses"

var logger = log.New(os.Stderr, "logger: ", log.Lshortfile)

var continuationToken string

func exportSurvey(apiToken, surveyId, dataCenter, fileFormat string) error {
	// mostly taken from the python code available on Qualtrics dev docs: https://api.qualtrics.com/guides/ZG9jOjg3NzY3Nw-new-survey-response-export-guide
	// initial request to create zip file
	headers := map[string]string{"content-type": "application/json", "x-api-token": apiToken}
	body := map[string]interface{}{"format": "csv", "useLabels": "true", "allowContinuation": "true"}
	if continuationToken != "" {
		body["continuationToken"] = continuationToken
	}
	url := parseUrl(baseUrl, dataCenter, surveyId)
	req, err := getInitialRequest(url, http.MethodPost, headers, body)
	if err != nil {
		return err
	}

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusBadRequest {
		return errors.New("bad request")
	}

	progResp, err := parseRequestId(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()

	progressId := progResp.Result.ProgressId
	progressStatus := progResp.Result.Status
	logger.Println("ProgressId: ", progressId)

	for progressStatus != "complete" && progressStatus != "failed" {
		time.Sleep(time.Second * 5)
		if progressStatus == "" {
			// error
			logger.Printf("Progress status empty, http status: %d\n", res.StatusCode)
			body, _ := io.ReadAll(res.Body)
			logger.Printf("%q", string(body))
			logger.Printf("%q", req.URL)
			return fmt.Errorf("status response returned %d", res.StatusCode)
		}
		logger.Println("Status: ", progressStatus)
		requestCheckUrl := parseUrl(baseUrl, dataCenter, surveyId, progressId)
		req, err := getInitialRequest(requestCheckUrl, http.MethodGet, headers, nil)
		if err != nil {
			return err
		}

		res, err = client.Do(req)
		if err != nil {
			return err
		}

		progResp, err = parseRequestId(res.Body)
		if err != nil {
			return err
		}
		res.Body.Close()

		progressStatus = progResp.Result.Status
		continuationToken = progResp.Result.ContinuationToken
	}

	if progressStatus == "failed" {
		return errors.New("failed to request survey response export")
	}

	fileId := progResp.Result.FileId

	requestDownloadUrl := parseUrl(baseUrl, dataCenter, surveyId, fileId, "/file")
	req, err = getInitialRequest(requestDownloadUrl, http.MethodGet, headers, nil)
	if err != nil {
		return err
	}

	res, err = client.Do(req)
	if err != nil {
		return err
	}

	readContent(res.Body)
	res.Body.Close()

	return nil
}

func readContent(body io.ReadCloser) {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		panic(err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(bodyBytes), int64(len(bodyBytes)))
	if err != nil {
		panic(err)
	}
	file, err := os.Create("data.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for _, zipFile := range zipReader.File {
		fmt.Println("Reading file:", zipFile.Name)
		unzippedFileBytes, err := readZipFile(zipFile)
		if err != nil {
			log.Println(err)
			continue
		}

		file.Write(unzippedFileBytes)
	}
}

func getInitialRequest(url, method string, headers map[string]string, body map[string]interface{}) (*http.Request, error) {
	// url := fmt.Sprintf(baseUrl, dataCenter, surveyId)
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	for key, val := range headers {
		req.Header.Add(key, val)
	}

	return req, nil
}

func parseUrl(baseUrl, dataCenter, surveyId string, paths ...string) string {
	url := fmt.Sprintf(baseUrl, dataCenter, surveyId)
	for _, path := range paths {
		url += "/" + path
	}

	logger.Println(url)

	return url
}

func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}

	defer f.Close()
	return io.ReadAll(f)
}

type progressResponse struct {
	Result struct {
		ProgressId        string `json:"progressId"`
		Status            string `json:"status"`
		FileId            string `json:"fileId"`
		ContinuationToken string `json:"continuationToken"`
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

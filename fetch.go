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
	"os/exec"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	// apiToken := os.Getenv("API_TOKEN")
	// surveyId := os.Getenv("SURVEY_ID")
	// dataCenter := os.Getenv("DATA_CENTER_ID")
	// fileFormat := os.Getenv("FILE_FORMAT")

	// if err := exportSurvey(apiToken, surveyId, dataCenter, fileFormat); err != nil {
	// 	logger.Fatalln(err)
	// }

	// run python script
	cmd := exec.Command("python3", "process_data.py", "data.csv")
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Fatal(err)
	}

	logger.Println(string(out))
}

const baseUrl = "https://%s.qualtrics.com/API/v3/surveys/%s/export-responses"

var logger = log.New(os.Stderr, "logger: ", log.LstdFlags|log.Lshortfile)

func exportSurvey(apiToken, surveyId, dataCenter, fileFormat string) error {
	// initial request to create zip file
	headers := map[string]string{"content-type": "application/json", "x-api-token": apiToken}
	body := map[string]string{"format": "csv", "useLabels": "true"}
	url := parseUrl(dataCenter, surveyId)
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
			logger.Printf("Progress status empty, http status: %d\n", res.StatusCode)
			body, _ := io.ReadAll(res.Body)
			logger.Printf("%q", string(body))
			logger.Printf("%q", req.URL)
			return fmt.Errorf("status response returned %d", res.StatusCode)
		}
		logger.Println("Status: ", progressStatus)
		requestCheckUrl := parseUrl(dataCenter, surveyId, progressId)
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
	}

	if progressStatus == "failed" {
		return errors.New("failed to request survey response export")
	}

	fileId := progResp.Result.FileId

	requestDownloadUrl := parseUrl(dataCenter, surveyId, fileId, "/file")
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

func getInitialRequest(url, method string, headers map[string]string, body map[string]string) (*http.Request, error) {
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

func parseUrl(dataCenter, surveyId string, paths ...string) string {
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

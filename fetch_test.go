package main

import (
	"os/exec"
	"testing"
)

func TestProcessing(t *testing.T) {
	t.Run("Processing the data", func(t *testing.T) {
		if err := exec.Command("python3", "process_data.py", "test_data.csv", "testdb.db").Run(); err != nil {
			t.Error(err)
		}
	})

	t.Run("processing the data twice.", func(t *testing.T) {
		if err := exec.Command("python3", "process_data.py", "test_data.csv", "testdb.db").Run(); err != nil {
			t.Error(err)
		}

		de := setupDB("sqlite3", "./testdb.db")
		defer de()

		// second round
		if err := exec.Command("python3", "process_data.py", "test_data.csv", "testdb.db").Run(); err != nil {
			t.Error(err)
		}

		post_record, err := getCompleteRecord("peer", "8", "2019")
		if err != nil {
			t.Fatal(err)
		}

		if len(post_record.Counts) > 1 {
			t.Error("Created different topic count")
		}

	})
}

// var responseValues = `{
// 	"values": {
// 	  "distributionChannel": "string",
// 	  "duration": 0,
// 	  "endDate": "2019-08-24T14:15:22Z",
// 	  "finished": 0,
// 	  "progress": 0,
// 	  "startDate": "2019-08-24T14:15:22Z",
// 	  "userLanguage": "string",
// 	  "QID65": 4,
// 	  "QID33": 48,
// 	  "QID35": ["1"],
// 	  "QID11_4": 1,
// 	  "QID11_5": 1,
// 	  "QID11_13": 1,
// 	  "QID11_14": 1,
// 	  "QID11_15": 1,
// 	  "QID11_3": 1
// 	}
//   }`

// func setSurveyResponse(t *testing.T, dataCenter, surveyId, apiToken string) error {
// 	url := parseUrl(baseUrl, dataCenter, surveyId)
// 	var values map[string]interface{}
// 	json.Unmarshal([]byte(responseValues), &values)
// 	headers := map[string]string{"content-type": "application/json", "x-api-token": apiToken}
// 	body := map[string]interface{}{"values": values}
// 	req, err := getInitialRequest(url, http.MethodPost, headers, body)
// 	if err != nil {
// 		return err
// 	}
// 	client := http.Client{}

// 	res, err := client.Do(req)

// 	if err != nil || res.StatusCode != http.StatusAccepted {
// 		return err
// 	}

// 	return nil

// }

// func setupEnv() (apiToken, surveyId, dataCenter, fileFormat string) {
// 	godotenv.Load()
// 	apiToken = os.Getenv("API_TOKEN")
// 	surveyId = os.Getenv("TEST_SURVEY_ID")
// 	dataCenter = os.Getenv("DATA_CENTER_ID")
// 	fileFormat = os.Getenv("FILE_FORMAT")
// 	return
// }

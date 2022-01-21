package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func main() {
	godotenv.Load()

	// apiToken := os.Getenv("API_TOKEN")
	// surveyId := os.Getenv("SURVEY_ID")
	// dataCenter := os.Getenv("DATA_CENTER_ID")
	// fileFormat := os.Getenv("FILE_FORMAT")

	// if err := exportSurvey(apiToken, surveyId, dataCenter, fileFormat); err != nil {
	// 	logger.Fatalln(err)
	// }

	records, err := getCompleteRecords("dec", "2021")
	if err != nil {
		log.Fatal(err)
	}

	df := setupDB()
	defer df()

}

func setupDB() func() {
	var err error
	db, err = sql.Open("sqlite3", "./sqlite.db")
	if err != nil {
		log.Fatal(err)
	}

	return func() { db.Close() }
}

func getCompleteRecords(month, year string) ([]Record, error) {
	records, err := getRecord("dec", "2021")
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		err = getTopicCounts(&record)
		if err != nil {
			return nil, err
		}
	}

	return records, nil
}

type Record struct {
	id          int
	department  string
	month       string
	year        string
	topicCounts TopicCounts
}

type TopicCounts map[string]int

func getRecord(month, year string) ([]Record, error) {
	stmt := fmt.Sprintf("select * from records where month=%q and year=%q", month, year)
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}

	records := make([]Record, 0)

	for rows.Next() {
		var record Record
		record.topicCounts = make(TopicCounts)
		err = rows.Scan(&record.id, &record.department, &record.month, &record.year)
		if err != nil {
			log.Println(err)
		} else {
			records = append(records, record)
		}
	}

	return records, rows.Close()
}

func getTopicCounts(record *Record) error {
	log.Println("getting topic counts")
	stmt := fmt.Sprintf("select topic, count from topic_counts where record_id=%d", record.id)
	rows, err := db.Query(stmt)
	if err != nil {
		return err
	}

	for rows.Next() {
		var topic string
		var count int
		err = rows.Scan(&topic, &count)
		if err != nil {
			rows.Close()
			return err
		}

		record.topicCounts[topic] = count
	}

	return rows.Close()

}

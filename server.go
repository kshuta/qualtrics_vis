package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
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

	db, err := sql.Open("sqlite3", "./sqlite.db")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	sqlStmt := `
	select * from records;
	`

	res := db.QueryRow(sqlStmt)

	var id int
	var dep string
	var month string
	var year string

	err = res.Scan(&id, &dep, &month, &year)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(dep, " ", month, " ", year)

}

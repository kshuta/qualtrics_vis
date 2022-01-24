package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"text/template"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Server struct {
	mux  *http.ServeMux
	once sync.Once
}

func (a *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.once.Do(func() {
		a.mux = http.NewServeMux()
		a.mux.HandleFunc("/home", indexHandlerFunc)
		a.mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		})
	})

	logger.Println("Request: ", r)

	a.mux.ServeHTTP(w, r)
}

func main() {
	godotenv.Load()

	// apiToken := os.Getenv("API_TOKEN")
	// surveyId := os.Getenv("SURVEY_ID")
	// dataCenter := os.Getenv("DATA_CENTER_ID")
	// fileFormat := os.Getenv("FILE_FORMAT")

	// if err := exportSurvey(apiToken, surveyId, dataCenter, fileFormat); err != nil {
	// 	logger.Fatalln(err)
	// }
	df := setupDB()
	defer df()

	server := Server{}

	logger.Fatal(http.ListenAndServe(":4000", &server))

}

func indexHandlerFunc(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	graphType := r.Form.Get("graph-type")
	advisingType := r.Form.Get("advising-type")
	period := r.Form.Get("m-y")

	if graphType == "" {
		graphType = "pie"
	}
	if advisingType == "" {
		advisingType = "academic"
	}

	var year, month string
	if period == "" {
		year = "2021"
		month = "12"
	} else {
		tmp := strings.Split(period, "-")
		fmt.Println(tmp)
		year, month = tmp[0], tmp[1]
	}

	record, err := getCompleteRecord(advisingType, month, year)
	if err != nil {
		logger.Println(err)
	}

	// t, err := template.ParseFiles("templates/index.html", "templates/pie.html")
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		logger.Fatal(err)
	}
	switch graphType {
	case "bar":
		t, err = t.Parse(barChart)
	case "radar":
		t, err = t.Parse(radarChart)
	default:
		t, err = t.Parse(pieChart)
	}
	if err != nil {
		logger.Fatal(err)
	}

	t.ExecuteTemplate(w, "index.html", record)

}

var pieChart = `
{{ define "script" }}
const myChart = new Chart(ctx, {
    type: 'pie',
    plugins: [ChartDataLabels],
    data: {
        labels: {{ .TopicsString }},
        datasets: [{
			label: {{ .Month }} + "/" + {{ .Year }},
            data: {{ .CountsString }},
            backgroundColor: [
                'rgba(255, 99, 132, 0.2)',
                'rgba(54, 162, 235, 0.2)',
                'rgba(255, 206, 86, 0.2)',
                'rgba(75, 192, 192, 0.2)',
                'rgba(153, 102, 255, 0.2)',
                'rgba(255, 159, 64, 0.2)'
            ],
            borderColor: [
                'rgba(255, 99, 132, 1)',
                'rgba(54, 162, 235, 1)',
                'rgba(255, 206, 86, 1)',
                'rgba(75, 192, 192, 1)',
                'rgba(153, 102, 255, 1)',
                'rgba(255, 159, 64, 1)'
            ],
            borderWidth: 1
        }]
    },
    options: {
        responsive: true,
        plugins: {
            legend: {
                position: 'right',
            },
            datalabels: {
                font: {
                    size: 20,
                    weight: 'bold'
                },
                formatter: function(value, context) {
                    if (context.dataIndex <= 3) {
                        return context.chart.data.labels[context.dataIndex] + "\n" + Math.ceil(value/{{ .TotalCount }}*100) + "%";
                    } else {
                        return "";
                    }
                },
                textAlign: 'center',

            },
        },
        responsive: true,
        maintainAspectRatio: false,
    }
    });
{{ end }}
`

var barChart = `
{{ define "script" }}
const myChart = new Chart(ctx, {
    type: 'bar',
    plugins: [ChartDataLabels],
    data: {
        labels: {{ .TopicsString }},
        datasets: [{
			label: {{ .Month }} + "/" + {{ .Year }},
            data: {{ .CountsString }},
            backgroundColor: [
                'rgba(255, 99, 132, 0.2)',
            ],
            borderColor: [
                'rgba(255, 99, 132, 1)',
            ],
            borderWidth: 1
        }
		]
    },
    options: {
        responsive: true,
        plugins: {
            legend: {
                position: 'right',
            },
            datalabels: {
                font: {
                    size: 20,
                    weight: 'bold'
                },
                textAlign: 'center',

            },
        },
        responsive: true,
        maintainAspectRatio: false,
    }
    });
{{ end }}
`

var radarChart = `
	{{ define "script" }}
	const myChart = new Chart(ctx, {
		type: 'radar',
		plugins: [ChartDataLabels],
		data: {
			labels: {{ .TopicsString }},
			datasets: [{
				label: {{ .Month }} + "/" + {{ .Year }},
				data: {{ .CountsString }},
				backgroundColor: [
					'rgba(255, 99, 132, 0.2)',
					'rgba(54, 162, 235, 0.2)',
					'rgba(255, 206, 86, 0.2)',
					'rgba(75, 192, 192, 0.2)',
					'rgba(153, 102, 255, 0.2)',
					'rgba(255, 159, 64, 0.2)'
				],
				borderColor: [
					'rgba(255, 99, 132, 1)',
					'rgba(54, 162, 235, 1)',
					'rgba(255, 206, 86, 1)',
					'rgba(75, 192, 192, 1)',
					'rgba(153, 102, 255, 1)',
					'rgba(255, 159, 64, 1)'
				],
				borderWidth: 1
			}]
		},
		options: {
			responsive: true,
			plugins: {
				legend: {
					position: 'right',
				},
				datalabels: {
					font: {
						size: 20,
						weight: 'bold'
					},
					textAlign: 'center',
	
				},
			},
			responsive: true,
			maintainAspectRatio: false,
		}
		});
	{{end}}
`

func setupDB() func() {
	var err error
	db, err = sql.Open("sqlite3", "./sqlite.db")
	if err != nil {
		logger.Fatal(err)
	}

	return func() { db.Close() }
}

func getCompleteRecord(department, month, year string) (Record, error) {
	record, err := getRecord(department, month, year)
	if err != nil {
		record.Empty = true
		return record, err
	}

	err = getTopicCounts(&record)
	if err != nil {
		return record, err
	}

	getTotalCount(&record)
	getStringFields(&record)

	return record, nil
}

// convert map to this format
// data: [{x:'Sales', y:20}, {x:'Revenue', y:10}]

type Record struct {
	Id           int
	Department   string
	Month        string
	Year         string
	TopicCounts  TopicCounts
	Topics       []string
	TopicsString string
	Counts       []int
	CountsString string
	TotalCount   int
	Empty        bool
}

type TopicCounts map[string]int

func getStringFields(record *Record) {
	b, _ := json.Marshal(record.Topics)
	record.TopicsString = string(b)
	b1, _ := json.Marshal(record.Counts)
	record.CountsString = string(b1)
}

func getTotalCount(record *Record) {
	for _, count := range record.Counts {
		record.TotalCount += count
	}

	if record.TotalCount == 0 {
		record.Empty = true
	} else {
	}
}

func getRecord(department, month, year string) (Record, error) {
	var record Record
	stmt := fmt.Sprintf("select * from records where department=%q and month=%q and year=%q", department, month, year)
	row := db.QueryRow(stmt)

	record.TopicCounts = make(TopicCounts)
	err := row.Scan(&record.Id, &record.Department, &record.Month, &record.Year)
	if err != nil {
		return record, err
	}

	return record, nil
}

func getTopicCounts(record *Record) error {
	stmt := fmt.Sprintf("select topic, count from topic_counts where record_id=%d order by count DESC", record.Id)
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

		record.Topics = append(record.Topics, topic)
		record.Counts = append(record.Counts, count)
		record.TopicCounts[topic] = count
	}

	return rows.Close()

}

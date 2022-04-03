package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB
var nextMonth time.Month
var nextYear int

type Server struct {
	mux  *http.ServeMux
	once sync.Once
}

func (a *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.once.Do(func() {
		a.mux = http.NewServeMux()
		a.mux.HandleFunc("/", indexHandlerFunc)
		a.mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		})
	})

	logger.Println("Request: ", r)
	a.mux.ServeHTTP(w, r)
}

var fetch = flag.Bool("fetch", true, "flag to determine if we are fetching data from the API or not. Mainly for development")

func main() {
	flag.Parse()
	godotenv.Load()

	// fetch data from the api if it's the beginning of a new month.
	year, month, _ := time.Now().Date()
	if *fetch && (month == 0 || (month >= nextMonth || year >= nextYear)) {
		// fetch data
		if err := fetchData(); err != nil {
			logger.Println(err)
			logger.Println("Failed to fetch data from api")
		}
		nextPeriod := time.Now().AddDate(0, 1, 0)
		nextYear = nextPeriod.Year()
		nextMonth = nextPeriod.Month()

		// record to database
		out, err := exec.Command("python", "./process_data.py", "data.csv").Output()
		if err != nil {
			logger.Fatal(err)
		}
		fmt.Println(string(out))
	}

	df := setupDB("postgres", "null")
	defer df()

	server := Server{}

	port := fmt.Sprintf(":%s", os.Getenv("PORT"))

	logger.Printf("Serving at port %s\n", port)
	logger.Fatal(http.ListenAndServe(port, &server))

}

func fetchData() error {
	godotenv.Load()

	apiToken := os.Getenv("API_TOKEN")
	surveyId := os.Getenv("SURVEY_ID")
	dataCenter := os.Getenv("DATA_CENTER_ID")
	fileFormat := os.Getenv("FILE_FORMAT")

	err := exportSurvey(apiToken, surveyId, dataCenter, fileFormat)

	return err
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

	time.Now()

	var year, month string
	if period == "" {
		t := time.Now()
		year = strconv.Itoa(t.Year())
		month = strconv.Itoa(int(t.Month()))
	} else {
		tmp := strings.Split(period, "-")
		year, month = tmp[0], tmp[1]
	}

	record, err := getCompleteRecord(advisingType, month, year)
	if err != nil {
		logger.Println(err)
	}
	prevMonth, prevYear, err := getPrevMonth(month, year)
	if err != nil {
		logger.Println(err)
	}

	prevRecord, err := getPrevRecord(advisingType, prevMonth, prevYear, record)
	if err != nil {
		logger.Println(err)
	}

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

	getStringFields(&record)
	getStringFields(&prevRecord)

	// logger.Println(record)
	logger.Println(prevRecord)
	t.ExecuteTemplate(w, "index.html", []Record{record, prevRecord})

}

var pieChart = `
{{ define "script" }}

{{ $record := index . 0 }}
{{ $prevRecord := index . 1}}

document.getElementsByName("chart-selection")[0].setAttribute("selected", "selected")
const myChart = new Chart(ctx, {
    type: 'pie',
    plugins: [ChartDataLabels],
    data: {
        labels: {{ $record.TopicsString }},
        datasets: [{
			label: {{ $record.Month }} + "/" + {{ $record.Year }},
            data: {{ $record.CountsString }},
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
                        return context.chart.data.labels[context.dataIndex] + "\n" + Math.ceil(value/{{ $record.TotalCount }}*100) + "%";
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

{{ $record := index . 0 }}
{{ $prevRecord := index . 1}}

document.getElementsByName("chart-selection")[1].setAttribute("selected", "selected")
const myChart = new Chart(ctx, {
    type: 'bar',
    plugins: [ChartDataLabels],
    data: {
        labels: {{ $record.TopicsString }},
        datasets: [
		{
			label: {{ $record.Month }} + "/" + {{ $record.Year }},
            data: {{ $record.CountsString }},
            backgroundColor: [
                'rgba(255, 99, 132, 0.2)',
            ],
            borderColor: [
                'rgba(255, 99, 132, 1)',
            ],
            borderWidth: 1
        },
		{{ if not $prevRecord.Empty }}
		{
			label: {{ $prevRecord.Month }} + "/" + {{ $prevRecord.Year }},
			data: {{ $prevRecord.CountsString }},
			backgroundColor: [
				'rgba(54, 162, 235, 1)',
			],
			borderColor: [
				'rgba(54, 162, 235, 1)',
			]
		}
		{{ end }}
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

	{{ $record := index . 0 }}
	{{ $prevRecord := index . 1}}

	document.getElementsByName("chart-selection")[2].setAttribute("selected", "selected")
	const myChart = new Chart(ctx, {
		type: 'radar',
		plugins: [ChartDataLabels],
		data: {
			labels: {{ $record.TopicsString }},
			datasets: [{
				label: {{ $record.Month }} + "/" + {{ $record.Year }},
				data: {{ $record.CountsString }},
				backgroundColor: [
					'rgba(255, 99, 132, 0.2)',
				],
				borderColor: [
					'rgba(255, 99, 132, 1)',
				],
				borderWidth: 1
			},
			{{ if not $prevRecord.Empty }}
			{
				label: {{ $prevRecord.Month }} + "/" + {{ $prevRecord.Year }},
				data: {{ $prevRecord.CountsString }},
				backgroundColor: [
					'rgba(54, 162, 235, 1)',
				],
				borderColor: [
					'rgba(54, 162, 235, 1)',
				]
			}
			{{ end }}
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
					formatter: function(value, context) {
						if (context.dataIndex <= 4) {
							return value;
						} else {
							return "";
						}
					},
				},
			},
			responsive: true,
			maintainAspectRatio: false,
		}
		});
	{{end}}
`

func getPrevMonth(month, year string) (string, string, error) {
	ipm, err := strconv.Atoi(month)
	if err != nil {
		return "", "", err
	}
	if ipm-1 >= 1 {
		if ipm-1 < 10 {
			return "0" + strconv.Itoa(ipm-1), year, nil
		} else {
			return strconv.Itoa(ipm - 1), year, nil
		}
	}

	ipm = 12
	ipy, err := strconv.Atoi(year)
	if err != nil {
		return "", "", err
	}

	return strconv.Itoa(ipm), strconv.Itoa(ipy - 1), nil
}

func setupDB(dbname, dbfile string) func() {
	var err error
	const dsnUrlFormat = "postgres://%s:%s@%s:%s/%s?sslmode=require"
	DNS := fmt.Sprintf(dsnUrlFormat, os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DBNAME"))

	db, err = sql.Open(dbname, DNS)
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

	return record, nil
}

func getPrevRecord(department, month, year string, target Record) (Record, error) {
	record, err := getRecord(department, month, year)
	if err != nil {
		record.Empty = true
		return record, err
	}

	err = getTopicCounts(&record)
	if err != nil {
		return record, err
	}

	record.Counts = make([]int, len(target.Counts))

	for idx, topic := range target.Topics {
		val, ok := record.TopicCounts[topic]
		if ok {
			record.Counts[idx] = val
		} else {
			record.Counts[idx] = 0
		}
	}

	getTotalCount(&record)

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

	if len(record.Month) == 1 {
		record.Month = "0" + record.Month
	}
}

func getTotalCount(record *Record) {
	for _, count := range record.Counts {
		record.TotalCount += count
	}

	if record.TotalCount == 0 {
		record.Empty = true
	}
}

func getRecord(department, month, year string) (Record, error) {
	var record Record

	i_month, err := strconv.Atoi(month)
	if err != nil {
		return record, err
	}
	i_year, err := strconv.Atoi(year)
	if err != nil {
		return record, err
	}

	logger.Println(i_month)
	logger.Println(i_year)

	row := db.QueryRow("select * from records where department=$1 and month=$2 and year=$3", department, i_month, i_year)

	record.TopicCounts = make(TopicCounts)
	err = row.Scan(&record.Id, &record.Department, &record.Month, &record.Year)
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

package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/MihaiBlebea/go-logger/repos"
	_ "github.com/go-sql-driver/mysql"
)

func main() {

	repos.Name
	os.Setenv("MYSQL_USER", "admin")
	os.Setenv("MYSQL_PASSWORD", "root")
	os.Setenv("MYSQL_DATABASE", "dev_logger")
	os.Setenv("MYSQL_HOST", "127.0.0.1:32783")

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/log", logHandler)
	http.HandleFunc("/logs", logsHandler)
	http.HandleFunc("/delete", deleteHandler)

	err := http.ListenAndServe(":8089", nil)
	if err != nil {
		log.Panic(err)
	}
}

// rootHandler handles the connections for the "/" path
func rootHandler(w http.ResponseWriter, r *http.Request) {

	createLogTable()

	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte("All good"))
}

// logHandler handles the connections for the POST "/log" path
func logHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Panic("Only POST method is supported")
	}

	decoder := json.NewDecoder(r.Body)
	var logJSON LogJSON
	err := decoder.Decode(&logJSON)
	if err != nil {
		log.Panic(err)
	}

	model := LogModel{
		MemberID:  logJSON.MemberID,
		Action:    logJSON.Action,
		EventCode: logJSON.EventCode}

	result := insertLog(model)

	lastID, err := result.LastInsertId()
	if err != nil {
		log.Panic(err)
	}
	model.ID = int(lastID)

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logJSON)
}

// logsHandler handles the connections for the GET "/logs" path
func logsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Panic("Only GET method is supported")
	}

	// code := r.URL.Query().Get("code")
	// member := r.URL.Query().Get("member")

	models := selectLogs()

	w.Header().Add("Content-Type", "application/json")

	jsonModels, err := json.Marshal(models)
	if err != nil {
		log.Print(err)
	}
	w.Write(jsonModels)
}

// deleteHandler handles the connections for the DELETE "/delete" path
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		log.Panic("Only DELETE method is supported")
	}

	stringID := r.URL.Query().Get("id")
	if stringID == "" {
		deleteAll()
	}
	id, err := strconv.Atoi(stringID)
	if err != nil {
		log.Panic(err)
	}

	deleteLog(id)
	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte("All good"))
}

// Persistence function, connect to the database
func databaseConnection() *sql.DB {
	user := os.Getenv("MYSQL_USER")
	password := os.Getenv("MYSQL_PASSWORD")
	host := os.Getenv("MYSQL_HOST")
	database := os.Getenv("MYSQL_DATABASE")

	db, err := sql.Open("mysql", user+":"+password+"@tcp("+host+")/"+database)
	if err != nil {
		log.Panic(err)
	}
	return db
}

// Persistence function, create the logs table
func createLogTable() sql.Result {
	db := databaseConnection()

	statement, err := db.Prepare(`CREATE TABLE IF NOT EXISTS logs(
		id INT AUTO_INCREMENT PRIMARY KEY,
		member_id INT,
		action TEXT,
		event_code VARCHAR(250),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)

	if err != nil {
		log.Panic(err)
	}

	result, err := statement.Exec()
	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

	return result
}

// Persistence function, insert a log into the database
func insertLog(model LogModel) sql.Result {
	db := databaseConnection()

	statement, err := db.Prepare("INSERT INTO logs (member_id, action, event_code) VALUES(?,?,?)")
	if err != nil {
		log.Panic(err)
	}

	result, err := statement.Exec(model.MemberID, model.Action, model.EventCode)
	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

	return result
}

// Persistence function, select all logs from the database
func selectLogs() []LogModel {
	db := databaseConnection()

	rows, err := db.Query("SELECT * FROM logs")
	if err != nil {
		log.Panic(err)
	}

	var (
		id        int
		memberID  int
		action    string
		eventCode string
		createAt  string
	)
	var models []LogModel
	for rows.Next() {
		err := rows.Scan(&id, &memberID, &action, &eventCode, &createAt)
		if err != nil {
			log.Print(err)
		}

		model := LogModel{
			id,
			memberID,
			action,
			eventCode,
			createAt,
		}

		models = append(models, model)
	}
	defer db.Close()

	return models
}

// Persistence function, delete a log from the database
func deleteLog(id int) sql.Result {
	db := databaseConnection()
	statement, err := db.Prepare("DELETE FROM logs WHERE id = ?")
	if err != nil {
		log.Print(err)
	}

	result, err := statement.Exec(id)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	return result
}

// Persistence function, delete all logs from the database
func deleteAll() sql.Result {
	db := databaseConnection()
	statement, err := db.Prepare("DELETE FROM logs")
	if err != nil {
		log.Print(err)
	}
	result, err := statement.Exec()
	if err != nil {
		log.Panic(err)
	}

	defer db.Close()
	return result
}

// LogJSON used as DTO from Json
type LogJSON struct {
	MemberID  int
	Action    string
	EventCode string
}

// LogModel is the model for Log
type LogModel struct {
	ID        int    `json:"id"`
	MemberID  int    `json:"memberID"`
	Action    string `json:"action"`
	EventCode string `json:"eventCode"`
	CreatedAt string `json:"createdAt"`
}

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

type Data struct {
	Name          string
	Author        string
	Difficulty    uint8
	SongID        uint8
	SongStartTime int
	FloorID       uint8
	BackgroundID  uint8
	StartingColor [3]uint8
	LevelData     []int
	PathData      []int
	CameraData    []int
}

var LID int64

func check(err error, where string) bool {
	if err != nil {
                log.Fatal(where + ": ", err)
		return false
        }
	return true
}

var database, dberr = sql.Open("sqlite3", "./levels.db")

func QLID() int64 {
	rows, err := database.Query("SELECT id FROM levels")
	check(err, "querying ids")
	err = nil
	var lid int64
	for rows.Next() {
		err = rows.Scan(&lid)
		check(err, "scanning id")
		err = nil
	}

	return lid
}

func hewo(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.Write([]byte("<h2>Hewo!</h2>"))
}

func getLevel(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/json")
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}
	r.ParseForm()
	if r.Form["id"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	_, err := strconv.Atoi(r.Form["id"][0])
	if recover() != nil || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	var data string
	row := database.QueryRow("SELECT data FROM levels WHERE id = " + r.Form["id"][0])
	row.Scan(&data)
	w.Write([]byte(data))
}

func postLevel(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}
	r.ParseForm()
	if r.Form["data"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Shit gets checked for valid JSON
	if !json.Valid([]byte(r.Form["data"][0])) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Shit gets checked for a valid schema
	var UD Data // Unmarshalled Data
	err := json.Unmarshal([]byte(r.Form["data"][0]), &UD)
	if recover() != nil || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Shit gets checked for acutal valid data
	if (len(UD.Name) > 24 || len(UD.Author) > 24 || UD.SongID > 21 || UD.Difficulty > 5 || UD.FloorID > 3 || UD.BackgroundID > 2 || len(UD.Name) == 0 || len(UD.Author) == 0 || UD.SongID < 0 || UD.Difficulty < 0 || UD.FloorID < 0 || UD.BackgroundID < 0) { 
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = nil
	_, err = database.Exec("INSERT INTO levels VALUES(?, ?)", LID+1, r.Form["data"][0])
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: levels.data" {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			log.Fatal("Inserting level: ", err)
			return
		}
	}
	// Kinda stable, if you get an error message, go report it as a bug... Please check the FAQ first. OFC check other issues and the schedule.md
	LID++
	w.Write([]byte(fmt.Sprint(LID)))
}

func getRecents(w http.ResponseWriter, r *http.Request) {
	rows, err := database.Query("SELECT * FROM levels ORDER BY id DESC LIMIT 20") // Specify the amount of recents you want to see
	check(err, "quering recent levels")
	err = nil

	var result string
	
	for rows.Next() {
		var id int64
		var data string
		var unmarshalleddata Data
		err = rows.Scan(&id, &data)
		if err != nil {
			continue
		}
		if !json.Valid([]byte(data)) {
			continue
		}
		err = json.Unmarshal([]byte(data), &unmarshalleddata)
		if err != nil {
			continue
		}
	
		result += fmt.Sprint(id) + "\n"
		result += unmarshalleddata.Name + "\n"
		result += unmarshalleddata.Author + "\n"
		result += fmt.Sprint(unmarshalleddata.Difficulty) + "\n"
	}
	w.Write([]byte(result))
}

func main() {
	//# Database using sqlite3
	check(dberr, "loading database")
	InitTable, err := database.Prepare("CREATE TABLE IF NOT EXISTS levels(id INTEGER PRIMARY KEY, data BLOB UNIQUE)")
	check(err, "creating levels' table")
	err = nil
	InitTable.Exec()
	LID = QLID()
	defer database.Close()

	//# Routing
	http.HandleFunc("/", hewo)
	http.HandleFunc("/level/get", getLevel)
	http.HandleFunc("/levels/recent", getRecents)
	http.HandleFunc("/level/publish", postLevel)

	//# Listen And Serve
	fmt.Println("Listening and serving...")
	//HTTPS is protection against man in the middle attacks, which will never happen, unless your in a public network AND someone is TARGETTING YOU
	//err = http.ListenAndServeTLS(":9991", "TLS.crt", "TLS.key", nil)
	err = http.ListenAndServe(":9991", nil)
	check(err, "listen n serving")
}

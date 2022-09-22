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

func check(err error, where string) bool {
	if err != nil {
    log.Fatal(where + ": ", err)
		return false
  }
	return true
}

var database, dberr = sql.Open("sqlite3", "./levels.db")

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
	row := database.QueryRow("SELECT data FROM levels WHERE ROWID = " + r.Form["id"][0])
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
	var Iid int64
	_, err = database.Exec("INSERT INTO levels VALUES(?)", r.Form["data"][0])
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: levels.data" {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			log.Fatal("Inserting level: ", err)
			return
		}
	}
	row := database.QueryRow("SELECT ROWID FROM levels ORDER BY DESC LIMIT 1")
	row.Scan(&Iid)

	w.Write([]byte(fmt.Sprint(Iid+1)))
}

func getRecents(w http.ResponseWriter, r *http.Request) {
	rows, err := database.Query("SELECT data,ROWID FROM levels ORDER BY ROWID DESC LIMIT 20") // Specify the amount of recents you want to see
	check(err, "quering recent levels")
	err = nil

	var result string
	
	for rows.Next() {
		var id int64
		var data string
		var unmarshalleddata Data
		err = rows.Scan(&data, &id)
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
	InitTable, err := database.Prepare("CREATE TABLE IF NOT EXISTS levels(data BLOB UNIQUE)")
	check(err, "creating levels' table")
	err = nil
	InitTable.Exec()
	defer database.Close()

	//# Routing
	http.HandleFunc("/", hewo)
	http.HandleFunc("/level/get", getLevel)
	http.HandleFunc("/levels/recent", getRecents)
	http.HandleFunc("/level/publish", postLevel)

	//# Listen And Serve
	fmt.Println("Listening and serving...")
	//HTTPS is protection against man in the middle attacks, which will never happen, unless your in a public network AND someone is TARGETING YOU
	//Although it doesn't work on unity's network thing sadly, and thats why it didn't work. If it ever does, please make an issue
	//err = http.ListenAndServeTLS(":9991", "TLS.crt", "TLS.key", nil)
	err = http.ListenAndServe(":9991", nil)
	check(err, "listen n serving")
}

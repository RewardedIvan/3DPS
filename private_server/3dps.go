package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

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

var EmptyData = Data{Name: "", Author: "", Difficulty: 0, SongID: 0, SongStartTime: 0, FloorID: 0, BackgroundID: 0, StartingColor: [3]uint8{0, 0, 0}, LevelData: []int{}, PathData: []int{}, CameraData: []int{}}
var LID int64

func check(err error, where string) bool {
	if err != nil {
                log.Fatal(where + ": ", err)
		return false
        }
	return true
}

func QLID() int64 {
	database, err := sql.Open("sqlite3", "./levels.db")
	check(err, "loading database")
	err = nil
	rows, err := database.Query("select id from levels")
	check(err, "querying ids")
	err = nil
	var lid int64
	for rows.Next() {
		err = rows.Scan(&lid)
		check(err, "scanning id")
		err = nil
	}

	database.Close()
	return lid
}

func QRDB(Query string) *sql.Row {
	database, err := sql.Open("sqlite3", "./levels.db")
	check(err, "loading database")
	row := database.QueryRow(Query)

	database.Close()
	return row
}

func QDB(Query string) *sql.Rows {
	database, err := sql.Open("sqlite3", "./levels.db")
	check(err, "loading database")
	err = nil
	rows, err := database.Query(Query)
	check(err, "quering \"" + Query + "\"")

	database.Close()
	return rows
}

func EDB(Exec string, V1 int64, V2 string) bool {
	database, err := sql.Open("sqlite3", "./levels.db")
	check(err, "loading database")
	err = nil
	_, err := database.Exec(Exec, V1, V2)

	if err != nil {
		if err.Error() == "UNIQUE constraint failed: levels.data" {
			return false
		} else {
			log.Fatal("Exec", err)
		}
	}

	database.Close()
	return true
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
	row := QRDB("SELECT data FROM levels WHERE id = " + r.Form["id"][0])
	row.Scan(&data)
	w.Write([]byte(data))
}

func ReverseLines(str string) string {
	// This is a mess lmao
	lines := strings.Split(str, "\n")
	var ret string

	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}

	for i := range lines {
		if i == len(lines)-1 {
			ret += lines[i]
		} else {
			ret += lines[i] + "\n"
		}
	}
	return ret
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
	// Shit gets checked on steroids
	var UD Data // Unmarshalled Data
	err := json.Unmarshal([]byte(r.Form["data"][0]), &UD)
	if recover() != nil || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !EDB("INSERT INTO levels VALUES (?, ?)", LID+1, r.Form["data"][0]) {
		w.WriteHeader(420) //SPAM
		return
	}
	// Kinda stable, if you get an error message, go report it as a bug... Please check the FAQ first. OFC check other issues and the schedule.md
	LID++
	w.Write([]byte(fmt.Sprint(LID)))
}

func getRecents(w http.ResponseWriter, r *http.Request) {
	rows := QDB("SELECT * FROM levels ORDER BY id DESC")

	var result string

	for i := 0; i < 20; i++ { // Specify the amount of recents you want to see
		if rows.Next() {
			var id int64
			var data string
			var unmarshalleddata Data
			err := rows.Scan(&id, &data)
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
		} else {
			break
		}
	}
	w.Write([]byte(result))
}

func main() {
	//# Database using sqlite3
	database, err := sql.Open("sqlite3", "./levels.db")
	check(err, "loading database")
	err = nil
	InitTable, err := database.Prepare("CREATE TABLE IF NOT EXISTS levels(id integer primary key autoincrement, data blob unique)")
	check(err, "creating levels' table")
	err = nil
	InitTable.Exec()
	database.Close()
	LID = QLID()

	//# Routing
	http.HandleFunc("/", hewo)
	http.HandleFunc("/level/get", getLevel)
	http.HandleFunc("/levels/recent", getRecents)
	http.HandleFunc("/level/publish", postLevel)

	//# Listen And Serve
	//HTTPS is protection against man in the middle attacks, which will never happen, unless your in a public network AND someone is TARGETTING YOU
	//err = http.ListenAndServeTLS(":9991", "TLS.crt", "TLS.key", nil)
	err = http.ListenAndServe(":9991", nil)
	fmt.Println("Now Serving...")
	check(err, "listen n serving")
}

// Whoever reads this, just know I am new to golang, some how made a personal record for the best project

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

func QLID() int64 {
	database, err := sql.Open("sqlite3", "./levels.db")
	if err != nil {
		log.Fatal("db: ", err)
	}
	rows, erro := database.Query("select id from levels")
	var lid int64
	for rows.Next() {
		error := rows.Scan(&lid)
		if error != nil {
			log.Fatal("Query lid: ", erro)
		}
	}

	if erro != nil {
		log.Fatal("Query", erro)
	}

	database.Close()
	return lid
}

func QRDB(Query string) *sql.Row {
	database, err := sql.Open("sqlite3", "./levels.db")
	if err != nil {
		log.Fatal("db: ", err)
	}
	row := database.QueryRow(Query)

	database.Close()
	return row
}

func QDB(Query string) *sql.Rows {
	database, err := sql.Open("sqlite3", "./levels.db")
	if err != nil {
		log.Fatal("db: ", err)
	}
	rows, erro := database.Query(Query)
	if erro != nil {
		log.Fatal("Query", erro)
	}

	database.Close()
	return rows
}

func EDB(Exec string, V1 int64, V2 string) bool {
	database, err := sql.Open("sqlite3", "./levels.db")
	if err != nil {
		log.Fatal("db: ", err)
	}
	_, erro := database.Exec(Exec, V1, V2)

	if erro != nil {
		if erro.Error() == "UNIQUE constraint failed: levels.data" {
			return false
		} else {
			log.Fatal("Exec", erro)
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

	if _, err := strconv.Atoi(r.Form["id"][0]); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var data string
	row := QRDB("SELECT data FROM levels WHERE id = " + r.Form["id"][0])
	row.Scan(&data)
	w.Write([]byte(data))
}

func ReverseLines(str string) string {
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
	// Kinda stable, if you get an error message, go report it as a bug... Please check the FAQ first
	LID++
	w.Write([]byte(fmt.Sprint(LID)))
}

func getRecents(w http.ResponseWriter, r *http.Request) {
	rows := QDB("SELECT * FROM levels ORDER BY id DESC")

	var result string

	for i := 0; i < 10; i++ { // Specify the amount of recents you want to see
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
			erro := json.Unmarshal([]byte(data), &unmarshalleddata)
			if erro != nil {
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
	database, erro := sql.Open("sqlite3", "./levels.db")
	InitTable, _ := database.Prepare("CREATE TABLE IF NOT EXISTS levels(id integer primary key autoincrement, data longtext unique)")
	InitTable.Exec()
	database.Close()
	LID = QLID()
	if erro != nil {
		log.Fatal("db: ", erro)
		panic(erro)
	}

	//# Routing
	http.HandleFunc("/", hewo)
	http.HandleFunc("/level/get", getLevel)
	http.HandleFunc("/levels/recent", getRecents)
	http.HandleFunc("/level/publish", postLevel)

	//# Listen And Serve
	//error := http.ListenAndServeTLS(":9991", "TLS.crt", "TLS.key", nil)
	//HTTPS is protection against man in the middle attacks, which will never happen, unless your in a public network AND someone is TARGETTING YOU
	fmt.Println("Now Serving...")
	error := http.ListenAndServe(":9991", nil)
	if error != nil {
		log.Fatal("ListenAndServeTLS: ", error)
	}
}

// Whoever reads this, just know I am new to golang, some how made a personal record for the best project

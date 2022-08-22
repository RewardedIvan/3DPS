package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

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

func EDB(Exec string, V1 int64, V2 string) int64 {
	database, err := sql.Open("sqlite3", "./levels.db")
	if err != nil {
		log.Fatal("db: ", err)
	}
	res, erro := database.Exec(Exec, V1, V2)
	affect, _ := res.RowsAffected()

	if erro != nil {
		log.Fatal("Exec", erro)
	}

	database.Close()
	return affect
}

func LID() int64 {
	database, err := sql.Open("sqlite3", "./levels.db")
	if err != nil {
		log.Fatal("db: ", err)
	}

	r, _ := database.Exec("")
	lid, _ := r.LastInsertId()

	database.Close()
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

	if _, err := strconv.Atoi(r.Form["id"][0]); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var data string
	row := QRDB("SELECT data FROM levels WHERE id = " + r.Form["id"][0])
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
	if !strings.HasPrefix(r.Form["data"][0], "{") && !strings.HasSuffix(r.Form["data"][0], "}") {
		//this only checks if its actually a json object, nothing else is checked... Sadly
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	EDB("INSERT INTO levels VALUES (?, ?)", LID()+int64(1), r.Form["data"][0])
	w.Write([]byte(strconv.FormatInt(LID()+1, 10)))
	// Still not stable, should improve this later... TODO
}

func getRecents(w http.ResponseWriter, r *http.Request) {
	rows := QDB("SELECT data FROM levels")

	//TODO

	//var result string
	var i int64 = 0
	for rows.Next() {
		if LID() == i || i == 10 {
			break
		}

		var data string
		rows.Scan(&data)
		fmt.Println(data)

		i++
	}

	//w.Write()
}

func main() {
	//# Database using sqlite3
	database, erro := sql.Open("sqlite3", "./levels.db")
	InitTable, _ := database.Prepare("CREATE TABLE IF NOT EXISTS levels(id integer primary key autoincrement, data longtext unique)")
	InitTable.Exec()
	database.Close()
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

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

const SV = "1.4"
const CV = "1.2.1"

type UploadData struct {
	Name       string `json:"name"`
	Author     string `json:"author"`
	Difficulty uint8  `json:"difficulty"`
	Data       string `json:"data"` // why is this a string ):
}

type Data struct {
	SongID        uint8    `json:"songId"`
	SongStartTime int      `json:"songStartTime"`
	FloorID       uint8    `json:"floorId"`
	BackgroundID  uint8    `json:"backgroundId"`
	StartingColor [3]uint8 `json:"startingColor"`
	LevelData     []int    `json:"levelData"`
	PathData      []int    `json:"pathData"`
	CameraData    []int    `json:"cameraData"`
}

// global vars
var database *sql.DB

func check(err error, where string, exit bool) bool {
	if err != nil {
		log.Fatal(where+": ", err)
		if exit {
			os.Exit(1)
		}
	}
	return err != nil
}

func Error(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprint(w, error) // this used to be Fprint*ln*
}

func hewo(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.Write([]byte("<h2>Hewo!</h2>"))
}

func VersionMWFunc(next func(http.ResponseWriter, *http.Request)) http.Handler {
	return VersionMW(http.HandlerFunc(next))
}
func AuthMWFunc(next func(http.ResponseWriter, *http.Request)) http.Handler {
	return AuthMW(http.HandlerFunc(next))
}

func VersionMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version := r.Header.Get("Version")

		if requireVersion && version != "1.2.1" { // fuck higher versions i can't be bothered
			Error(w, "Error: Version unsupported. Please\nupdate to version 1.2.1 or higher.", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func AuthMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requireAuth {
			auth := r.Header.Get("Authorization")

			if auth == "" {
				if usefulErrors {
					Error(w, "Error: no token", http.StatusBadRequest)
				} else {
					Error(w, "Error: invalid token", http.StatusBadRequest)
				}
				return
			}

			row := database.QueryRow("SELECT ip FROM tokens WHERE token = ?", auth)

			var ip string
			err := row.Scan(&ip)
			//fmt.Println("authmw ip: " + ip + ", remoteaddr: " + r.RemoteAddr + " [" + getEverythingBeforeSubstr(r.RemoteAddr, ":") + "]")

			if err != nil {
				if usefulErrors {
					Error(w, "Error: token doesn't exist", http.StatusBadRequest)
				} else {
					Error(w, "Error: invalid token", http.StatusBadRequest)
				}
				return
			}

			//fmt.Printf("should check ip on tokens: %t; current ip: %s; sqlite ip: %s; ip != empty: %t; sqlite ip != ip: %t\n", !ipIndependentTokens, getEverythingBeforeSubstr(r.RemoteAddr, ":"), ip, ip != "", ip != getEverythingAfterSubstr(r.RemoteAddr, ":"))

			if !ipIndependentTokens {
				if ip != "" && ip != getEverythingBeforeSubstr(r.RemoteAddr, ":") {
					if usefulErrors {
						Error(w, "Error: valid token used from different ip", http.StatusBadRequest)
					} else {
						Error(w, "Error: invalid token", http.StatusBadRequest)
					}
					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func downloadLevel(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/json")

	id, err := strconv.Atoi(getEverythingAfterSubstr(r.URL.Path, "/"))
	if recover() != nil || err != nil {
		if usefulErrors {
			Error(w, "Error: id isn't a number", http.StatusBadRequest)
		} else {
			Error(w, "Error: id not in int format", http.StatusBadRequest)
		}
		return
	}

	var data string
	row := database.QueryRow("SELECT data FROM levels WHERE ROWID = ?", id)
	err = row.Scan(&data)
	if err != nil {
		Error(w, "Level not found", http.StatusBadRequest)
		return
	}
	w.Write([]byte(data))
}

func postLevel(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		if usefulErrors {
			Error(w, "content type isn't json", http.StatusUnsupportedMediaType)
		} else {
			w.WriteHeader(http.StatusUnsupportedMediaType)
		}
		return
	}

	jsonbuf, err := io.ReadAll(r.Body)
	check(err, "reading upload body", false)

	//fmt.Println(string(jsonbuf))
	//d, _ := json.Marshal(UploadData{})
	//fmt.Println(string(d))

	if !json.Valid(jsonbuf) {
		if usefulErrors {
			Error(w, "body doesn't contain json", http.StatusBadRequest)
		} else {
			Error(w, "Level name or author name not allowed.\nPlease revise and try again in 5 minutes.", http.StatusBadRequest)
		}
		return
	}

	// Shit gets checked for a valid schema
	var UD UploadData // Unmarshalled/Upload Data
	err = json.Unmarshal(jsonbuf, &UD)
	if recover() != nil || err != nil {
		if usefulErrors {
			Error(w, "body's json isn't in the right scheme", http.StatusBadRequest)
		} else {
			Error(w, "Level name or author name not allowed.\nPlease revise and try again in 5 minutes.", http.StatusBadRequest)
		}
		return
	}

	var LUD Data // "Level" in Upload Data
	err = json.Unmarshal([]byte(UD.Data), &LUD)
	if recover() != nil || err != nil {
		if usefulErrors {
			Error(w, "\"data\" in body's json isn't in the right scheme", http.StatusBadRequest)
		} else {
			Error(w, "Level name or author name not allowed.\nPlease revise and try again in 5 minutes.", http.StatusBadRequest)
		}
		return
	}

	// Shit gets checked for acutal valid data
	if len(UD.Name) > 24 || len(UD.Author) > 24 || LUD.SongID > 21 || UD.Difficulty > 5 || LUD.FloorID > 3 || LUD.BackgroundID > 2 || len(UD.Name) == 0 || len(UD.Author) == 0 || LUD.SongID < 0 || UD.Difficulty < 0 || LUD.FloorID < 0 || LUD.BackgroundID < 0 {
		if usefulErrors {
			Error(w, "level's name or author or songid or difficulty or floorid or backgroundid is invalid", http.StatusBadRequest)
		} else {
			Error(w, "Level name or author name not allowed.\nPlease revise and try again in 5 minutes.", http.StatusBadRequest)
		}
		return
	}

	//remarshal for extra safety :D
	var marshaled []byte
	marshaled, err = json.Marshal(UD)
	check(err, "remarshaling upload data", false)

	var Iid int64
	_, err = database.Exec("INSERT INTO levels VALUES(?)", marshaled)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: levels.data" {
			if usefulErrors {
				Error(w, "spam/level with the same data exists", http.StatusBadRequest)
			} else {
				// :troll:
				Error(w, "Level name or author name not allowed.\nPlease revise and try again in 5 minutes.", http.StatusBadRequest)
			}
			return
		} else {
			log.Fatal("Inserting level: ", err)
			return
		}
	}
	row := database.QueryRow("SELECT ROWID FROM levels ORDER BY ROWID DESC LIMIT 1")
	row.Scan(&Iid)

	w.Write([]byte(fmt.Sprint(Iid)))
}

func getRecents(w http.ResponseWriter, r *http.Request) {
	rows, err := database.Query("SELECT data,ROWID FROM levels ORDER BY ROWID DESC LIMIT ?", recentLevels)
	check(err, "quering recent levels", false)

	var result string

	for rows.Next() {
		var id int64
		var data string
		var unmarshalleddata UploadData
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

func getEverythingAfterSubstr(input string, substr string) string {
	lastIndex := strings.LastIndex(input, substr)
	if lastIndex == -1 {
		return input
	}
	return input[lastIndex+1:]
}
func getEverythingBeforeSubstr(input string, substr string) string {
	firstIndex := strings.Index(input, substr)
	if firstIndex == -1 {
		return input
	}
	return input[:firstIndex]
}

func authenticate(w http.ResponseWriter, r *http.Request) {
	if requireBean {
		buf, err := io.ReadAll(r.Body)
		check(err, "reading body buffer", false)

		if !bytes.Equal(buf, []byte("bean")) {
			if usefulErrors {
				Error(w, "use \"bean\" as a body", http.StatusBadRequest)
			} else {
				// lmao this doesn't exist in the original server
				Error(w, "Error: modified client behavior", http.StatusBadRequest)
			}
			return
		}
	}

	//fmt.Println("auth ip: " + r.RemoteAddr + " [" + getEverythingBeforeSubstr(r.RemoteAddr, ":") + "]")

	token := uuid.New()
	var err error
	if ipIndependentTokens {
		_, err = database.Exec("INSERT INTO tokens VALUES(?,?,?)", token.String(), time.Now().Unix(), "")
	} else {
		_, err = database.Exec("INSERT INTO tokens VALUES(?,?,?)", token.String(), time.Now().Unix(), getEverythingBeforeSubstr(r.RemoteAddr, ":"))
	}
	check(err, "inserting a token", false)

	w.Write([]byte(token.String()))
}

// Flags
var recentLevels int
var ipIndependentTokens bool
var requireBean bool
var requireAuth bool
var requireVersion bool
var usefulErrors bool

func main() {
	/// Flags
	DBStr := flag.String("db", "./3DPS.db", "Database connection string/file")
	printVer := flag.Bool("print-versions", true, "Print the client and server version on startup")
	flag.IntVar(&recentLevels, "recent-levels", 30, "The amount of recent levels")
	address := flag.String("address", ":30924", "The address to listen on")
	flag.BoolVar(&requireAuth, "require-auth", true, "Require a valid token")
	flag.BoolVar(&requireVersion, "require-version", true, "Require the \"Version\" header to be set to \"1.2.1\"")
	flag.BoolVar(&ipIndependentTokens, "ip-independent-tokens", false, "Whether to store and check ips on tokens")
	flag.BoolVar(&requireBean, "require-bean", false, "When logging in, require a body of \"bean\"")
	flag.BoolVar(&usefulErrors, "useful-errors", false, "Provide more info about what went wrong instead of \"revise the fucking level name\" (may break auth system)")

	flag.Parse()

	/// Database using sqlite3
	var err error
	database, err = sql.Open("sqlite3", *DBStr)
	check(err, "loading database", true)

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS levels(data BLOB UNIQUE)")
	check(err, "creating levels table", true)

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS tokens(token UUID UNIQUE, timestamp INTEGER, ip string)")
	check(err, "creating tokens table", true)

	defer database.Close()

	/// Routing
	http.HandleFunc("/", hewo)
	http.Handle("/login", VersionMWFunc(authenticate))
	http.Handle("/download/", VersionMW(AuthMWFunc(downloadLevel)))
	http.Handle("/recent", VersionMW(AuthMWFunc(getRecents)))
	http.Handle("/upload", VersionMW(AuthMWFunc(postLevel)))

	/// Listen And Serve
	if *printVer {
		fmt.Printf("Server %s, Client %s; Listening on port %s\n", SV, CV, getEverythingAfterSubstr(*address, ":"))
	}
	//HTTPS is protection against man in the middle attacks, which will never happen, unless your in a public network AND someone is TARGETING YOU
	//Although https doesn't work on unity's network thing (curl) sadly. (maybe im just dumb)
	//err = http.ListenAndServeTLS(address, "TLS.crt", "TLS.key", nil)
	err = http.ListenAndServe(*address, nil)
	check(err, "listen n serving", true)
}

package main;

import (
	"net/url"
	"net/http"
	"os"
	"fmt"
	"io"
	"strconv"
	"flag"

	"strings"
)


func check(err error, where string, exit bool) {
	if err != nil {
		fmt.Printf("\033[31m%s: %s\033[0m\n", where, err.Error())
		if exit == true {
			os.Exit(1);
		}
	}
}

func frmt(what string, res bool) {
	if res == true {
		fmt.Printf("%s; return \033[32m%t\033[0m\n", what, res)
	} else {
		fmt.Printf("%s; return \033[31m%t\033[0m\n", what, res)
	}
}

var hostname *string
var port *int

func main() {
	hostname = flag.String("hostname", "localhost", "The hostname that the server is hosted on")
	port = flag.Int("port", 9991, "The port that the server is runnning on")

	flag.Parse()

	// This script assumes that the database is empty
	c := http.Client{};

	// Comments help you figure out what its testing
	frmt("postLevel(c, 'Impossible title and description')", !postLevel(c, "./lvl.json")) // valid data
	frmt("postLevel(c, 'Same but with a little change')", !postLevel(c, "./lvl2.json")) // spam detection
	frmt("postLevel(c, 'Valid Level')", postLevel(c, "./lvl3.json")) // valid data
	frmt("getRecents(c)", getRecents(c)) // if the levels have been created and if it sends valid data
}

func getRecents(c http.Client) bool {
	res, err := http.Get(fmt.Sprintf("http://%s:%d/levels/recent", *hostname, *port));

	check(err, "GETting recent levels", false);
	if err != nil { return false }
	if res == nil { return false }

	bdy, err := io.ReadAll(res.Body);
	check(err, "reading the response body", false);
	
	if err != nil { return false }
	if bdy == nil { fmt.Println("\033[31mfailed: body is nil\033[0m"); return false }

	// Parse the response
	// 'Sample' by 'RewardedIvan' (5 difficulty) #69
	//.....

	vals := strings.Split(string(bdy), "\n")

	for i := 0; i < len(vals); i++ { // Remove empty elements
		if (vals[i] == "") {
			vals = append(vals[:i], vals[i+1:]...)
		}
	}

	if len(vals) % 4 != 0 {
		fmt.Println("\033[31mfailed: wrong amount of values\033[0m"); // This generally never happens
		return false
	}

	// This is litterly the reverseengineered code, but in go
	for i := 0; i < len(vals); i += 4 {
		if _, err := strconv.Atoi(vals[i + 3]); err != nil {
			fmt.Println("\033[31mfailed: difficulty isn't a number\033[0m");
			return false
		}

		if _, err := strconv.Atoi(vals[i]); err != nil {
			fmt.Println("\033[31mfailed: the ID isn't a number\033[0m");
			return false
		}

		fmt.Printf("'%s' by '%s' (%s difficulty) #%s\n", vals[i + 1], vals[i + 2], vals[i + 3], vals[i]);
	}

	// /*(used for debugging)*/ for i := 0; i < len(vals); i++ { fmt.Printf(";%s", vals[i]); } ; fmt.Println("")

	return true;
}

func postLevel(c http.Client, fstr string) bool {
	fp, err := os.Open(fstr);
	check(err, "opening the level file", false);
	if err != nil { return false }
	fpi, err := fp.Stat();
	check(err, "stating the level file", false);
	if err != nil { return false }
	lvl := make([]byte, fpi.Size());
	_, err = fp.Read(lvl);
	check(err, "reading the level file", false);
	if err != nil { return false }

	form := url.Values{};

	form.Add("data", string(lvl));

	bdy := strings.NewReader(form.Encode());
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/level/publish", *hostname, *port), bdy);
	
	check(err, "making request", false);
	if err != nil { return false }
	
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded");

	res, err := c.Do(req);
	
	check(err, "couldn't do request to post a level", false);
	if err != nil { return false }
	if res.StatusCode != 200 { fmt.Println("\033[31mfailed to publish: status code != 200\033[0m"); return false; }

	bb, err := io.ReadAll(res.Body);
	
	check(err, "reading the body", false);
	if err != nil { return false }
	if bb == nil { fmt.Println("\033[31mfailed to publish: body is nil\033[0m"); return false; }

	fmt.Printf("ID: %s\n", string(bb));

	return true
}

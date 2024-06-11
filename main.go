package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db   *sql.DB
	dbMu sync.Mutex
)

type User struct {
	ID    int
	Name  string
	Month int
	Day   int
}

type PageData struct {
	Users []User
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		month := r.FormValue("month")
		day := r.FormValue("day")

		dbMu.Lock()
		_, err := db.Exec("INSERT INTO birthdays (name, month, day) VALUES(?,?,?)", name, month, day)
		if err != nil {
			http.Error(w, "Error inserting into database", http.StatusInternalServerError)
			return
		}
		dbMu.Unlock()

		http.Redirect(w, r, "/", http.StatusFound)
		return
	} else {
		var users []User

		rows, _ := db.Query("SELECT * FROM birthdays")
		defer rows.Close()

		for rows.Next() {
			var user User
			rows.Scan(&user.ID, &user.Name, &user.Month, &user.Day)
			users = append(users, user)
		}

		data := PageData{
			Users: users,
		}

		err := tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func init() {
	var err error
	db, err = sql.Open("sqlite3", "birthdays.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", indexPage)
	http.ListenAndServe(":8080", nil)
}

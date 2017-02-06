package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Book struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Description string `json:"description,omitempty"`
	ImgUrl      string `json:"imageUrl,omitempty"`
	Notes       string `json:"notes,omitempty"`
	YearWritten string `json:"yearWritten,omitempty"`
}

var port = "4001"
var db *sql.DB
var err error

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok!"))
}

func sayHelloName(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println(r.Form)
	fmt.Printf("path: ", r.URL.Path)
	fmt.Print("scheme: ", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key: ", k)
		fmt.Println("value: ", strings.Join(v, ""))
	}
	fmt.Fprintf(w, "Hey there!")
}

func login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method: ", r.Method)
	if r.Method == "GET" {
		t, _ := template.ParseFiles("login.gtpl")
		t.Execute(w, "ok!")
		return
	} else {
		r.ParseForm()
		fmt.Println("username: ", r.Form["username"])
		fmt.Println("password: ", r.Form["password"])
	}

}

func GetBooks(w http.ResponseWriter, r *http.Request) {

	var book Book
	var books []Book
	rows, err := db.Query("SELECT * FROM books")
	if err != nil {
		log.Println("Error Getting Rows", err)
	}

	defer rows.Close()

	for rows.Next() {
		var ID string
		var Title string
		var Author string
		var Description sql.NullString
		var ImageUrl sql.NullString
		var Notes sql.NullString
		var YearWritten sql.NullString

		if err := rows.Scan(&ID, &Title, &Author, &Description, &ImageUrl, &Notes, &YearWritten); err != nil {
			log.Println("Error in Rows: ", err)
		}
		book.ID = ID
		book.Title = Title
		book.Author = Author
		book.Description = Description.String
		book.ImgUrl = ImageUrl.String
		book.Notes = Notes.String
		book.YearWritten = YearWritten.String

		books = append(books, book)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(books)
}

func main() {

	db, err = sql.Open("mysql", "root:@/library")
	if err != nil {
		log.Println("Error connecting to db: ", err.Error())
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Println("Error on db ping: ", err.Error())
	}

	router := mux.NewRouter()

	//---------------------------
	// Health Check
	//---------------------------
	router.HandleFunc("/health", healthCheck)

	//---------------------------
	// Main routes
	//---------------------------
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/", sayHelloName)
	api.HandleFunc("/login", login)
	api.HandleFunc("/books", GetBooks).Methods("GET")

	server := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:" + port,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Server started on port: " + port)
	log.Fatal(server.ListenAndServe()) // pass the router as the 2nd argument to ListenAndServe
}

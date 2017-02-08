package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type Book struct {
	ID          string  `json:"id" db:"id"`
	Title       string  `json:"title" db:"title"`
	Author      string  `json:"author" db:"author"`
	Description *string `json:"description,omitempty" db:"description,omitempty"`
	ImageUrl    *string `json:"imageUrl,omitempty" db:"imageUrl,omitempty"`
	Notes       *string `json:"notes,omitempty" db:"notes,omitempty"`
	YearWritten *string `json:"yearWritten,omitempty" db:"yearWritten,omitempty"`
	Read        bool    `json:"read" db:"read"`
}

var port = "4001"
var db *sqlx.DB
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

func GetBookByTitle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	title := vars["title"]

	var resp []Book
	resp, err = GetBooksBy("title", title)
	if err != nil {
		log.Println("Error querying added book", err)
	}
	json.NewEncoder(w).Encode(resp)
}

func GetBooksByAuthor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]

	var resp []Book
	resp, err = GetBooksBy("author", author)
	if err != nil {
		log.Println("Error querying added book", err)
	}

	json.NewEncoder(w).Encode(resp)
}

func GetBooksBy(filter string, id interface{}) ([]Book, error) {

	var result []Book
	var buffer bytes.Buffer
	buffer.WriteString("SELECT * FROM books WHERE ")
	buffer.WriteString(filter)
	buffer.WriteString(" = ?")
	err = db.Select(&result, buffer.String(), id)

	if err != nil {
		log.Println("Error querying db: ", err)
		return nil, err
	}

	return result, nil
}

func GetAllBooks(w http.ResponseWriter, r *http.Request) {

	var books []Book
	err = db.Select(&books, "SELECT * FROM books")

	if err != nil {
		log.Println("Error Getting Rows", err)
	}

	json.NewEncoder(w).Encode(books)
}

func AddBook(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading body: ", err)
	}

	var book Book
	err = json.Unmarshal(body, &book)
	if err != nil {
		log.Println("Error Unmarshallin body: ", err)
	}

	stmt, err := db.Prepare("INSERT INTO `books`(`title`, `author`, `description`, `imageUrl`, `notes`, `yearWritten`, `read`) VALUES(?,?,?,?,?,?,?);")
	if err != nil {
		fmt.Println("Error preparing the query statement: ", err)
	}
	result, err := stmt.Exec(book.Title, book.Author, book.Description, book.ImageUrl, book.Notes, book.YearWritten, book.Read)
	if err != nil {
		log.Println("Error Creating Record", err)
	}

	insertedId, err := result.LastInsertId()
	if err != nil {
		log.Println("Error getting id of inserted book", err)
	}

	var resp []Book

	resp, err = GetBooksBy("id", insertedId)
	if err != nil {
		log.Println("Error querying added book", err)
	}

	json.NewEncoder(w).Encode(resp)
}

func main() {

	db, err = sqlx.Open("mysql", "root:@/library")
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
	api.HandleFunc("/books", GetAllBooks).Methods("GET")
	api.HandleFunc("/book/title/{title}", GetBookByTitle).Methods("GET")
	api.HandleFunc("/books/author/{author}", GetBooksByAuthor).Methods("GET")
	api.HandleFunc("/book", AddBook).Methods("POST")

	//---------------------------
	// Create the Server
	//---------------------------
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

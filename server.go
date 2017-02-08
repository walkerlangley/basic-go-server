package main

import (
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

//func GetBookByTitle(w http.ResponseWriter, r *http.Request) {
//vars := mux.Vars(r)
//title := vars["title"]

//stmt, err := db.Prepare("SELECT * FROM books WHERE title = ?")
//if err != nil {
//fmt.Println("Error preparing query statement: ", err)
//}
//var book Book
//var ID string
//var Title string
//var Author string
//var Description sql.NullString
//var ImageUrl sql.NullString
//var Notes sql.NullString
//var YearWritten sql.NullString
//var Read bool
//err = stmt.QueryRow(title).Scan(&ID, &Title, &Author, &Description, &ImageUrl, &Notes, &YearWritten, &Read)
//switch {
//case err == sql.ErrNoRows:
//log.Printf("No book with that title")
//case err != nil:
//log.Fatal(err)
//default:
//book.ID = ID
//book.Title = Title
//book.Author = Author
//book.Description = Description.String
//book.ImageUrl = ImageUrl.String
//book.Notes = Notes.String
//book.YearWritten = YearWritten.String
//book.Read = Read
//}
//json.NewEncoder(w).Encode(book)
//}

//func GetBooksByAuthor(w http.ResponseWriter, r *http.Request) {
//vars := mux.Vars(r)
//author := vars["author"]

//stmt, err := db.Prepare("SELECT * FROM books WHERE author = ?")
//if err != nil {
//fmt.Println("Error preparing query statement: ", err)
//}
//var book Book
//var books []Book

//rows, err := stmt.Query(author)
//if err != nil {
//log.Println("Error Getting Rows", err)
//}

//defer rows.Close()

//for rows.Next() {
//var ID string
//var Title string
//var Author string
//var Description sql.NullString
//var ImageUrl sql.NullString
//var Notes sql.NullString
//var YearWritten sql.NullString
//var Read bool

//if err := rows.Scan(&ID, &Title, &Author, &Description, &ImageUrl, &Notes, &YearWritten, &Read); err != nil {
//log.Println("Error in Rows: ", err)
//}

//book.ID = ID
//book.Title = Title
//book.Author = Author
//book.Description = Description.String
//book.ImageUrl = ImageUrl.String
//book.Notes = Notes.String
//book.YearWritten = YearWritten.String
//book.Read = Read
//books = append(books, book)
//}
//err = rows.Err()
//if err != nil {
//log.Fatal(err)
//}
//json.NewEncoder(w).Encode(book)
//}

func GetBooks(w http.ResponseWriter, r *http.Request) {

	var books []Book
	//stmt, err := db.Prepare("SELECT * FROM books")
	//if err != nil {
	//fmt.Println("Error preparing the query statement: ", err)
	//}
	err = db.Select(&books, "SELECT * FROM books")

	//rows, err := stmt.Queryx()
	if err != nil {
		log.Println("Error Getting Rows", err)
	}

	//defer rows.Close()

	//for rows.Next() {
	//// This is another way of scanning into the struct, but lose types since everything is an interface
	//col, _ := rows.Columns()
	//numCols := len(col)

	//rowStruct := make([]interface{}, numCols)
	//if err := rows.Scan(&rowStruct...); err != nil {
	//log.Println("Error in Rows: ", err)
	//}

	//var ID string
	//var Title string
	//var Author string
	//var Description sql.NullString // Could also use a pointer to a string instead of sql.NullString since pointers can be null
	//var ImageUrl sql.NullString
	//var Notes sql.NullString
	//var YearWritten sql.NullString
	//var Read bool

	//if err := rows.Scan(&ID, &Title, &Author, &Description, &ImageUrl, &Notes, &YearWritten, &Read); err != nil {
	//log.Println("Error in Rows: ", err)
	//}

	//book.ID = ID
	//book.Title = Title
	//book.Author = Author
	//book.Description = Description.String
	//book.ImageUrl = ImageUrl.String
	//book.Notes = Notes.String
	//book.YearWritten = YearWritten.String
	//book.Read = Read
	//books = append(books, book)
	//}
	//err = rows.Err()
	//if err != nil {
	//log.Fatal(err)
	//}

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

	log.Println("Query Result: ", result.LastInsertId)

	log.Println("BODY: ", book)
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
	api.HandleFunc("/books", GetBooks).Methods("GET")
	//api.HandleFunc("/book/title/{title}", GetBookByTitle).Methods("GET")
	//api.HandleFunc("/books/author/{author}", GetBooksByAuthor).Methods("GET")
	api.HandleFunc("/book", AddBook).Methods("POST")
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

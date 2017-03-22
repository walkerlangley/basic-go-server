package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type Book struct {
	ID          string  `json:"id" db:"id"`
	UserId      int     `json:"userId" db:"userId"`
	Title       string  `json:"title" db:"title"`
	Author      string  `json:"author" db:"author"`
	Description *string `json:"description,omitempty" db:"description,omitempty"`
	ImageUrl    *string `json:"imageUrl,omitempty" db:"imageUrl,omitempty"`
	Notes       *string `json:"notes,omitempty" db:"notes,omitempty"`
	YearWritten *string `json:"yearWritten,omitempty" db:"yearWritten,omitempty"`
	Read        bool    `json:"read" db:"read"`
}

type User struct {
	ID        string `json:"id" db:"id"`
	FirstName string `json:"firstName" db:"firstName"`
	LastName  string `json:"lastName" db:"lastName"`
	Username  string `json:"username" db:"username"`
	Password  string `json:"password" db:"password"`
}

type UserPayload struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type LoginPayload struct {
	Username string
	Password string
}

var port = "3000"
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

// Not using this.  Just keeping around so I can see how to parse forms
//func login(w http.ResponseWriter, r *http.Request) {
//fmt.Println("method: ", r.Method)
//if r.Method == "GET" {
//t, _ := template.ParseFiles("login.gtpl")
//t.Execute(w, "ok!")
//return
//} else {
//r.ParseForm()
//fmt.Println("username: ", r.Form["username"])
//fmt.Println("password: ", r.Form["password"])
//}

//}

func GetBookByTitle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	title := vars["title"]

	var book Book
	err = db.Get(&book, "SELECT * FROM books WHERE title = ?", title)
	if err != nil {
		fmt.Println("Error getting book: ", title, ' ', err)
		http.Error(w, "Error getting book", 400)
		return
	}
	json.NewEncoder(w).Encode(book)
}

func GetBooksByAuthor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]

	var books []Book

	err = db.Select(&books, "SELECT * FROM books WHERE author = ?", author)
	if err != nil {
		log.Println("Error Getting Rows", err)
		http.Error(w, "Error Getting Rows", 400)
		return
	}

	json.NewEncoder(w).Encode(books)
}

func GetBooksBy(filter string, id interface{}) ([]Book, error) {

	var result []Book

	log.Println("FILTER AND STRING", filter, " ", id)
	var buffer bytes.Buffer
	buffer.WriteString("SELECT * FROM books WHERE ")
	buffer.WriteString(filter)
	buffer.WriteString(" = ?")
	err = db.Select(&result, buffer.String(), id)
	if err != nil {
		log.Println("Error querying db: ", err)
		return nil, err
	}

	log.Println("RESULTS OF GET BY:", result)
	return result, nil
}

func GetBooks(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	vars := mux.Vars(r)
	userId := vars["userId"]
	var books []Book
	err = db.Select(&books, "SELECT * FROM books WHERE userId = ?", userId)

	if err != nil {
		log.Println("Error Getting Rows", err)
		http.Error(w, "Error Getting Rows", 400)
		return
	}

	json.NewEncoder(w).Encode(books)
}

func AddBook(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading body: ", err)
		http.Error(w, "Error reading body", 400)
		return
	}

	var book Book
	err = json.Unmarshal(body, &book)
	if err != nil {
		log.Println("Error Unmarshallin body: ", err)
		http.Error(w, "Error unmarshaling body", 400)
		return
	}

	stmt, err := db.Prepare("INSERT INTO `books`(`title`, `author`, `description`, `imageUrl`, `notes`, `yearWritten`, `read`) VALUES(?,?,?,?,?,?,?);")
	if err != nil {
		fmt.Println("Error preparing the query statement: ", err)
		http.Error(w, "Error preparing the query statement", 400)
		return
	}
	result, err := stmt.Exec(book.Title, book.Author, book.Description, book.ImageUrl, book.Notes, book.YearWritten, book.Read)
	if err != nil {
		log.Println("Error Creating Record", err)
		http.Error(w, "Error Creating Record", 400)
		return
	}

	insertedId, err := result.LastInsertId()
	if err != nil {
		log.Println("Error getting id of inserted book", err)
		http.Error(w, "Error getting id of inserted book", 400)
		return
	}

	var resp []Book

	resp, err = GetBooksBy("id", insertedId)
	if err != nil {
		log.Println("Error querying added book", err)
		http.Error(w, "Error querying added book", 400)
		return
	}

	json.NewEncoder(w).Encode(resp[0])
	//log.Println("Query Result: ", insertedId)

	//json.NewEncoder(w).Encode(book)
}

func createAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Sorry, looks like something broke.  Please try again", 404)
		return
	}

	var tmpUser UserPayload
	err = json.Unmarshal(body, &tmpUser)
	if err != nil {
		http.Error(w, "Sorry, it looks like something was wrong with one of the fields.  Please try again", 404)
		return
	}

	var existing User

	err = db.Get(&existing, "SELECT * FROM users WHERE username = ?", tmpUser.Username)
	if err == nil {
		http.Error(w, "Alread a user", 404)
		return
	}

	if err != sql.ErrNoRows {
		log.Println("Error something other than sql.ErrNoRows...", err)
		http.Error(w, "Some error....", 500)
		return
	}

	password := tmpUser.Password

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error generating hash: ", err)
		http.Error(w, "Unfortunate it looks like there was an error with your password.  Please try again", 404)
		return
	}

	var user User
	user.FirstName = tmpUser.FirstName
	user.LastName = tmpUser.LastName
	user.Username = tmpUser.Username
	user.Password = string(hash[:])

	stmt, err := db.Prepare("INSERT INTO `users`(`firstName`, `lastName`, `userName`, `password`) VALUES(?,?,?,?);")
	if err != nil {
		fmt.Println("Error preparing the query statement: ", err)
		http.Error(w, "Sorry, it looks like there was an error saving your account info.  Please try again", 404)
		return
	}
	_, err = stmt.Exec(user.FirstName, user.LastName, user.Username, user.Password)
	if err != nil {
		log.Println("Error Creating Record", err)
		http.Error(w, "Sorry, it looks like there was an error saving your account info.  Please try again", 404)
		return
	}

	w.Write([]byte("OK!"))

}

func Login(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body:", err)
		http.Error(w, "Error reading request body", 404)
		return
	}

	var tmpUser LoginPayload
	err = json.Unmarshal(body, &tmpUser)
	if err != nil {
		log.Println("Error unmarshaling into user: ", err)
		http.Error(w, "Error unmarshaling into user", 404)
		return
	}

	password := tmpUser.Password
	var user User
	err = db.Get(&user, "SELECT * FROM users WHERE username = ?", tmpUser.Username)

	if err != nil {
		log.Println("Error getting user: ", err)
		http.Error(w, "Error getting user.", 404)
		return
	}

	hash := user.Password

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		log.Println("Error when comparing password and hash")
		http.Error(w, "Unfortunately that password doesn't match our records.  Please try again", 404)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func main() {
	log.Println("Own Process Identifier: ", strconv.Itoa(os.Getpid()))
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
	//router.Headers("Access-Control-Allow-Origin", "*")
	//---------------------------
	// Health Check
	//---------------------------
	router.HandleFunc("/health", healthCheck)

	//---------------------------
	// Main routes
	//---------------------------
	//api := router.PathPrefix("/api").Headers("Access-Control-Allow-Origin", "*").Subrouter()
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/", sayHelloName)
	api.HandleFunc("/login", Login).Methods("POST")
	api.HandleFunc("/createAccount", createAccount).Methods("POST")
	//api.HandleFunc("/signUp", signUp).Methods("POST")
	api.HandleFunc("/books/{userId}", GetBooks).Methods("GET")
	api.HandleFunc("/book/title/{title}", GetBookByTitle).Methods("GET")
	api.HandleFunc("/books/author/{author}", GetBooksByAuthor).Methods("GET")
	api.HandleFunc("/book", AddBook).Methods("POST")
	server := &http.Server{
		Handler: &MyServer{router},
		Addr:    "127.0.0.1:" + port,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Server started on port: " + port)
	log.Fatal(server.ListenAndServe()) // pass the router as the 2nd argument to ListenAndServe
}

type MyServer struct {
	r *mux.Router
}

// This is to get CORS to work on OPTIONS.  There has to be a better way, yeah???
func (s *MyServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", origin)
		rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		rw.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}
	// Stop here if its Preflighted OPTIONS request
	if req.Method == "OPTIONS" {
		return
	}
	// Lets Gorilla work
	s.r.ServeHTTP(rw, req)
}

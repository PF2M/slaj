package main

// Import dependencies.
import (
	// internals
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	// externals
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB
var err error

var templates = template.Must(template.ParseFiles("views/index.html", "views/header.html", "views/footer.html", "views/communities.html", "views/post.html", "views/create_post.html"))

// Main function.
func main() {

	// Connect to the database.
	db, err = sql.Open("mysql", "slaj:slaj@tcp(127.0.0.1:3306)/slaj")
	if err != nil {

		// we were unable to connect to the database
		log.Printf("[err]: unable to connect to the database...\n")
		log.Printf("       %v\n", err)
		os.Exit(1)

	}

	// ping the database once to ensure that it is available
	err = db.Ping()
	if err != nil {

		// we were unable to ping it
		log.Printf("[err]: unable to ping the database...\n")
		log.Printf("       %v\n", err)
		os.Exit(1)

	}

	// close the database connection after this function exits
	defer db.Close()

	// Initialize routes.
	r := mux.NewRouter()

	// Index route.
	r.HandleFunc("/", index)

	// Auth routes.
	r.HandleFunc("/act/register", register)
	r.HandleFunc("/act/login", login)
	r.HandleFunc("/act/logout", logout)

	// Post routes.
	r.HandleFunc("/posts/{id:[0-9]+}", showPost)

	// Community routes.
	r.HandleFunc("/communities/{id:[0-9]+}", showCommunity)
	r.HandleFunc("/communities/{id:[0-9]+}/posts", createPost).Methods("POST")

	// Serve static assets.
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	// tell the http server to handle routing with the router we just made
	http.Handle("/", r)

	// tell the person who started this that we have started the server
	log.Printf("listening on :8080")

	// start the server
	http.ListenAndServe(":8080", nil)

}

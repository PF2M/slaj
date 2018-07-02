package main

// Import dependencies.
import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	sessions "github.com/kataras/go-sessions"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var err error

var templates = template.Must(template.ParseFiles("views/index.html"))

// Variable declarations for users.
type user struct {
	ID                int
	Username          string
	Nickname          string
	Avatar            string
	Email             string
	Password          string
	IP                string
	Level             int
	Role              int
	LastSeen          string
	Color             string
	YeahNotifications bool
}

// Initialize database.
func connect() {
	db, err = sql.Open("mysql", "root:password@tcp(127.0.0.1)/slaj")
	if err != nil {
		log.Fatalln(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}
}

// Define routes.
func routes() {
	// Index route.
	http.HandleFunc("/", index)
	// Auth routes.
	http.HandleFunc("/act/register", register)
	http.HandleFunc("/act/login", login)
	http.HandleFunc("/act/logout", logout)
	// Serve static assets.
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
}

// Main function.
func main() {
	// Connect to database.
	connect()
	// Initialize routes.
	routes()

	defer db.Close()

	http.ListenAndServe(":8080", nil)
}

// Checks for errors, returns boolean.
func checkErr(w http.ResponseWriter, r *http.Request, err error) bool {
	if err != nil {

		fmt.Println(r.Host + r.URL.Path)

		http.Redirect(w, r, r.Host+r.URL.Path, 301)
		return false
	}

	return true
}

// Find a user by username.
func QueryUser(username string) user {
	var users = user{}
	err = db.QueryRow(`
		SELECT id,
		username,
		nickname,
		avatar,
		email,
		password,
		ip,
		level,
		role,
		last_seen,
		color,
		yeah_notifications
		FROM users WHERE username=?
		`, username).
		Scan(
			&users.ID,
			&users.Username,
			&users.Nickname,
			&users.Avatar,
			&users.Email,
			&users.Password,
			&users.IP,
			&users.Level,
			&users.Role,
			&users.LastSeen,
			&users.Color,
			&users.YeahNotifications,
		)
	return users
}

func index(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	if len(session.GetString("username")) == 0 {
		http.Redirect(w, r, "/act/login", 301)
	}

	var data = map[string]string{
		"username": session.GetString("username"),
	}

	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return
}

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "views/auth/register.html")
		return
	}
	// Define user registration info.
	username := r.FormValue("username")
	nickname := r.FormValue("nickname")
	avatar := r.FormValue("avatar")
	email := r.FormValue("email")
	password := r.FormValue("password")
	ip := r.Header.Get("X-Forwarded-For")
	level := "0"
	role := "0"
	last_seen := time.Now()
	color := ""
	yeah_notifications := "1"

	users := QueryUser(username)

	if (user{}) == users {
		// Let's hash the password. We're using bcrypt for this.
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		if len(hashedPassword) != 0 && checkErr(w, r, err) {
			// Prepare the statement.
			stmt, err := db.Prepare("INSERT users SET username=?, nickname=?, avatar=?, email=?, password=?, ip=?, level=?, role=?, last_seen=?, color=?, yeah_notifications=?")
			if err == nil {
				// If there's no errors, we can go ahead and execute the statement.
				_, err := stmt.Exec(&username, &nickname, &avatar, &email, &hashedPassword, &ip, &level, &role, &last_seen, &color, &yeah_notifications)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				users := QueryUser(username)

				session := sessions.Start(w, r)
				session.Set("username", users.Username)
				http.Redirect(w, r, "/", 302)
			}
		} else {
			http.Redirect(w, r, "/act/register", 302)
		}
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	// Start the session.
	session := sessions.Start(w, r)
	if len(session.GetString("username")) != 0 && checkErr(w, r, err) {
		// Redirect to index page if the user isn't signed in. Will remove later.
		http.Redirect(w, r, "/", 302)
	}

	if r.Method != "POST" {
		http.ServeFile(w, r, "views/auth/login.html")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	users := QueryUser(username)

	// Compare inputted password to the password in the database. If they're the same, return nil.
	var password_compare = bcrypt.CompareHashAndPassword([]byte(users.Password), []byte(password))

	if password_compare == nil {
		session := sessions.Start(w, r)
		session.Set("username", users.Username)
		http.Redirect(w, r, "/", 302)
	} else {
		http.Redirect(w, r, "/act/login", 302)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	session.Clear()
	sessions.Destroy(w, r)
	http.Redirect(w, r, "/", 302)
}

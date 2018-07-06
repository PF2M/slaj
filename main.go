package main

// Import dependencies.
import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	sessions "github.com/kataras/go-sessions"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var err error

var templates = template.Must(template.ParseFiles("views/index.html", "views/header.html", "views/footer.html", "views/communities.html", "views/create_post.html"))

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

// Variable declarations for posts.
type post struct {
	ID             int
	CreatedBy      int
	CreatedAt      string
	Body           string
	Image          string
	URL            string
	IsSpoiler      bool
	IsRm           bool
	IsRmByAdmin    bool
	PosterUsername string
	PosterNickname string
	PosterIcon     string
}

// Variable declarations for communities.
type community struct {
	ID            int
	Title         string
	Description   string
	Icon          string
	Banner        string
	IsFeatured    bool
	DeveloperOnly bool
	StaffOnly     bool
	IsRm          bool
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
	// Initialize router.
	r := mux.NewRouter()
	// Index route.
	r.HandleFunc("/", index)
	// Auth routes.
	r.HandleFunc("/act/register", register)
	r.HandleFunc("/act/login", login)
	r.HandleFunc("/act/logout", logout)
	// Community routes.
	r.HandleFunc("/communities/{id:[0-9]+}", showCommunity)
	r.HandleFunc("/communities/{id:[0-9]+}/posts", createPost).Methods("POST")
	// Serve static assets.
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.Handle("/", r)
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

// Find a post by ID.
func QueryPost(id int) post {
	var posts = post{}
	err = db.QueryRow(`
		SELECT id,
		created_by,
		created_at,
		body,
		image,
		url,
		is_spoiler,
		is_rm,
		is_rm_by_admin
		FROM posts WHERE id = ?
		`, id).
		Scan(
			&posts.ID,
			&posts.CreatedBy,
			&posts.CreatedAt,
			&posts.Body,
			&posts.Image,
			&posts.URL,
			&posts.IsSpoiler,
			&posts.IsRm,
			&posts.IsRmByAdmin,
		)
	return posts
}

// Find a community by ID.
func QueryCommunity(id string) community {
	var communities = community{}
	err = db.QueryRow(`
		SELECT id,
		title,
		description,
		icon,
		banner,
		is_featured,
		developer_only,
		staff_only,
		rm
		FROM communities WHERE id = ?
		`, id).
		Scan(
			&communities.ID,
			&communities.Title,
			&communities.Description,
			&communities.Icon,
			&communities.Banner,
			&communities.IsFeatured,
			&communities.DeveloperOnly,
			&communities.StaffOnly,
			&communities.IsRm,
		)
	return communities
}

func index(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	if len(session.GetString("username")) == 0 {
		http.Redirect(w, r, "/act/login", 301)
	}

	users := QueryUser(session.GetString("username"))

	featured_rows, _ := db.Query("SELECT id, title, icon, banner FROM communities WHERE is_featured = 1 LIMIT 4")
	var featured []community

	for featured_rows.Next() {
		var row = community{}

		err = featured_rows.Scan(&row.ID, &row.Title, &row.Icon, &row.Banner)
		if err != nil {
			fmt.Println(err)
		}
		featured = append(featured, row)
	}

	community_rows, _ := db.Query("SELECT id, title, icon, banner FROM communities ORDER BY id DESC LIMIT 6")
	var communities []community

	for community_rows.Next() {
		var row = community{}

		err = community_rows.Scan(&row.ID, &row.Title, &row.Icon, &row.Banner)
		if err != nil {
			fmt.Println(err)
		}
		communities = append(communities, row)
	}

	pjax := r.Header.Get("X-PJAX") == ""

	var data = map[string]interface{}{
		"Title":       "Communities",
		"Pjax":        pjax,
		"User":        users,
		"Featured":    featured,
		"Communities": communities,
	}
	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return
}

func showCommunity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("cache-control", "no-store, no-cache, must-revalidate")
	session := sessions.Start(w, r)

	if len(session.GetString("username")) == 0 {
		http.Redirect(w, r, "/act/login", 301)
	}

	users := QueryUser(session.GetString("username"))

	id := strings.Split(r.URL.RequestURI(), "/communities/")
	communities := QueryCommunity(id[1])

	post_rows, _ := db.Query("SELECT posts.id, created_by, created_at, body, username, nickname, avatar FROM posts LEFT JOIN users ON users.id = created_by WHERE community_id = ? ORDER BY created_at DESC LIMIT 50", id[1])
	var posts []post

	for post_rows.Next() {
		var row = post{}

		err = post_rows.Scan(&row.ID, &row.CreatedBy, &row.CreatedAt, &row.Body, &row.PosterUsername, &row.PosterNickname, &row.PosterIcon)
		if err != nil {
			fmt.Println(err)
		}
		posts = append(posts, row)
	}

	pjax := r.Header.Get("X-PJAX") == ""

	var data = map[string]interface{}{
		"Title":     communities.Title,
		"Pjax":      pjax,
		"User":      users,
		"Community": communities,
		"Posts":     posts,
	}

	err := templates.ExecuteTemplate(w, "communities.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return
}

func createPost(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)

	user_id := session.GetString("user_id")
	community_id := r.FormValue("community")
	body := r.FormValue("body")

	stmt, err := db.Prepare("INSERT posts SET created_by=?, community_id=?, body=?")
	if err == nil {
		// If there's no errors, we can go ahead and execute the statement.
		_, err := stmt.Exec(&user_id, &community_id, &body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var posts = post{}
		db.QueryRow("SELECT posts.id, created_by, created_at, body, username, nickname, avatar FROM posts LEFT JOIN users ON users.id = created_by WHERE created_by = ? ORDER BY created_at DESC LIMIT 1", user_id).
			Scan(&posts.ID, &posts.CreatedBy, &posts.CreatedAt, &posts.Body, &posts.PosterUsername, &posts.PosterNickname, &posts.PosterIcon)

		var data = map[string]interface{}{
			"Post": posts,
		}

		err = templates.ExecuteTemplate(w, "create_post.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

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
		session.Set("user_id", users.ID)
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

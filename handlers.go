/*

handlers.go

various handlers for the routes of slaj

*/

package main

import (
	// internals
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	// externals
	"github.com/gorilla/mux"
	sessions "github.com/kataras/go-sessions"
	"golang.org/x/crypto/bcrypt"
)

// the handler for the main page
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

// the handler for community pages
func showCommunity(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("cache-control", "no-store, no-cache, must-revalidate")
	session := sessions.Start(w, r)

	if len(session.GetString("username")) == 0 {

		http.Redirect(w, r, "/act/login", 301)

	}

	users := QueryUser(session.GetString("username"))

	id := strings.Split(r.URL.RequestURI(), "/communities/")
	communities := QueryCommunity(id[1])

	post_rows, _ := db.Query("SELECT posts.id, created_by, created_at, body, image, username, nickname, avatar FROM posts LEFT JOIN users ON users.id = created_by WHERE community_id = ? ORDER BY created_at DESC LIMIT 50", id[1])
	var posts []post

	for post_rows.Next() {

		var row = post{}
		var timestamp time.Time

		err = post_rows.Scan(&row.ID, &row.CreatedBy, &timestamp, &row.Body, &row.Image, &row.PosterUsername, &row.PosterNickname, &row.PosterIcon)
		row.CreatedAt = humanTiming(timestamp)
		if err != nil {
			fmt.Println(err)
		}
		db.QueryRow("SELECT COUNT(id) FROM comments WHERE post = ?", row.ID).Scan(&row.CommentCount)
		db.QueryRow("SELECT comments.id, created_at, body, username, nickname, avatar FROM comments LEFT JOIN users ON users.id = created_by WHERE post = ? ORDER BY created_at DESC LIMIT 1", row.ID).
			Scan(&row.CommentPreview.ID, &timestamp, &row.CommentPreview.Body, &row.CommentPreview.CommenterUsername, &row.CommentPreview.CommenterNickname, &row.CommentPreview.CommenterIcon)
		row.CommentPreview.CreatedAt = humanTiming(timestamp)

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

// the handler for a specific post
func showPost(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("cache-control", "no-store, no-cache, must-revalidate")
	session := sessions.Start(w, r)

	if len(session.GetString("username")) == 0 {
		http.Redirect(w, r, "/act/login", 301)
	}

	users := QueryUser(session.GetString("username"))

	id := strings.Split(r.URL.RequestURI(), "/posts/")

	var posts = post{}
	var timestamp time.Time

	db.QueryRow("SELECT posts.id, created_by, community_id, created_at, body, image, username, nickname, avatar FROM posts LEFT JOIN users ON users.id = created_by WHERE posts.id = ?", id[1]).
		Scan(&posts.ID, &posts.CreatedBy, &posts.CommunityID, &timestamp, &posts.Body, &posts.Image, &posts.PosterUsername, &posts.PosterNickname, &posts.PosterIcon)
	posts.CreatedAt = humanTiming(timestamp)

	db.QueryRow("SELECT COUNT(id) FROM comments WHERE post = ?", id[1]).Scan(&posts.CommentCount)

	comment_rows, _ := db.Query("SELECT comments.id, created_by, created_at, body, image, username, nickname, avatar FROM comments LEFT JOIN users ON users.id = created_by WHERE post = ? ORDER BY created_at ASC", id[1])
	var comments []comment

	for comment_rows.Next() {

		var row = comment{}
		var timestamp time.Time

		err = comment_rows.Scan(&row.ID, &row.CreatedBy, &timestamp, &row.Body, &row.Image, &row.CommenterUsername, &row.CommenterNickname, &row.CommenterIcon)
		row.CreatedAt = humanTiming(timestamp)
		if err != nil {

			fmt.Println(err)

		}
		comments = append(comments, row)

	}

	community := QueryCommunity(strconv.Itoa(posts.CommunityID))
	pjax := r.Header.Get("X-PJAX") == ""

	var data = map[string]interface{}{
		"Title":     posts.CreatedBy,
		"Pjax":      pjax,
		"User":      users,
		"Community": community,
		"Post":      posts,
		"Comments":  comments,
	}

	err := templates.ExecuteTemplate(w, "post.html", data)
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)

	}

	return

}

// the handler for post creation
func createPost(w http.ResponseWriter, r *http.Request) {

	session := sessions.Start(w, r)

	user_id := session.GetString("user_id")
	community_id := r.FormValue("community")
	body := r.FormValue("body")
	image := r.FormValue("image")
	url := r.FormValue("url")

	if len(body) > 2000 {
		http.Error(w, "Your post is too long. (2000 characters maximum)", http.StatusBadRequest)
		return
	}
	if len(body) == 0 && len(image) == 0 {
		http.Error(w, "Your post is empty.", http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare("INSERT posts SET created_by=?, community_id=?, body=?, image=?, url=?")
	if err == nil {

		// If there's no errors, we can go ahead and execute the statement.
		_, err := stmt.Exec(&user_id, &community_id, &body, &image, &url)
		if err != nil {

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		}

		var posts = post{}
		var timestamp time.Time

		db.QueryRow("SELECT posts.id, created_by, created_at, body, image, username, nickname, avatar FROM posts LEFT JOIN users ON users.id = created_by WHERE created_by = ? ORDER BY created_at DESC LIMIT 1", user_id).
			Scan(&posts.ID, &posts.CreatedBy, &timestamp, &posts.Body, &posts.Image, &posts.PosterUsername, &posts.PosterNickname, &posts.PosterIcon)
		posts.CreatedAt = humanTiming(timestamp)

		var data = map[string]interface{}{

			"Post": posts,
		}

		err = templates.ExecuteTemplate(w, "create_post.html", data)

		if err != nil {

			http.Error(w, err.Error(), http.StatusInternalServerError)

		}

		var postTpl bytes.Buffer

		templates.ExecuteTemplate(&postTpl, "create_post.html", data)

		var msg wsMessage

		msg.Type = "post"
		msg.Content = postTpl.String()

		for client := range clients {
			if clients[client].OnPage == "/communities/"+community_id && clients[client].UserID != strconv.Itoa(posts.CreatedBy) {
				err := client.WriteJSON(msg)
				if err != nil {
					fmt.Println(err)
					client.Close()
					delete(clients, client)
				}
			}
		}

		return

	}

}

// the handler for comment creation
func createComment(w http.ResponseWriter, r *http.Request) {

	session := sessions.Start(w, r)

	vars := mux.Vars(r)

	post_id := vars["id"]
	user_id := session.GetString("user_id")
	body := r.FormValue("body")
	image := r.FormValue("image")
	url := r.FormValue("url")

	if len(body) > 2000 {
		http.Error(w, "Your comment is too long. (2000 characters maximum)", http.StatusBadRequest)
		return
	}
	if len(body) == 0 && len(image) == 0 {
		http.Error(w, "Your comment is empty.", http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare("INSERT comments SET created_by=?, post=?, body=?, image=?, url=?")
	if err == nil {

		// If there's no errors, we can go ahead and execute the statement.
		_, err := stmt.Exec(&user_id, &post_id, &body, &image, &url)
		if err != nil {

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		}

		var comments = comment{}
		var timestamp time.Time

		db.QueryRow("SELECT comments.id, created_by, created_at, body, image, username, nickname, avatar FROM comments LEFT JOIN users ON users.id = created_by WHERE created_by = ? ORDER BY created_at DESC LIMIT 1", user_id).
			Scan(&comments.ID, &comments.CreatedBy, &timestamp, &comments.Body, &comments.Image, &comments.CommenterUsername, &comments.CommenterNickname, &comments.CommenterIcon)
		comments.CreatedAt = humanTiming(timestamp)

		var data = map[string]interface{}{

			"Comment": comments,
		}

		err = templates.ExecuteTemplate(w, "create_comment.html", data)

		if err != nil {

			http.Error(w, err.Error(), http.StatusInternalServerError)

		}

		var commentTpl bytes.Buffer
		var commentPreviewTpl bytes.Buffer

		templates.ExecuteTemplate(&commentTpl, "create_comment.html", data)
		templates.ExecuteTemplate(&commentPreviewTpl, "comment_preview.html", data)

		var msg wsMessage
		var community_id string

		db.QueryRow("SELECT community_id FROM posts WHERE id = ?", post_id).Scan(&community_id)

		for client := range clients {
			if clients[client].OnPage == "/posts/"+post_id && clients[client].UserID != strconv.Itoa(comments.CreatedBy) {
				msg.Type = "comment"
				msg.Content = commentTpl.String()
				err := client.WriteJSON(msg)
				if err != nil {
					fmt.Println(err)
					client.Close()
					delete(clients, client)
				}
			} else if clients[client].OnPage == "/communities/"+community_id {
				msg.Type = "commentPreview"
				msg.ID = post_id
				msg.Content = commentPreviewTpl.String()
				err := client.WriteJSON(msg)
				if err != nil {
					fmt.Println(err)
					client.Close()
					delete(clients, client)
				}
			}
		}

		return

	}

}

// Upload an image.
func uploadImage(w http.ResponseWriter, r *http.Request) {
	pretendconfigtype := "kek.gg" // temporary config variables to mimic a configuration file

	switch pretendconfigtype {
	case "kek.gg":
		resp, err := http.Post("https://u.kek.gg/v1/upload-to-kek", "text/plain", r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bodyFixed := strings.Replace(string(body), "?", "", -1) // kek.gg adds a ? to the response, so let's remove that
		bodyParsed := strings.Replace(bodyFixed, "\"", "", -1)  // remove the quotes too while we're here
		w.Write([]byte(bodyParsed))
	}

	return
}

// the handler for showing a single user page
func showUser(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)

	if len(session.GetString("username")) == 0 {
		http.Redirect(w, r, "/act/login", 301)
	}

	username := strings.Split(r.URL.RequestURI(), "/users/")
	user := QueryUser(username[1])
	pjax := r.Header.Get("X-PJAX") == ""

	var data = map[string]interface{}{
		"Title":     user.Username,
		"Pjax":      pjax,
		"User":      user,
	}

	err := templates.ExecuteTemplate(w, "user.html", data)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// the handler for user registration
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
		if len(username) > 32 || len(username) < 3 {
			http.Error(w, "invalid username length sorry br0o0o0o0o0o0", http.StatusBadRequest)
			return
		}

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
				session.Set("user_id", users.ID)
				http.Redirect(w, r, "/", 302)

			}

		} else {

			http.Redirect(w, r, "/act/register", 302)

		}

	}

}

// the handler for logging a user in
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

// the handler for logging a user out
func logout(w http.ResponseWriter, r *http.Request) {

	session := sessions.Start(w, r)
	session.Clear()
	sessions.Destroy(w, r)
	http.Redirect(w, r, "/", 302)

}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)

	// Upgrade initial GET request to a websocket
	ws, _ := upgrader.Upgrade(w, r, nil)

	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	// I need to use a 2nd variable to change .Connected because of a retarded issue in golang.
	client := clients[ws]
	client.Connected = true
	client.UserID = session.GetString("user_id")
	clients[ws] = client

	fmt.Println("new connection")

	for {
		var msg wsMessage
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			fmt.Println(err)
			delete(clients, ws)
			break
		}

		if msg.Type == "onPage" {
			client.OnPage = msg.Content
			clients[ws] = client

			fmt.Println(clients[ws].OnPage)
		}
	}
}

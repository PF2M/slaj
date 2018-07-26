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

		http.Redirect(w, r, "/login", 301)

	}

	currentUser := QueryUser(session.GetString("username"))

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
	featured_rows.Close()

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
	community_rows.Close()

	pjax := r.Header.Get("X-PJAX") == ""

	var data = map[string]interface{}{
		"Title":       "Communities",
		"Pjax":        pjax,
		"CurrentUser": currentUser,
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

		http.Redirect(w, r, "/login", 301)

	}

	currentUser := QueryUser(session.GetString("username"))

	vars := mux.Vars(r)
	community_id := vars["id"]
	communities := QueryCommunity(community_id)
	offset, _ := strconv.Atoi(r.FormValue("offset"))

	post_rows, _ := db.Query("SELECT posts.id, created_by, created_at, body, image, username, nickname, avatar, online FROM posts INNER JOIN users ON users.id = created_by WHERE community_id = ? ORDER BY created_at DESC LIMIT 25 OFFSET ?", &community_id, &offset)
	var posts []*post

	for post_rows.Next() {

		var row = &post{}
		var timestamp time.Time
		var yeahed int

		err = post_rows.Scan(&row.ID, &row.CreatedBy, &timestamp, &row.Body, &row.Image, &row.PosterUsername, &row.PosterNickname, &row.PosterIcon, &row.PosterOnline)
		row.CreatedAt = humanTiming(timestamp)
		if err != nil {
			fmt.Println(err)
		}

		// Check if the post has been yeahed.
		db.QueryRow("SELECT id FROM yeahs WHERE yeah_post = ? AND yeah_by = ? AND on_comment=0 LIMIT 1", row.ID, currentUser.ID).Scan(&yeahed)
		if yeahed != 0 {
			row.Yeahed = true
		}

		db.QueryRow("SELECT COUNT(*) FROM yeahs WHERE yeah_post = ? AND on_comment=0", row.ID).Scan(&row.YeahCount)
		db.QueryRow("SELECT COUNT(*) FROM comments WHERE post = ?", row.ID).Scan(&row.CommentCount)
		db.QueryRow("SELECT comments.id, created_at, body, username, nickname, avatar, online FROM comments INNER JOIN users ON users.id = created_by WHERE post = ? ORDER BY created_at DESC LIMIT 1", row.ID).
			Scan(&row.CommentPreview.ID, &timestamp, &row.CommentPreview.Body, &row.CommentPreview.CommenterUsername, &row.CommentPreview.CommenterNickname, &row.CommentPreview.CommenterIcon, &row.CommentPreview.CommenterOnline)
		row.CommentPreview.CreatedAt = humanTiming(timestamp)

		posts = append(posts, row)

	}
	post_rows.Close()

	offset += 25
	pjax := r.Header.Get("X-PJAX") == ""

	var data = map[string]interface{}{
		"Title":       communities.Title,
		"Pjax":        pjax,
		"CurrentUser": currentUser,
		"Community":   communities,
		"Offset":      offset,
		"Posts":       posts,
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
		http.Redirect(w, r, "/login", 301)
	}

	currentUser := QueryUser(session.GetString("username"))
	vars := mux.Vars(r)
	post_id := vars["id"]

	var posts = post{}
	var timestamp time.Time
	var yeahed string

	db.QueryRow("SELECT posts.id, created_by, community_id, created_at, body, image, username, nickname, avatar, online FROM posts LEFT JOIN users ON users.id = created_by WHERE posts.id = ?", post_id).
		Scan(&posts.ID, &posts.CreatedBy, &posts.CommunityID, &timestamp, &posts.Body, &posts.Image, &posts.PosterUsername, &posts.PosterNickname, &posts.PosterIcon, &posts.PosterOnline)
	posts.CreatedAt = humanTiming(timestamp)

	db.QueryRow("SELECT id FROM yeahs WHERE yeah_post = ? AND yeah_by = ? AND on_comment=0", posts.ID, currentUser.ID).Scan(&yeahed)
	if yeahed != "" {
		posts.Yeahed = true
	}

	db.QueryRow("SELECT COUNT(id) FROM yeahs WHERE yeah_post = ? AND on_comment=0", post_id).Scan(&posts.YeahCount)
	db.QueryRow("SELECT COUNT(id) FROM comments WHERE post = ?", post_id).Scan(&posts.CommentCount)

	yeah_rows, _ := db.Query("SELECT yeahs.id, username, avatar FROM yeahs LEFT JOIN users ON users.id = yeah_by WHERE yeah_post = ? AND yeah_by != ? AND on_comment=0 ORDER BY yeahs.id DESC", post_id, currentUser.ID)
	var yeahs []yeah

	for yeah_rows.Next() {

		var row = yeah{}

		err = yeah_rows.Scan(&row.ID, &row.Username, &row.Avatar)
		if err != nil {
			fmt.Println(err)
		}
		yeahs = append(yeahs, row)

	}
	yeah_rows.Close()

	comment_rows, _ := db.Query("SELECT comments.id, created_by, created_at, body, image, username, nickname, avatar, online FROM comments LEFT JOIN users ON users.id = created_by WHERE post = ? ORDER BY created_at ASC", post_id)
	var comments []comment

	for comment_rows.Next() {

		var row = comment{}
		var timestamp time.Time

		err = comment_rows.Scan(&row.ID, &row.CreatedBy, &timestamp, &row.Body, &row.Image, &row.CommenterUsername, &row.CommenterNickname, &row.CommenterIcon, &row.CommenterOnline)
		row.CreatedAt = humanTiming(timestamp)
		if err != nil {
			fmt.Println(err)
		}

		db.QueryRow("SELECT 1 FROM yeahs WHERE yeah_post = ? AND yeah_by = ? AND on_comment=1", row.ID, currentUser.ID).Scan(&row.Yeahed)

		db.QueryRow("SELECT COUNT(id) FROM yeahs WHERE yeah_post = ? AND on_comment=1", row.ID).Scan(&row.YeahCount)

		comments = append(comments, row)

	}
	comment_rows.Close()

	community := QueryCommunity(strconv.Itoa(posts.CommunityID))
	pjax := r.Header.Get("X-PJAX") == ""

	var data = map[string]interface{}{
		"Title":       posts.PosterNickname + "'s post",
		"Pjax":        pjax,
		"CurrentUser": currentUser,
		"Community":   community,
		"Post":        posts,
		"Yeahs":       yeahs,
		"Comments":    comments,
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
			// This is sent to the user who created the post so they can't yeah it.
			"CanYeah": false,
			"Post":    posts,
		}

		err = templates.ExecuteTemplate(w, "create_post.html", data)

		if err != nil {

			http.Error(w, err.Error(), http.StatusInternalServerError)

		}

		var postTpl bytes.Buffer

		// This will be sent other users so they can yeah it.
		data["CanYeah"] = true

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
			// This is sent to the user who created the comment so they can't yeah it.
			"CanYeah": false,
			"Comment": comments,
		}

		err = templates.ExecuteTemplate(w, "create_comment.html", data)

		if err != nil {

			http.Error(w, err.Error(), http.StatusInternalServerError)

		}

		var commentTpl bytes.Buffer
		var commentPreviewTpl bytes.Buffer

		// This will be sent other users so they can yeah it.
		data["CanYeah"] = true

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

// Creating post yeahs, I dont know why I need a comment here but everything else had one.
func createPostYeah(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	vars := mux.Vars(r)

	post_id := vars["id"]
	user_id := session.GetString("user_id")

	var post_by string
	// We'll need the community id later for websockets.
	var community_id string
	var yeah_exists string

	db.QueryRow("SELECT created_by, community_id FROM posts WHERE id = ?", post_id).Scan(&post_by, &community_id)

	// Check if the post exists, if it doesn't the yeah wont be added.
	if post_by != "" {
		db.QueryRow("SELECT id FROM yeahs WHERE yeah_post = ? AND yeah_by = ? AND on_comment=0", post_id, user_id).Scan(&yeah_exists)

		// Check if the post has already been yeahed or if its being yeahed by the creator, if it is the yeah wont be added.
		if yeah_exists == "" && post_by != user_id {

			stmt, err := db.Prepare("INSERT yeahs SET yeah_post=?, yeah_by=?, on_comment=0")
			if err == nil {

				// If there's no errors, we can go ahead and execute the statement.
				_, err := stmt.Exec(&post_id, &user_id)
				if err != nil {

					http.Error(w, err.Error(), http.StatusInternalServerError)
					return

				} else {
					// Websockets
					var msg wsMessage
					var yeahs = yeah{}

					db.QueryRow("SELECT yeahs.id, username, avatar FROM yeahs LEFT JOIN users ON users.id = yeah_by WHERE yeah_by = ? ORDER BY yeahs.id DESC LIMIT 1", user_id).
						Scan(&yeahs.ID, &yeahs.Username, &yeahs.Avatar)

					msg.Type = "postYeah"
					msg.ID = post_id
					// I dont think we need a separate template for such a small amount of html.
					msg.Content = fmt.Sprintf("<a href=\"/users/%s\" id=\"%d\" class=\"post-permalink-feeling-icon\"><img src=\"%s\" class=\"user-icon\"></a>", yeahs.Username, yeahs.ID, yeahs.Avatar)

					for client := range clients {
						if (clients[client].OnPage == "/communities/"+community_id || clients[client].OnPage == "/posts/"+post_id) && clients[client].UserID != user_id {
							err := client.WriteJSON(msg)
							if err != nil {
								fmt.Println(err)
								client.Close()
								delete(clients, client)
							}
						}
					}

				}
			}

		}
	}
}

func deletePostYeah(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	vars := mux.Vars(r)

	var yeah_id string
	var community_id string
	post_id := vars["id"]
	user_id := session.GetString("user_id")

	db.QueryRow("SELECT yeahs.id, posts.community_id FROM yeahs INNER JOIN posts ON posts.id = yeahs.yeah_post WHERE yeah_post = ? AND yeah_by = ? AND on_comment=0", post_id, user_id).Scan(&yeah_id, &community_id)

	if yeah_id != "" {
		stmt, _ := db.Prepare("DELETE FROM yeahs WHERE yeah_post=? AND yeah_by=? AND on_comment=0")
		stmt.Exec(&post_id, &user_id)

		var msg wsMessage
		msg.Type = "postUnyeah"
		msg.ID = post_id
		msg.Content = yeah_id

		for client := range clients {
			if (clients[client].OnPage == "/communities/"+community_id || clients[client].OnPage == "/posts/"+post_id) && clients[client].UserID != user_id {
				err := client.WriteJSON(msg)
				if err != nil {
					fmt.Println(err)
					client.Close()
					delete(clients, client)
				}
			}
		}
	}
}

// Creating comment yeahs, I dont know why I need a comment here but everything else had one.
func createCommentYeah(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	vars := mux.Vars(r)

	comment_id := vars["id"]
	user_id := session.GetString("user_id")

	var comment_by string
	// We'll need the post id later for websockets.
	var post_id string
	var yeah_exists string

	db.QueryRow("SELECT created_by, post FROM comments WHERE id = ?", comment_id).Scan(&comment_by, &post_id)

	// Check if the comment exists, if it doesn't the yeah wont be added.
	if comment_by != "" {
		db.QueryRow("SELECT id FROM yeahs WHERE yeah_post = ? AND yeah_by = ? AND on_comment = 1", comment_id, user_id).Scan(&yeah_exists)

		// Check if the comment has already been yeahed or if its being yeahed by the creator, if it is the yeah wont be added.
		if yeah_exists == "" && comment_by != user_id {

			stmt, err := db.Prepare("INSERT yeahs SET yeah_post=?, yeah_by=?, on_comment=1")
			if err == nil {

				// If there's no errors, we can go ahead and execute the statement.
				_, err := stmt.Exec(&comment_id, &user_id)
				if err != nil {

					http.Error(w, err.Error(), http.StatusInternalServerError)
					return

				} else {
					// Websockets
					var msg wsMessage
					var yeahs = yeah{}

					db.QueryRow("SELECT yeahs.id, username, avatar FROM yeahs LEFT JOIN users ON users.id = yeah_by WHERE yeah_by = ? ORDER BY yeahs.id DESC LIMIT 1", user_id).
						Scan(&yeahs.ID, &yeahs.Username, &yeahs.Avatar)

					msg.Type = "commentYeah"
					msg.ID = comment_id
					// I dont think we need a separate template for such a small amount of html.
					msg.Content = fmt.Sprintf("<a href=\"/users/%s\" id=\"%d\" class=\"post-permalink-feeling-icon\"><img src=\"%s\" class=\"user-icon\"></a>", yeahs.Username, yeahs.ID, yeahs.Avatar)

					for client := range clients {
						if (clients[client].OnPage == "/posts/"+post_id || clients[client].OnPage == "/comments/"+comment_id) && clients[client].UserID != user_id {
							err := client.WriteJSON(msg)
							if err != nil {
								fmt.Println(err)
								client.Close()
								delete(clients, client)
							}
						}
					}

				}
			}

		}
	}
}

func deleteCommentYeah(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	vars := mux.Vars(r)

	var yeah_id string
	var post_id string
	comment_id := vars["id"]
	user_id := session.GetString("user_id")

	db.QueryRow("SELECT yeahs.id, comments.post FROM yeahs INNER JOIN comments ON comments.id = yeahs.yeah_post WHERE yeah_post = ? AND yeah_by = ? AND on_comment=1", comment_id, user_id).Scan(&yeah_id, &post_id)

	if yeah_id != "" {
		stmt, _ := db.Prepare("DELETE FROM yeahs WHERE yeah_post=? AND yeah_by=? AND on_comment=1")
		stmt.Exec(&comment_id, &user_id)

		var msg wsMessage
		msg.Type = "commentUnyeah"
		msg.ID = comment_id
		msg.Content = yeah_id

		for client := range clients {
			if (clients[client].OnPage == "/posts/"+post_id || clients[client].OnPage == "/comments/"+comment_id) && clients[client].UserID != user_id {
				err := client.WriteJSON(msg)
				if err != nil {
					fmt.Println(err)
					client.Close()
					delete(clients, client)
				}
			}
		}
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
	w.Header().Set("cache-control", "no-store, no-cache, must-revalidate")
	session := sessions.Start(w, r)

	if len(session.GetString("username")) == 0 {
		http.Redirect(w, r, "/login", 301)
	}

	vars := mux.Vars(r)
	username := vars["username"]
	currentUser := QueryUser(session.GetString("username"))
	user := QueryUser(username)
	pjax := r.Header.Get("X-PJAX") == ""
	profile := QueryProfile(user.ID)
	var following bool

	db.QueryRow("SELECT 1 FROM follows WHERE follow_to = ? AND follow_by = ?", user.ID, currentUser.ID).Scan(&following)

	db.QueryRow("SELECT COUNT(*) FROM follows WHERE follow_by = ?", user.ID).Scan(&profile.FollowingCount)
	db.QueryRow("SELECT COUNT(*) FROM follows WHERE follow_to = ?", user.ID).Scan(&profile.FollowerCount)

	db.QueryRow("SELECT COUNT(*) FROM posts WHERE created_by = ?", user.ID).Scan(&profile.PostCount)
	db.QueryRow("SELECT COUNT(*) FROM comments WHERE created_by = ?", user.ID).Scan(&profile.CommentCount)
	db.QueryRow("SELECT COUNT(*) FROM yeahs WHERE yeah_by = ?", user.ID).Scan(&profile.YeahCount)

	post_rows, _ := db.Query("SELECT posts.id, community_id, created_at, body, image, title, icon FROM posts LEFT JOIN communities ON communities.id = community_id WHERE created_by = ? ORDER BY created_at DESC LIMIT 3", user.ID)
	var posts []post

	for post_rows.Next() {

		var row = post{}
		var timestamp time.Time
		var yeahed string

		err = post_rows.Scan(&row.ID, &row.CommunityID, &timestamp, &row.Body, &row.Image, &row.CommunityName, &row.CommunityIcon)
		row.CreatedAt = humanTiming(timestamp)
		if err != nil {
			fmt.Println(err)
		}

		// Check if the post has been yeahed.
		db.QueryRow("SELECT id FROM yeahs WHERE yeah_post = ? AND yeah_by = ? AND on_comment=0", row.ID, currentUser.ID).Scan(&yeahed)
		if yeahed != "" {
			row.Yeahed = true
		}

		db.QueryRow("SELECT COUNT(id) FROM yeahs WHERE yeah_post = ? AND on_comment=0", row.ID).Scan(&row.YeahCount)
		db.QueryRow("SELECT COUNT(id) FROM comments WHERE post = ?", row.ID).Scan(&row.CommentCount)

		posts = append(posts, row)
	}
	post_rows.Close()

	yeah_rows, _ := db.Query("SELECT posts.id, created_by, community_id, created_at, body, image, username, nickname, avatar, online, title, icon FROM yeahs INNER JOIN posts ON posts.id = yeah_post INNER JOIN users ON users.id = posts.created_by INNER JOIN communities ON communities.id = community_id WHERE yeah_by = ? AND on_comment = 0 ORDER BY created_at DESC LIMIT 3", user.ID)
	var yeahs []post

	for yeah_rows.Next() {

		var row = post{}
		var timestamp time.Time
		var yeahed string

		err = yeah_rows.Scan(&row.ID, &row.CreatedBy, &row.CommunityID, &timestamp, &row.Body, &row.Image, &row.PosterUsername, &row.PosterNickname, &row.PosterIcon, &row.PosterOnline, &row.CommunityName, &row.CommunityIcon)
		row.CreatedAt = humanTiming(timestamp)
		if err != nil {
			fmt.Println(err)
		}

		// Check if the post has been yeahed.
		db.QueryRow("SELECT id FROM yeahs WHERE yeah_post = ? AND yeah_by = ? AND on_comment=0", row.ID, currentUser.ID).Scan(&yeahed)
		if yeahed != "" {
			row.Yeahed = true
		}

		db.QueryRow("SELECT COUNT(id) FROM yeahs WHERE yeah_post = ? AND on_comment=0", row.ID).Scan(&row.YeahCount)
		db.QueryRow("SELECT COUNT(id) FROM comments WHERE post = ?", row.ID).Scan(&row.CommentCount)

		yeahs = append(yeahs, row)
	}
	yeah_rows.Close()

	var data = map[string]interface{}{
		"Title":       user.Nickname + "'s profile",
		"Pjax":        pjax,
		"CurrentUser": currentUser,
		"User":        user,
		"Profile":     profile,
		"Following":   following,
		"Posts":       posts,
		"Yeahs":       yeahs,
	}

	err := templates.ExecuteTemplate(w, "user.html", data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Creating follows.
func createFollow(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	vars := mux.Vars(r)

	username := vars["username"]
	current_username := session.GetString("username")

	if username != current_username {
		var user_id int
		db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&user_id)
		current_user_id := session.GetString("user_id")

		stmt, err := db.Prepare("INSERT follows SET follow_to=?, follow_by=?")
		if err == nil {

			// If there's no errors, we can go ahead and execute the statement.
			_, err := stmt.Exec(&user_id, &current_user_id)
			if err != nil {

				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// For some reason Arian's gay javascript needs following_count in the respose so here it is. It's not long I'm not using a template.
			w.Header().Add("content-type", "application/json")
			fmt.Fprint(w, "{\"following_count\": 1}")

			var msg wsMessage
			msg.Type = "follow"

			for client := range clients {
				if clients[client].OnPage == "/users/"+username {
					err := client.WriteJSON(msg)
					if err != nil {
						fmt.Println(err)
						client.Close()
						delete(clients, client)
					}
				}
			}

		}
	}
}

// Delte follows.
func deleteFollow(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	vars := mux.Vars(r)

	username := vars["username"]
	var user_id int
	db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&user_id)
	current_user_id := session.GetString("user_id")

	stmt, _ := db.Prepare("DELETE FROM follows WHERE follow_to=? AND follow_by=?")
	stmt.Exec(&user_id, &current_user_id)

	var msg wsMessage
	msg.Type = "unfollow"

	for client := range clients {
		if clients[client].OnPage == "/users/"+username {
			err := client.WriteJSON(msg)
			if err != nil {
				fmt.Println(err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

// the handler for user registration
func signup(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {

		var data = map[string]interface{}{
			"Title": "Sign Up",
		}

		err := templates.ExecuteTemplate(w, "signup.html", data)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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

				user := users.ID
				created_at := time.Now()
				nnid := ""
				gender := 0
				region := "" // ooh what if we replace this with a country from a GeoIP later????????????????
				comment := ""
				nnid_visibility := 1
				yeah_visibility := 1
				reply_visibility := 0

				stmt, err := db.Prepare("INSERT profiles SET user=?, created_at=?, nnid=?, gender=?, region=?, comment=?, nnid_visibility=?, yeah_visibility=?, reply_visibility=?")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				_, err = stmt.Exec(&user, &created_at, &nnid, &gender, &region, &comment, &nnid_visibility, &yeah_visibility, &reply_visibility)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				session := sessions.Start(w, r)
				session.Set("username", users.Username)
				session.Set("user_id", users.ID)
				http.Redirect(w, r, "/", 302)

			}

		} else {

			http.Redirect(w, r, "/signup", 302)

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
		var data = map[string]interface{}{
			"Title": "Log In",
		}

		err := templates.ExecuteTemplate(w, "login.html", data)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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

		http.Redirect(w, r, "/login", 302)

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

	stmt, _ := db.Prepare("UPDATE users SET online = 1 WHERE id = ?")
	stmt.Exec(&client.UserID)

	var username string
	db.QueryRow("SELECT username FROM users WHERE id = ?", &client.UserID).Scan(&username)

	var msg wsMessage
	msg.Type = "online"
	msg.Content = username

	for client := range clients {
		err := client.WriteJSON(msg)
		if err != nil {
			fmt.Println(err)
			client.Close()
			delete(clients, client)
		}
	}

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

	stmt, _ = db.Prepare("UPDATE users SET online = 0 WHERE id = ?")
	stmt.Exec(&client.UserID)

	msg.Type = "offline"
	msg.Content = username

	for client := range clients {
		err := client.WriteJSON(msg)
		if err != nil {
			fmt.Println(err)
			client.Close()
			delete(clients, client)
		}
	}
}

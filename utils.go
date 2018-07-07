/*

utils.go

file that holds various utility functions used in slaj

*/

package main

import (
	"fmt"
	"log"
	"net/http"
<<<<<<< HEAD
	"time"
)

// function for changing ugly gay sql time to stuff people can read
func humanTiming(timestamp time.Time) string {
	now := time.Now()

	if now.Day() == timestamp.Day() && now.Month() == timestamp.Month() && now.Year() == timestamp.Year() {
		return timestamp.Format("Today at 3:04 PM")
	} else if now.Day()-1 == timestamp.Day() && now.Month() == timestamp.Month() && now.Year() == timestamp.Year() {
		return timestamp.Format("Yesterday at 3:04 PM")
	} else if now.Day()-2 == timestamp.Day() && now.Month() == timestamp.Month() && now.Year() == timestamp.Year() {
		return timestamp.Format("Last Monday at 3:04 PM")
	} else {
		return timestamp.Format("01/02/2006 3:04 PM")
	}

}

=======
)

>>>>>>> 6fcb87b263e0038b73555b1ba7d6b719e31e4d04
// function that checks for an error and logs it. it returns either a true or false value
func checkErr(w http.ResponseWriter, r *http.Request, err error) bool {

	// check for an error
	if err != nil {

		// log the error
		log.Printf("[err]: %s%s", r.Host, r.URL.Path)
		log.Printf("       %v\n", err)

		// return an error message to the client
		http.Error(w, fmt.Sprintf("unable to handle request. sorry :(\nerror: %v", err), http.StatusInternalServerError)

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

<<<<<<< HEAD
=======
// Find a post by ID.
func QueryPost(id string) post {
	var posts = post{}
	err = db.QueryRow(`
		SELECT id,
		created_by,
		community_id,
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
			&posts.CommunityID,
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

>>>>>>> 6fcb87b263e0038b73555b1ba7d6b719e31e4d04
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

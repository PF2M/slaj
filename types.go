/*

types.go

file that holds all types/structures declared in slaj

*/

package main

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
	CommunityID    int
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
	CommentCount   int
	CommentPreview comment
}

// Variable declarations for comments.
type comment struct {
	ID                int
	CreatedBy         int
	PostID            int
	CreatedAt         string
	Body              string
	Image             string
	URL               string
	IsSpoiler         bool
	IsRm              bool
	IsRmByAdmin       bool
	CommenterUsername string
	CommenterNickname string
	CommenterIcon     string
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

// Variable declarations for websocket sessions.
type wsSession struct {
	Connected bool
	UserID    string
	OnPage    string
}

// Variable declarations for websocket messages.
type wsMessage struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Content string `json:"content"`
}

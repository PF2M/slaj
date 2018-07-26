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
	Online            bool
	LastSeen          string
	Color             string
	YeahNotifications bool
}

// Variable declarations for profiles.
type profile struct {
	User            int
	CreatedAt       string
	NNID            string
	Gender          int
	Region          string
	Comment         string
	NNIDVisibility  int
	YeahVisibility  int
	ReplyVisibility int
	FollowingCount  int
	FollowerCount   int
	PostCount       int
	CommentCount    int
	YeahCount       int
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
	PosterOnline   bool
	CommunityID    int
	CommunityName  string
	CommunityIcon  string
	Yeahed         bool
	YeahCount      int
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
	CommenterOnline   bool
	Yeahed            bool
	YeahCount         int
}

type yeah struct {
	ID       int
	Username string
	Avatar   string
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

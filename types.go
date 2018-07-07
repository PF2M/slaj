/*

types.go

file that holds all types/structures declared in slaj

*/

package main

<<<<<<< HEAD
=======
// type for the database parts of the config
type databaseConfig struct {
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// type for the entire config
type configType struct {
	Port     int            `json:"port"`
	Database databaseConfig `json:"database"`
}

>>>>>>> 6fcb87b263e0038b73555b1ba7d6b719e31e4d04
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

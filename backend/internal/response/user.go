package response

import (
	"fmt"

	"github.com/mmikhail2001/team_chat/internal/database"
)

type User struct {
	ID        string            `json:"id"`
	Avatar    string            `json:"avatar"`
	Username  string            `json:"username"`
	Status    int               `json:"status"`
	CreatedAt int64             `json:"created_at"`
	Reactions []ReactionMessage `json:"reactions"`
}

type ReactionMessage struct {
	MessageID string `json:"message_id"`
	Reaction  string `json:"reaction"`
}

func GetUrl(user *database.User) string {
	// TODO: что за путь http://localhost/api/avatars/user_id_hex/avatar_id_hex/unknown.jpg ...
	endpoint := fmt.Sprint(user.ID, "/", user.Avatar.ID.Hex(), "/unknown."+user.Avatar.Ext)
	fullUrl := fmt.Sprintf("%s/avatars/%s", URL, endpoint)
	return fullUrl
}

func NewUser(user *database.User, status int) User {
	return User{
		ID:        user.ID,
		Avatar:    GetUrl(user),
		Username:  user.Username,
		Status:    status,
		CreatedAt: user.CreatedAt,
	}
}

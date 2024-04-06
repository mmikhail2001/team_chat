package response

import (
	"fmt"
	"os"

	"github.com/mmikhail2001/team_chat/internal/database"
)

var (
	TLS         = os.Getenv("TLS")
	PUBLIC_HOST = os.Getenv("PUBLIC_HOST")
	URL         = fmt.Sprintf("http%s://%s/api", TLS, PUBLIC_HOST)
)

type Message struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	// а в слое db так ...........
	// AccountID     primitive.ObjectID `bson:"account_id"`
	Author        User                `json:"author"`
	ChannelID     string              `json:"channel_id"`
	ThreadID      string              `json:"thread_id"`
	HasThread     bool                `json:"has_thread"`
	SystemMessage bool                `json:"system_message"`
	CreatedAt     int64               `json:"created_at"`
	EditedAt      int64               `json:"edited_at"`
	Attachments   []Attachment        `json:"attachments"`
	Reactions     []database.Reaction `json:"reactions"`
}

func NewMessage(message *database.Message, user User) Message {
	res_message := Message{
		ID:            message.ID.Hex(),
		Content:       message.Content,
		ChannelID:     message.ChannelID.Hex(),
		ThreadID:      message.ThreadID.Hex(),
		HasThread:     message.HasThread,
		SystemMessage: message.SystemMessage,
		CreatedAt:     message.CreatedAt,
		EditedAt:      message.UpdatedAt,
		Attachments:   []Attachment{},
		Reactions:     message.Reactions,
	}

	if !message.SystemMessage {
		res_message.Author = user
	}

	if len(message.Attachments) > 0 {
		res_attachments := NewAttachments(message)
		res_message.Attachments = res_attachments
	}

	return res_message
}

type Attachment struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	Url         string `json:"url"`
}

func NewAttachments(message *database.Message) []Attachment {
	res_attachments := []Attachment{}
	for _, attachment := range message.Attachments {
		url := fmt.Sprintf("%s/attachments/%s/%s/%s/%s", URL, message.ChannelID.Hex(), message.ID.Hex(), attachment.ID.Hex(), attachment.Filename)
		res_attachment := Attachment{
			ID:          attachment.ID.Hex(),
			Filename:    attachment.Filename,
			Size:        attachment.Size,
			ContentType: attachment.ContentType,
			Url:         url,
		}
		res_attachments = append(res_attachments, res_attachment)
	}

	return res_attachments
}

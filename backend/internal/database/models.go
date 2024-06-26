package database

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID         string               `bson:"_id"`
	Avatar     Avatar               `bson:"avatar"`
	Username   string               `bson:"username"`
	Email      string               `bson:"email"`
	Password   []byte               `bson:"password"`
	CreatedAt  int64                `bson:"created_at"`
	LastLogout int64                `bson:"last_logout"`
	Roles      []primitive.ObjectID `bson:"roles"`
	IsGuest    bool                 `bson:"is_guest"`
}

type Avatar struct {
	ID     primitive.ObjectID `bson:"_id"`
	Avatar string             `bson:"avatar"`
	Type   string             `bson:"type"`
	Ext    string             `bson:"ext"`
}

type Relationship struct {
	ID         primitive.ObjectID `bson:"_id"`
	Type       int                `bson:"type"`
	FromUserID string             `bson:"from_user_id"`
	ToUserID   string             `bson:"to_user_id"`
}

type Channel struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Type       int                `bson:"type,omitempty"`
	Name       string             `bson:"name,omitempty"`
	Icon       Icon               `bson:"icon,omitempty"`
	OwnerID    string             `bson:"owner_id,omitempty"`
	Recipients []string           `bson:"recipients,omitempty"`
	CreatedAt  int64              `bson:"created_at,omitempty"`
	UpdatedAt  int64              `bson:"updated_at,omitempty"`
}

type Icon struct {
	ID   primitive.ObjectID `bson:"_id"`
	Icon string             `bson:"icon"`
	Type string             `bson:"type"`
	Ext  string             `bson:"ext"`
}

type Role struct {
	ID       primitive.ObjectID   `bson:"_id"`
	Name     string               `bson:"name"`
	Channels []primitive.ObjectID `bson:"channels"`
}

type Message struct {
	ID            primitive.ObjectID `bson:"_id"`
	Content       string             `bson:"content"`
	ChannelID     primitive.ObjectID `bson:"channel_id"`
	ThreadID      primitive.ObjectID `bson:"thread_id"`
	HasThread     bool               `bson:"has_thread"`
	AccountID     string             `bson:"account_id"`
	SystemMessage bool               `bson:"system_message"`
	CreatedAt     int64              `bson:"created_at"`
	UpdatedAt     int64              `bson:"updated_at"`
	Attachments   []Attachment       `bson:"attachments,omitempty"`
	Reactions     []Reaction         `bson:"reactions,omitempty"`
}

type Reaction struct {
	Reaction string `bson:"reaction" json:"reaction"`
	UserID   string `bson:"user_id" json:"user_id"`
}

type Attachment struct {
	ID          primitive.ObjectID `bson:"_id"`
	Filename    string             `bson:"filename"`
	Size        int64              `bson:"size"`
	ContentType string             `bson:"content-type"`
	Data        []byte             `bson:"data"`
}

type Pins struct {
	ID        primitive.ObjectID `bson:"_id"`
	ChannelID primitive.ObjectID `bson:"channel_id"`
	MessageID primitive.ObjectID `bson:"message_id"`
	CreatedAt int64              `bson:"created_at"`
}

type Invites struct {
	ID         primitive.ObjectID `bson:"_id"`
	InviteCode string             `bson:"invite_code"`
	ChannelID  primitive.ObjectID `bson:"channel_id"`
	AccountID  string             `bson:"account_id"`
	CreatedAt  time.Time          `bson:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at"`
}

type Ban struct {
	ID         primitive.ObjectID `bson:"_id"`
	BannedUser string             `bson:"banned_user"`
	ChannelID  primitive.ObjectID `bson:"channel_id"`
	BannedBy   string             `bson:"banned_by"`
	Reason     string             `bson:"reason"`
	CreatedAt  int64              `bson:"created_at"`
}

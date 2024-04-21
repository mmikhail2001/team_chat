package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/mmikhail2001/team_chat/internal/database"
	"github.com/mmikhail2001/team_chat/internal/response"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func generatePushEventInfo(data map[string]interface{}) string {
	userName, _ := data["user_name"].(string)
	repoName, _ := data["repository"].(map[string]interface{})["name"].(string)
	commits, _ := data["commits"].([]interface{})

	eventInfo := fmt.Sprintf("event: %s, author: %s, repo: %s, commits:", "push", userName, repoName)
	for i, commit := range commits {
		commitTitle := commit.(map[string]interface{})["title"].(string)
		if i == len(commits)-1 {
			eventInfo += fmt.Sprintf(" %s.", commitTitle)
		} else {
			eventInfo += fmt.Sprintf(" %s,", commitTitle)
		}
	}

	return eventInfo
}

func generateMergeRequestEventInfo(data map[string]interface{}) string {
	userName := data["user"].(map[string]interface{})["username"].(string)
	repoName := data["repository"].(map[string]interface{})["name"].(string)
	lastCommitTitle := data["object_attributes"].(map[string]interface{})["last_commit"].(map[string]interface{})["title"].(string)
	sourceBranch := data["object_attributes"].(map[string]interface{})["source_branch"].(string)
	targetBranch := data["object_attributes"].(map[string]interface{})["target_branch"].(string)

	return fmt.Sprintf("event: merge_request, author: %s, repo: %s, last_commit: %s, source: %s, target: %s", userName, repoName, lastCommitTitle, sourceBranch, targetBranch)
}

func WebhookGitlab(ctx *Context) {
	log.Println("handler WebhookGitlab")

	body, err := io.ReadAll(ctx.Req.Body)
	if err != nil {
		http.Error(ctx.Res, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(ctx.Res, "Error parsing JSON data", http.StatusInternalServerError)
		return
	}

	eventType, _ := data["object_kind"].(string)
	var eventInfo string
	switch eventType {
	case "push":
		eventInfo = generatePushEventInfo(data)
	case "merge_request":
		eventInfo = generateMergeRequestEventInfo(data)
	default:
		ctx.Res.WriteHeader(http.StatusOK)
		return
	}

	log.Println(eventInfo)

	channelsCollection := ctx.Db.Mongo.Collection("channels")
	messagesCollection := ctx.Db.Mongo.Collection("messages")
	var channel database.Channel
	err = channelsCollection.FindOne(context.TODO(), bson.M{"name": "GitLab Events"}).Decode(&channel)
	if err != nil {
		log.Println("GetChannel: FindOne Gitlab: err: ", err)
		ctx.Res.WriteHeader(http.StatusNotFound)
		return
	}

	new_message := database.Message{
		ID:            primitive.NewObjectID(),
		Content:       eventInfo,
		ChannelID:     channel.ID,
		SystemMessage: false,
		CreatedAt:     time.Now().Unix(),
		UpdatedAt:     time.Now().Unix(),
		Reactions:     []database.Reaction{},
		AccountID:     ctx.User.ID,
	}

	_, err = messagesCollection.InsertOne(context.TODO(), new_message)
	if err != nil {
		log.Println("CreateMessage Gitlab: InsertOne: err: ", err)
		ctx.Res.WriteHeader(http.StatusNotFound)
		return
	}

	messageRes := response.NewMessage(&new_message, response.NewUser(&ctx.User, 0))
	ctx.Conn.BroadcastToChannel(channel.ID.Hex(), "MESSAGE_CREATE", messageRes)

	ctx.Res.WriteHeader(http.StatusOK)
}

package main

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID    string               `bson:"_id,omitempty"`
	Email string               `bson:"email"`
	Roles []primitive.ObjectID `bson:"roles"`
}

type Channel struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Icon       Icon               `bson:"icon"`
	OwnerID    string             `bson:"owner_id"`
	Type       int                `bson:"type"`
	Name       string             `bson:"name"`
	Recipients []string           `bson:"recipients"`
}

type Icon struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Icon string             `bson:"icon"`
	Type string             `bson:"type"`
	Ext  string             `bson:"ext"`
}

type Role struct {
	ID       primitive.ObjectID   `bson:"_id,omitempty"`
	Name     string               `bson:"name"`
	Channels []primitive.ObjectID `bson:"channels"`
}

func main() {
	// Установка клиента MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://USERNAME:PASSWORD@127.0.0.1:27018")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Выбор базы данных и коллекции
	db := client.Database("DATABASE_NAME")
	usersCollection := db.Collection("users")
	channelsCollection := db.Collection("channels")
	rolesCollection := db.Collection("roles")

	// Генерация ObjectID для каналов и ролей
	generalChatID := primitive.NewObjectID()
	generalNewsID := primitive.NewObjectID()
	managersChatID := primitive.NewObjectID()
	developersChatID := primitive.NewObjectID()
	projectByOrderChatID := primitive.NewObjectID()

	generalRoleID := primitive.NewObjectID()
	guestRoleID := primitive.NewObjectID()
	managersRoleID := primitive.NewObjectID()
	developersRoleID := primitive.NewObjectID()

	gitlabEventsChatID := primitive.NewObjectID()

	adminID := "1e82445a-8e05-47eb-bd5e-f6b2ecc94fe9"

	_, err = usersCollection.InsertOne(ctx, bson.M{
		"_id":      adminID,
		"email":    "administrator@mail.ru",
		"username": "artur petrov",
		"roles":    []primitive.ObjectID{generalRoleID, managersRoleID, developersRoleID},
	})
	if err != nil {
		log.Fatal(err)
	}

	var administrator User
	err = usersCollection.FindOne(ctx, bson.M{"email": "administrator@mail.ru"}).Decode(&administrator)
	if err != nil {
		log.Fatal(err)
	}

	// Считывание иконок из файловой системы и кодирование их в base64
	icons := map[string]string{
		"general_chat.png":     "",
		"general_news.png":     "",
		"managers.png":         "",
		"developers.jpg":       "",
		"project_by_order.png": "",
		"gitlab.png":           "", // Добавляем gitlab.png
	}
	for filename := range icons {
		imageData, err := ioutil.ReadFile("./icons/" + filename)
		if err != nil {
			log.Fatal(err)
		}
		imageBase64 := base64.StdEncoding.EncodeToString(imageData)
		icons[filename] = imageBase64
	}

	// Создание каналов
	channels := []Channel{
		{
			ID:         generalChatID,
			Icon:       Icon{ID: primitive.NewObjectID(), Icon: icons["general_chat.png"], Type: "image/png", Ext: "png"},
			OwnerID:    administrator.ID,
			Type:       2,
			Name:       "General Chat",
			Recipients: []string{adminID},
		},
		{
			ID:         generalNewsID,
			Icon:       Icon{ID: primitive.NewObjectID(), Icon: icons["general_news.png"], Type: "image/png", Ext: "png"},
			OwnerID:    administrator.ID,
			Type:       5,
			Name:       "General News",
			Recipients: []string{adminID},
		},
		{
			ID:         managersChatID,
			Icon:       Icon{ID: primitive.NewObjectID(), Icon: icons["managers.png"], Type: "image/png", Ext: "png"},
			OwnerID:    administrator.ID,
			Type:       2,
			Name:       "Managers",
			Recipients: []string{adminID},
		},
		{
			ID:         developersChatID,
			Icon:       Icon{ID: primitive.NewObjectID(), Icon: icons["developers.jpg"], Type: "image/jpeg", Ext: "jpg"},
			OwnerID:    administrator.ID,
			Type:       2,
			Name:       "Developers",
			Recipients: []string{adminID},
		},
		{
			ID:         projectByOrderChatID,
			Icon:       Icon{ID: primitive.NewObjectID(), Icon: icons["project_by_order.png"], Type: "image/png", Ext: "png"},
			OwnerID:    administrator.ID,
			Type:       2,
			Name:       "Project by order",
			Recipients: []string{},
		},
		// Добавляем новый канал GitLab Events
		{
			ID:         gitlabEventsChatID,
			Icon:       Icon{ID: primitive.NewObjectID(), Icon: icons["gitlab.png"], Type: "image/png", Ext: "png"},
			OwnerID:    administrator.ID,
			Type:       4,
			Name:       "GitLab Events",
			Recipients: []string{},
		},
	}

	// Вставка каналов в коллекцию
	var channelInterfaces []interface{}
	for _, c := range channels {
		channelInterfaces = append(channelInterfaces, c)
	}
	_, err = channelsCollection.InsertMany(ctx, channelInterfaces)
	if err != nil {
		log.Fatal(err)
	}

	// Создание ролей
	roles := []Role{
		{
			ID:       generalRoleID,
			Name:     "general",
			Channels: []primitive.ObjectID{generalChatID, generalNewsID},
		},
		{
			ID:       guestRoleID,
			Name:     "guest",
			Channels: []primitive.ObjectID{projectByOrderChatID},
		},
		{
			ID:       managersRoleID,
			Name:     "managers",
			Channels: []primitive.ObjectID{managersChatID},
		},
		{
			ID:       developersRoleID,
			Name:     "developers",
			Channels: []primitive.ObjectID{developersChatID, projectByOrderChatID},
		},
	}

	// Вставка ролей в коллекцию
	var roleInterfaces []interface{}
	for _, r := range roles {
		roleInterfaces = append(roleInterfaces, r)
	}
	_, err = rolesCollection.InsertMany(ctx, roleInterfaces)
	if err != nil {
		log.Fatal(err)
	}

	// Добавление пользователя gitlab
	_, err = usersCollection.InsertOne(ctx, bson.M{
		"_id":      primitive.NewObjectID(),
		"email":    "gitlab",
		"username": "gitlab",
		"avatar": bson.M{
			"_id":    primitive.NewObjectID(),
			"avatar": icons["gitlab.png"],
			"type":   "image/png",
			"ext":    "png",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Documents inserted successfully")
}

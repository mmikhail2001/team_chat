package database

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (db *Database) CreateChannel(name string, icon string, recipient_id string, user *User, isNews bool) (*Channel, int) {
	channels := db.Mongo.Collection("channels")

	// определять тип чата по пустой строке recipient_id... совсем не читаемо
	if recipient_id != "" {
		recipient, statusCode := db.GetUser(recipient_id)
		if statusCode != http.StatusOK {
			return nil, statusCode
		}

		var channel Channel
		err := channels.FindOne(context.TODO(), bson.M{
			"type": 1,
			"$or": []bson.M{
				{"recipients": []string{user.ID, recipient.ID}},
				{"recipients": []string{recipient.ID, user.ID}},
			},
		}).Decode(&channel)
		if err == nil {
			return &channel, http.StatusOK
		}

		channel = Channel{
			ID:         primitive.NewObjectID(),
			Type:       1,
			Recipients: []string{user.ID, recipient.ID},
			CreatedAt:  time.Now().Unix(),
			UpdatedAt:  time.Now().Unix(),
		}

		_, err = channels.InsertOne(context.TODO(), channel)
		if err != nil {
			log.Println("CreateChannel: InsertOne: err: ", err)
			return nil, http.StatusInternalServerError
		}

		return &channel, http.StatusOK
	} else {
		var channelType int = 2
		if isNews {
			channelType = 5
		}
		channel := Channel{
			ID:      primitive.NewObjectID(),
			Type:    channelType,
			Name:    name,
			OwnerID: user.ID,
			Recipients: []string{
				user.ID,
			},
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}

		if icon != "" {
			// TODO: копипаста
			file_type_regx := regexp.MustCompile("image/(png|jpeg|gif)")
			file_ext_regx := regexp.MustCompile("png|jpeg|gif")

			file_type := file_type_regx.FindString(icon)
			if file_type == "" {
				return nil, http.StatusBadRequest
			}

			file_ext := file_ext_regx.FindString(file_type)
			iconBS64 := icon[strings.Index(icon, ",")+1:]

			newIcon := Icon{
				ID:   primitive.NewObjectID(),
				Type: file_type,
				Ext:  file_ext,
				Icon: iconBS64,
			}

			channel.Icon = newIcon
		}

		_, err := channels.InsertOne(context.TODO(), channel)
		if err != nil {
			log.Println("CreateChannel: InsertOne: err: ", err)
			return nil, http.StatusInternalServerError
		}

		return &channel, http.StatusOK
	}
}

func (db *Database) ModifyChannel(id string, name string, icon string, user *User) (*Channel, int) {
	channels := db.Mongo.Collection("channels")

	channel, statusCode := db.GetChannel(id, user)
	if statusCode != http.StatusOK {
		return nil, statusCode
	}

	if channel.Type == 1 || channel.OwnerID != user.ID {
		// TODO: может фронту и не предоставлять возможность редачить канал не создателям?...............
		// кажется, это звучит логично
		return nil, http.StatusForbidden
	}

	if name != "" {
		channel.Name = name
	}

	if icon != "" {
		// jpg ???
		file_type_regx := regexp.MustCompile("image/(png|jpeg|gif)")
		file_ext_regx := regexp.MustCompile("png|jpeg|gif")

		file_type := file_type_regx.FindString(icon)
		if file_type == "" {
			return nil, http.StatusBadRequest
		}

		file_ext := file_ext_regx.FindString(file_type)

		iconBS64 := icon[strings.Index(icon, ",")+1:]

		newIcon := Icon{
			ID:   primitive.NewObjectID(),
			Type: file_type,
			Ext:  file_ext,
			Icon: iconBS64,
		}

		channel.Icon = newIcon
	}

	channel.UpdatedAt = time.Now().Unix()

	_, err := channels.ReplaceOne(context.TODO(), bson.M{"_id": channel.ID}, channel)
	if err != nil {
		log.Println("ModifyChannel: ReplaceOne: err: ", err)
		return nil, http.StatusInternalServerError
	}

	// а где рассылка всем участникам? или обновление картинки чата не публикуется в вебсокете?

	return channel, http.StatusOK
}

// Выход из канала пользователя
// а не удаление канала......
func (db *Database) DeleteChannel(id string, user *User) (*Channel, int) {
	channelsCollection := db.Mongo.Collection("channels")

	channel, statusCode := db.GetChannel(id, user)
	if statusCode != http.StatusOK {
		return nil, statusCode
	}

	if channel.Type == 1 {
		return nil, http.StatusForbidden
	}

	rd := options.After
	result := channelsCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": channel.ID}, bson.M{"$pull": bson.M{"recipients": user.ID}}, &options.FindOneAndUpdateOptions{ReturnDocument: &rd})
	if result.Err() != nil {
		log.Println("DeleteChannel: FindOneAndUpdate: err: ", result.Err())
		return nil, http.StatusInternalServerError
	}

	result.Decode(&channel)
	return channel, http.StatusOK
}

// TODO: GetChannelWithUser ?
func (db *Database) GetChannel(id string, user *User) (*Channel, int) {
	channelsCollection := db.Mongo.Collection("channels")
	object_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, http.StatusBadRequest
	}

	var channel Channel
	err = channelsCollection.FindOne(context.TODO(), bson.M{"_id": object_id, "recipients": user.ID}).Decode(&channel)
	if err != nil {
		log.Println("GetChannel: FindOne: err: ", err)
		return nil, http.StatusNotFound
	}

	return &channel, http.StatusOK
}

// TODO: GetChannel ?
func (db *Database) GetChannelWithoutUser(id string) (*Channel, int) {
	channelsCollection := db.Mongo.Collection("channels")
	object_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, http.StatusBadRequest
	}

	var channel Channel
	err = channelsCollection.FindOne(context.TODO(), bson.M{"_id": object_id}).Decode(&channel)
	if err != nil {
		log.Println("GetChannelWithoutUser: FindOne: err: ", err)
		return nil, http.StatusNotFound
	}

	return &channel, http.StatusOK
}

// GetChannelsByUser
func (db *Database) GetChannels(user *User) []Channel {
	channelsCollection := db.Mongo.Collection("channels")

	var channels []Channel
	cursor, err := channelsCollection.Find(context.TODO(), bson.M{"recipients": user.ID})
	if err != nil {
		log.Println("GetChannels: Find: err: ", err)
		return []Channel{}
	}

	for cursor.Next(context.TODO()) {
		var channel Channel
		cursor.Decode(&channel)
		channels = append(channels, channel)
	}

	return channels
}

func (db *Database) GetChannelsMailings() []Channel {
	channelsCollection := db.Mongo.Collection("channels")

	var channels []Channel
	filter := bson.M{"type": bson.M{"$in": []int{4, 5}}}
	cursor, err := channelsCollection.Find(context.TODO(), filter)
	if err != nil {
		log.Println("GetChannelsMailings: Find: err: ", err)
		return []Channel{}
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var channel Channel
		if err := cursor.Decode(&channel); err != nil {
			log.Println("GetChannelsMailings: Decode: err: ", err)
			continue
		}
		channels = append(channels, channel)
	}
	if err := cursor.Err(); err != nil {
		log.Println("GetChannelsMailings: Cursor error: ", err)
		return []Channel{}
	}

	return channels
}

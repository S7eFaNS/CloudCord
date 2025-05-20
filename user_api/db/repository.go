package db

import (
	"cloudcord/user/models"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository struct {
	Collection *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		Collection: db.Collection("users"),
	}
}

func (r *Repository) CreateUser(ctx context.Context, user *models.User) error {
	user.UserID = primitive.NewObjectID()
	_, err := r.Collection.InsertOne(ctx, user)
	return err
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Invalid user ID format: %v", err)
		return nil, err
	}

	var user models.User
	err = r.Collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		log.Printf("Error getting user by ID: %v", err)
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByAuth0ID(ctx context.Context, auth0ID string) (*models.User, error) {
	var user models.User
	err := r.Collection.FindOne(ctx, bson.M{"auth0_id": auth0ID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	cursor, err := r.Collection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Error retrieving users: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		log.Printf("Error decoding users: %v", err)
		return nil, err
	}

	return users, nil
}

package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Room struct {
    ID           primitive.ObjectID   `bson:"_id"`
    RoomNo       string               `json:"roomNo" bson:"roomNo"`
    RoomType     string               `json:"roomType" bson:"roomType"` // Single or Double, add more types as needed
    OccupiedBy   []string             `json:"occupiedBy" bson:"occupiedBy"` // Employee IDs occupying the room
}
package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

//Rule data struct
type Rule struct {
	ID     primitive.ObjectID `json:"-" bson:"_id"`
	Intent string             `json:"intent" bson:"intent"`
	Result string             `json:"result"  bson:"result"`
}

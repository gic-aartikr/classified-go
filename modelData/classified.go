package modelData

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Classified struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title      string             `bson:"title,omitempty" json:"title,omitempty"`
	Address    string             `bson:"address,omitempty" json:"address,omitempty"`
	Latitude   string             `bson:"latitude, omitempty" json:"latitude,omitempty"`
	City       string             `bson:"city,omitempty" json:"city,omitempty"`
	Website    string             `bson:"website, omitempty" json:"website,omitempty"`
	ContactcNo string             `bson:"contactc_no,omitempty" json:"contactc_no,omitempty"`
	User       string             `bson:"user, omitempty" json:"user,omitempty"`
	CategoryId primitive.ObjectID `bson:"category_id, omitempty" json:"category_id,omitempty"`
}

type Search struct {
	City         string `bson:"city,omitempty" json:"city,omitempty"`
	CategoryName string `bson:"category_name,omitempty" json:"category_name,omitempty"`
}

type Category struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	CategoryName string             `bson:"category_name,omitempty" json:"category_name,omitempty"`
	Status       string             `bson:"status,omitempty" json:"status,omitempty"`
}

// type SearchData struct {
// 	Key   string `bson:"key,omitempty" json:"key,omitempty"`
// 	Value string `bson:"value,omitempty" json:"value,omitempty"`
// }

// func (c Classified) String() string {
// 	return fmt.Sprintf("%s age %d says %q", c.Address, c.City, c.ContactcNo, c.Latitude, c.Title, c.User, c.Website, c.CategoryId, c.ID)
// }

package dao

import (
	"classified/model"
	"context"
	"strconv"

	"errors"
	"fmt"
	"log"
	"time"

	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ClassifiedDAO struct {
	Server      string
	Database    string
	Collection  string
	Collection2 string
}

var Collection *mongo.Collection
var CollectionCategory *mongo.Collection
var ctx = context.TODO()
var insertDocs int

func (c *ClassifiedDAO) Connect() {
	clientOptions := options.Client().ApplyURI(c.Server)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	Collection = client.Database(c.Database).Collection(c.Collection)
	CollectionCategory = client.Database(c.Database).Collection(c.Collection2)
}

func (c *ClassifiedDAO) Insert(classified model.Classified) error {

	_, err := Collection.InsertOne(ctx, classified)

	if err != nil {
		return errors.New("Unable to create new record")
	}

	return nil
}

func (e *ClassifiedDAO) InsertData(classified []model.Classified, field string) (int, error) {

	fmt.Println("field:", field)
	data, err := e.SearchDataInCategories(field)
	if err != nil {
		return 0, err
	}
	id := data[0].ID

	for i := range classified {
		classified[i].CategoryId = id
		_, err := Collection.InsertOne(ctx, classified[i])

		if err != nil {
			return 0, errors.New("Unable To Insert New Record")
		}
		insertDocs = i + 1
		fmt.Println("insertDocs:", insertDocs)
	}
	return insertDocs, nil
}

func (e *ClassifiedDAO) SearchByCityAndCategory(cla model.Search) (*excelize.File, string, error) {
	var recordData []*model.Classified

	var cursor *mongo.Cursor
	var err error
	var str = ""
	var cityRecord = cla.City
	var categoryRecord = cla.CategoryName
	var excelData *excelize.File
	// os.MkdirAll("data/download", os.ModePerm)
	// dir := "data/download/"
	file := "searchResult" + fmt.Sprintf("%v", time.Now().Format("3_4_5_pm"))

	if (cityRecord != "") && (categoryRecord != "") {

		data, err := e.SearchDataInCategories(categoryRecord)
		if err != nil {
			return excelData, file, err
		}

		id := data[0].ID
		cursor, err = Collection.Find(ctx, bson.D{primitive.E{Key: "category_id", Value: id}, {Key: "city", Value: cityRecord}})

		if err != nil {
			return excelData, file, err
		}

		str = "No data present in db for given city name and category name "
	} else if (cla.CategoryName) != "" {

		categoryName := cla.CategoryName
		data, err := e.SearchDataInCategories(categoryName)
		if err != nil {
			return excelData, file, err
		}

		id := data[0].ID
		cursor, err = Collection.Find(ctx, bson.D{primitive.E{Key: "category_id", Value: id}})

		if err != nil {
			return excelData, file, err
		}
		str = "No data present in db for given category name"
	} else {
		fmt.Println("city:", cla.City)
		cursor, err = Collection.Find(ctx, bson.D{primitive.E{Key: "city", Value: cla.City}})

		if err != nil {
			return excelData, file, err
		}
		str = "No  data present in db for given city name"
	}

	for cursor.Next(ctx) {
		var e model.Classified
		err := cursor.Decode(&e)
		if err != nil {
			return excelData, file, err
		}
		recordData = append(recordData, &e)
	}

	if recordData == nil {
		return excelData, file, errors.New(str)
	}

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("recordData:", recordData)
	excelData = excelize.NewFile()
	excelData.SetCellValue("Sheet1", "A1", "ID")
	excelData.SetCellValue("Sheet1", "B1", "Title")
	excelData.SetCellValue("Sheet1", "C1", "Address")
	excelData.SetCellValue("Sheet1", "D1", "Latitude")
	excelData.SetCellValue("Sheet1", "E1", "Website")
	excelData.SetCellValue("Sheet1", "F1", "ContactcNo")
	excelData.SetCellValue("Sheet1", "G1", "User")
	excelData.SetCellValue("Sheet1", "H1", "CategoryId")

	for i := 0; i < len(recordData); i++ {
		excelData.SetCellValue("Sheet1", "A"+strconv.Itoa(i+2), recordData[i].ID)
		excelData.SetCellValue("Sheet1", "B"+strconv.Itoa(i+2), recordData[i].Title)
		excelData.SetCellValue("Sheet1", "C"+strconv.Itoa(i+2), recordData[i].Address)
		excelData.SetCellValue("Sheet1", "D"+strconv.Itoa(i+2), recordData[i].Latitude)
		excelData.SetCellValue("Sheet1", "E"+strconv.Itoa(i+2), recordData[i].Website)
		excelData.SetCellValue("Sheet1", "F"+strconv.Itoa(i+2), recordData[i].ContactcNo)
		excelData.SetCellValue("Sheet1", "G"+strconv.Itoa(i+2), recordData[i].User)
		excelData.SetCellValue("Sheet1", "H"+strconv.Itoa(i+2), recordData[i].CategoryId)
	}

	return excelData, file, err
}

func (e *ClassifiedDAO) DeleteRecord(id string) error {

	fmt.Println("id:", id)
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.D{primitive.E{Key: "_id", Value: docID}}

	res, err := Collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return errors.New("no record were deleted")
	}

	return nil
}

func (c *ClassifiedDAO) UpdateRecord(id string, cl model.Classified) error {

	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// title := cl.Title
	// city := cl.City
	// address := cl.Address
	// website := cl.Website
	fmt.Println("docID:", docID)
	// primitive.ObjectIDFromHex(id)
	filter := bson.D{primitive.E{Key: "_id", Value: docID}}

	update := bson.D{primitive.E{Key: "$set", Value: cl}}

	// filter := bson.M{"_id": bson.M{"$eq": docID}}

	// update := bson.M{
	// 	"$set": bson.M{
	// 		"title":   title,
	// 		"address": address,
	// 		"website": website,
	// 		"city":    city,
	// 	},
	// }

	e := &model.Classified{}
	return Collection.FindOneAndUpdate(ctx, filter, update).Decode(e)

}

// func InsertCategory(category model.Category) error {

// 	_, err := Collection.InsertOne(ctx, category)

// 	if err != nil {
// 		return errors.New("Unable to create new record")
// 	}

// 	return nil
// }

// /////////category record////////////////////////////
func (c *ClassifiedDAO) InsertRecord(category model.Category) error {

	_, err := CollectionCategory.InsertOne(ctx, category)

	if err != nil {
		return errors.New("Unable to create new record")
	}

	return nil
}

func (e *ClassifiedDAO) SearchDataInCategories(name string) ([]*model.Category, error) {
	var data []*model.Category

	cursor, err := CollectionCategory.Find(ctx, bson.D{primitive.E{Key: "category_name", Value: name}})
	fmt.Println("cursor:", cursor)
	if err != nil {
		return data, err
	}

	for cursor.Next(ctx) {
		var e model.Category
		err := cursor.Decode(&e)
		if err != nil {
			return data, err
		}
		data = append(data, &e)
	}

	if data == nil {
		return data, errors.New("No data present in db for given category")
	}
	return data, nil
}

func (e *ClassifiedDAO) SearchUsingBothTables(field string) ([]*model.Classified, error) {
	var finalData []*model.Classified

	data, err := e.SearchDataInCategories(field)
	if err != nil {
		return finalData, err
	}

	id := data[0].ID
	cursor, err := Collection.Find(ctx, bson.D{primitive.E{Key: "category_id", Value: id}})

	if err != nil {
		return finalData, err
	}

	for cursor.Next(ctx) {
		var e model.Classified
		err := cursor.Decode(&e)
		if err != nil {
			return finalData, err
		}
		finalData = append(finalData, &e)
	}

	if finalData == nil {
		return finalData, errors.New("No data present in city data db for given category")
	}
	return finalData, nil

}

func (e *ClassifiedDAO) DeleteDataInCategories(categoryId string) (string, error) {

	id, err := primitive.ObjectIDFromHex(categoryId)

	if err != nil {
		return "", err
	}

	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	cur, err := CollectionCategory.DeleteOne(ctx, filter)

	if err != nil {
		return "", err
	}

	if cur.DeletedCount == 0 {
		return "", errors.New("Unable To Delete Data")
	}

	return "Deleted Successfully", nil
}

func (e *ClassifiedDAO) UpdateDataInCategories(categoryData model.Category, field string) (string, error) {

	id, err := primitive.ObjectIDFromHex(field)

	if err != nil {
		return "", err
	}

	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	fmt.Println("filter:", filter)
	update := bson.D{primitive.E{Key: "$set", Value: categoryData}}

	err2 := CollectionCategory.FindOneAndUpdate(ctx, filter, update).Decode(e)

	if err2 != nil {
		return "", err2
	}
	return "Data Updated Successfully", nil
}

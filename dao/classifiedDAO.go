package dao

import (
	"bytes"
	"classified/modelData"
	"context"
	"os"
	"strconv"

	"errors"
	"fmt"
	"log"
	"time"

	"github.com/signintech/gopdf"
	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/creator"
	"github.com/unidoc/unipdf/v3/model"
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

func (c *ClassifiedDAO) Insert(classified modelData.Classified) error {

	_, err := Collection.InsertOne(ctx, classified)

	if err != nil {
		return errors.New("Unable to create new record")
	}

	return nil
}

func (e *ClassifiedDAO) InsertData(classified []modelData.Classified, field string) (int, error) {

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

func (e *ClassifiedDAO) SearchByCityAndCategory(cla modelData.Search) (*excelize.File, string, error) {
	var recordData []*modelData.Classified

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
		var e modelData.Classified
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

func (e *ClassifiedDAO) ConvertDatatoPDF(cla modelData.Search) ([]byte, string, error) {
	var recordData []*modelData.Classified

	var cursor *mongo.Cursor
	var err error
	var str = ""
	var cityRecord = cla.City
	var categoryRecord = cla.CategoryName
	var pdfData []byte
	os.MkdirAll("data/download", os.ModePerm)
	dir := "data/download/"
	file := "searchData" + fmt.Sprintf("%v", time.Now().Format("3_4_5_pm"))

	if (cityRecord != "") && (categoryRecord != "") {

		data, err := e.SearchDataInCategories(categoryRecord)
		if err != nil {
			return pdfData, file, err
		}

		id := data[0].ID
		cursor, err = Collection.Find(ctx, bson.D{primitive.E{Key: "category_id", Value: id}, {Key: "city", Value: cityRecord}})

		if err != nil {
			return pdfData, file, err
		}

		str = "No data present in db for given city name and category name "
	} else if (cla.CategoryName) != "" {

		categoryName := cla.CategoryName
		data, err := e.SearchDataInCategories(categoryName)
		if err != nil {
			return pdfData, file, err
		}

		id := data[0].ID
		cursor, err = Collection.Find(ctx, bson.D{primitive.E{Key: "category_id", Value: id}})

		if err != nil {
			return pdfData, file, err
		}
		str = "No data present in db for given category name"
	} else {

		cursor, err = Collection.Find(ctx, bson.D{primitive.E{Key: "city", Value: cla.City}})

		if err != nil {
			return pdfData, file, err
		}
		str = "No  data present in db for given city name"
	}

	for cursor.Next(ctx) {
		var e modelData.Classified
		err := cursor.Decode(&e)
		if err != nil {
			return pdfData, file, err
		}
		recordData = append(recordData, &e)
	}

	if recordData == nil {
		return pdfData, file, errors.New(str)
	}

	if err != nil {
		fmt.Println(err)
	}

	_, err = writeToPdf(dir, file, recordData)

	if err != nil {
		return pdfData, file, err
	}
	// }
	return pdfData, file, err

}

func writeDataIntoPdf(dir, file string, data []*modelData.Classified) error {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	err := pdf.AddTTFFont("wts11", "./font/Lato-Black.ttf")
	if err != nil {
		log.Print(err.Error())
		fmt.Println(err)
		return err
	}

	err = pdf.SetFont("wts11", "", 10)
	if err != nil {
		log.Print(err.Error())
		return err
	}
	pdf.SetXY(50, 50)
	x := 10.0
	y := 10.0

	for i := range data {
		pdf.SetXY(50, 50+y)
		pdf.Cell(nil, fmt.Sprintf("%v", data[i].ID))
		pdf.Cell(nil, data[i].Title)
		pdf.Cell(nil, data[i].Address)
		pdf.Cell(nil, data[i].Address)
		pdf.Cell(nil, fmt.Sprintf("%v", data[i].Latitude))
		pdf.Cell(nil, data[i].Website)
		pdf.Cell(nil, fmt.Sprintf("%v", data[i].ContactcNo))
		pdf.Cell(nil, data[i].User)
		pdf.Cell(nil, data[i].City)
		pdf.SplitText("", 10.0)
		pdf.Cell(nil, fmt.Sprintf("%v", data[i].CategoryId))
		x = x + 50.0
		y = y + 50.0
	}

	pdf.WritePdf(dir + file + ".pdf")
	fmt.Printf("Completed")
	return nil
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

func (c *ClassifiedDAO) UpdateRecord(id string, cl modelData.Classified) error {

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

	e := &modelData.Classified{}
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
func (c *ClassifiedDAO) InsertRecord(category modelData.Category) error {

	_, err := CollectionCategory.InsertOne(ctx, category)

	if err != nil {
		return errors.New("Unable to create new record")
	}

	return nil
}

func (e *ClassifiedDAO) SearchDataInCategories(name string) ([]*modelData.Category, error) {
	var data []*modelData.Category

	cursor, err := CollectionCategory.Find(ctx, bson.D{primitive.E{Key: "category_name", Value: name}})
	fmt.Println("cursor:", cursor)
	if err != nil {
		return data, err
	}

	for cursor.Next(ctx) {
		var e modelData.Category
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

func (e *ClassifiedDAO) SearchUsingBothTables(field string) ([]*modelData.Classified, error) {
	var finalData []*modelData.Classified

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
		var e modelData.Classified
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

func (e *ClassifiedDAO) UpdateDataInCategories(categoryData modelData.Category, field string) (string, error) {

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

func writeToPdf(dir, file string, data []*modelData.Classified) (*creator.Creator, error) {
	c := creator.New()
	err := license.SetMeteredKey("72c4ab06d023bbc8b2e186d089f9e052654afea32b75141f39c7dc1ab3b108ca")

	if err != nil {

		fmt.Println(err)
		return c, err
	}

	c.SetPageMargins(50, 50, 50, 50)

	// Create report fonts.
	// UniPDF supports a number of font-families, which can be accessed using model.
	// Here we are creating two fonts, a normal one and its bold version
	font, err := model.NewStandard14Font(model.HelveticaName)
	if err != nil {
		log.Fatal(err)
		return c, err
	}

	// Bold font
	fontBold, err := model.NewStandard14Font(model.HelveticaBoldName)
	if err != nil {
		log.Fatal(err)
		return c, err
	}

	// Generate basic usage chapter.
	if err := basicUsage(c, font, fontBold, data); err != nil {
		log.Fatal(err)
		return c, err
	}

	var buf bytes.Buffer
	c.Write(&buf)
	c.WriteToFile(dir + file + "report.pdf")
	return c, nil
}

func basicUsage(c *creator.Creator, font, fontBold *model.PdfFont, data []*modelData.Classified) error {
	// Create chapter.
	ch := c.NewChapter("Search Data")
	ch.SetMargins(0, 0, 50, 0)
	ch.GetHeading().SetFont(font)
	ch.GetHeading().SetFontSize(18)
	ch.GetHeading().SetColor(creator.ColorRGBFrom8bit(72, 86, 95))
	// You can also set inbuilt colors using creator
	// ch.GetHeading().SetColor(creator.ColorBlack)

	// Draw subchapters. Here we are only create horizontally aligned chapter.
	// You can also vertically align and perform other optimizations as well.
	// Check GitHub example for more.
	contentAlignH(c, ch, font, fontBold, data)

	// Draw chapter.
	if err := c.Draw(ch); err != nil {
		return err
	}

	return nil
}

func contentAlignH(c *creator.Creator, ch *creator.Chapter, font, fontBold *model.PdfFont, data []*modelData.Classified) {
	// Create subchapter.
	// sc := ch.NewSubchapter("Content horizontal alignment")
	// sc.GetHeading().SetFontSize(13)
	// sc.GetHeading().SetColor(creator.ColorBlue)

	// // Create subchapter description.
	// desc := c.NewStyledParagraph()
	// desc.Append("Cell content can be aligned horizontally left, right or it can be centered.")

	// sc.Add(desc)

	// Create table.
	table := c.NewTable(9)
	table.SetMargins(0, 0, 10, 0)

	drawCell := func(text string, font *model.PdfFont, align creator.CellHorizontalAlignment) {
		p := c.NewStyledParagraph()
		p.Append(text).Style.Font = font

		cell := table.NewCell()
		cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
		cell.SetHorizontalAlignment(align)
		cell.SetContent(p)
	}

	// Draw table header.
	drawCell("ID", fontBold, creator.CellHorizontalAlignmentLeft)
	drawCell("Address", fontBold, creator.CellHorizontalAlignmentCenter)
	drawCell("City", fontBold, creator.CellHorizontalAlignmentCenter)
	drawCell("Title", fontBold, creator.CellHorizontalAlignmentCenter)
	drawCell("Latitude", fontBold, creator.CellHorizontalAlignmentCenter)
	drawCell("Website", fontBold, creator.CellHorizontalAlignmentCenter)
	drawCell("ContactcNo", fontBold, creator.CellHorizontalAlignmentCenter)
	drawCell("User", fontBold, creator.CellHorizontalAlignmentCenter)
	drawCell("CategoryId", fontBold, creator.CellHorizontalAlignmentCenter)

	// Draw table content.
	for i := 0; i < len(data); i++ {

		drawCell(fmt.Sprintf("%v", data[i].ID), font, creator.CellHorizontalAlignmentLeft)
		drawCell(data[i].Address, font, creator.CellHorizontalAlignmentCenter)
		drawCell(data[i].City, font, creator.CellHorizontalAlignmentRight)
		drawCell(data[i].Title, font, creator.CellHorizontalAlignmentLeft)
		drawCell(data[i].Latitude, font, creator.CellHorizontalAlignmentCenter)
		drawCell(data[i].Website, font, creator.CellHorizontalAlignmentRight)
		drawCell(data[i].ContactcNo, font, creator.CellHorizontalAlignmentLeft)
		drawCell(data[i].User, font, creator.CellHorizontalAlignmentCenter)
		drawCell(data[i].CategoryId.Hex(), font, creator.CellHorizontalAlignmentRight)
	}

	ch.Add(table)
}

package main

import (
	"bytes"
	"classified/dao"
	"classified/modelData"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const tokenIdForAdmin = "tokenAdmin123456783"
const tokenIdForPosters = "tokenPosters562348"

var buf []bytes.Buffer
var cla = dao.ClassifiedDAO{}

// var cat = categoryService.CategoryDAO{}

func init() {
	cla.Server = "mongodb://localhost:27017"
	//  cla.Server = "mongodb+srv://m001-student:m001-mongodb-basics@sandbox.7zffz3a.mongodb.net/?retryWrites=true&w=majority"
	cla.Database = "classifiedData"
	cla.Collection = "classified"
	cla.Collection2 = "category"

	cla.Connect()
}

func addClassifiedData(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	if r.Method != "POST" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
	}

	var classified modelData.Classified

	if err := json.NewDecoder(r.Body).Decode(&classified); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
	}

	if err := cla.Insert(classified); err != nil {

		respondWithError(w, http.StatusBadRequest, "Invalid request")
	} else {
		respondWithJson(w, http.StatusAccepted, map[string]string{
			"message": "Record inserted successfully",
		})
	}
}

func addCategoryRecord(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	if r.Method != "POST" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
	}

	var classified modelData.Classified

	if err := json.NewDecoder(r.Body).Decode(&classified); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
	}

	if err := cla.Insert(classified); err != nil {

		respondWithError(w, http.StatusBadRequest, "Invalid request")
	} else {
		respondWithJson(w, http.StatusAccepted, map[string]string{
			"message": "Record inserted successfully",
		})
	}
}

func searchClassifiedData(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	if r.Method != "POST" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
		return
	}

	var cl modelData.Search

	if err := json.NewDecoder(r.Body).Decode(&cl); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	// if cl.City == "" {
	// 	respondWithError(w, http.StatusBadRequest, "Please provide city for search")
	// 	return
	// }

	fmt.Println(cl)
	if searchdocs, fileName, err := cla.SearchByCityAndCategory(cl); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename="+fileName+".xlsx")
		w.Header().Set("File-Name", "searchdocs.xlsx")
		w.Header().Set("Content-Transfer-Encoding", "binary")
		w.Header().Set("Expires", "0")
		err := searchdocs.Write(w)
		fmt.Println(err)

	}

}

func convertClassifiedDataToPDF(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	if r.Method != "POST" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
		return
	}

	var cl modelData.Search

	if err := json.NewDecoder(r.Body).Decode(&cl); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	// if cl.City == "" {
	// 	respondWithError(w, http.StatusBadRequest, "Please provide city for search")
	// 	return
	// }

	if pdfData, fileName, err := cla.ConvertDatatoPDF(cl); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename="+fileName+".pdf")
		w.Header().Set("Content-Transfer-Encoding", "binary")
		http.ServeContent(w, r, "Workbook.xlxs", time.Now(), bytes.NewReader(pdfData))
	}

}

func addRecord(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	token := r.Header.Get("tokenid")

	admin := token == tokenIdForAdmin
	poster := token == tokenIdForPosters

	if !(admin || poster) {
		respondWithError(w, http.StatusBadRequest, "Unauthorized")
		return
	}

	if r.Method != "POST" {

		respondWithError(w, http.StatusBadRequest, "Invalid Method")
		return
	}

	path := r.URL.Path
	segments := strings.Split(path, "/")
	field := segments[len(segments)-1]
	fmt.Println("fieldRecord:", field)
	var classified []modelData.Classified

	if err := json.NewDecoder(r.Body).Decode(&classified); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	if inserted, err := cla.InsertData(classified, field); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		respondWithJson(w, http.StatusAccepted, map[string]string{
			"message": strconv.Itoa(inserted) + " Record Inserted Successfully",
		})
	}
}

func deleteRecord(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	token := r.Header.Get("tokenid")

	id := strings.Split(r.URL.Path, "/")[2]

	if token != tokenIdForAdmin {
		respondWithError(w, http.StatusBadRequest, "Unauthorized")
		return
	}

	if r.Method != "DELETE" {
		respondWithError(w, http.StatusBadRequest, "Invalid Method")
		return
	}

	fmt.Println("idRecord:", id)

	if err := cla.DeleteRecord(id); err != nil {

		respondWithError(w, http.StatusBadRequest, err.Error())
	} else {
		respondWithJson(w, http.StatusAccepted, map[string]string{
			"message": "Record deleted successfully",
		})
	}
}

func updateRecord(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	token := r.Header.Get("tokenid")

	id := strings.Split(r.URL.Path, "/")[2]

	if token != tokenIdForAdmin {
		respondWithError(w, http.StatusBadRequest, "Unauthorized")
		return
	}

	if r.Method != "PUT" {
		respondWithError(w, http.StatusBadRequest, "Invalid Method")
		return
	}

	// fmt.Println("idRecord:", id)
	var cl modelData.Classified

	err := json.NewDecoder(r.Body).Decode(&cl)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
	}

	// / id := cl.ID
	// str := strconv.Itoa(id)
	if err := cla.UpdateRecord(id, cl); err != nil {

		respondWithError(w, http.StatusBadRequest, err.Error())
	} else {
		respondWithJson(w, http.StatusAccepted, map[string]string{
			"message": "Record updated successfully",
		})
	}
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {

	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, map[string]string{"error": msg})
}

/////////////end point for category//////////

func addCategoryData(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	if r.Method != "POST" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
	}

	var category modelData.Category

	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
	}

	if err := cla.InsertRecord(category); err != nil {

		respondWithError(w, http.StatusBadRequest, "Invalid request")
	} else {
		respondWithJson(w, http.StatusAccepted, map[string]string{
			"message": "Record inserted successfully",
		})
	}
}

func searchBothTable(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "GET" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
		return
	}

	var reqBody map[string]string

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
	}

	name := reqBody["category_name"]

	if searchData, err := cla.SearchUsingBothTables(name); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		respondWithJson(w, http.StatusAccepted, searchData)
	}
}

func updateDataInCategory(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	token := r.Header.Get("tokenid")

	admin := token == tokenIdForAdmin

	id := strings.Split(r.URL.Path, "/")[2]

	if !(admin) {
		respondWithError(w, http.StatusBadRequest, "Unauthorized")
		return
	}

	if r.Method != "PUT" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
		return
	}

	var data modelData.Category

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	// path := r.URL.Path
	// segments := strings.Split(path, "/")
	// field := segments[len(segments)-1]
	// fmt.Println("field:", field)
	if updated, err := cla.UpdateDataInCategories(data, id); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		respondWithJson(w, http.StatusAccepted, map[string]string{
			"message": updated,
		})
	}
}

func deleteDataInCategory(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	token := r.Header.Get("tokenid")

	id := strings.Split(r.URL.Path, "/")[2]

	if token != tokenIdForAdmin {
		respondWithError(w, http.StatusBadRequest, "Unauthorized")
		return
	}

	if r.Method != "DELETE" {
		respondWithError(w, http.StatusBadRequest, "Invalid Method")
		return
	}

	if deleted, err := cla.DeleteDataInCategories(id); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		respondWithJson(w, http.StatusAccepted, map[string]string{
			"message": deleted,
		})
	}
}

func searchByCategory(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "GET" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
		return
	}

	var reqBody map[string]string

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
	}

	name := reqBody["category_name"]

	if searchData, err := cla.SearchDataInCategories(name); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		respondWithJson(w, http.StatusAccepted, searchData)
	}
}

func main() {
	http.HandleFunc("/add-classified/", addClassifiedData)
	http.HandleFunc("/search-classified/", searchClassifiedData)
	http.HandleFunc("/add-record/", addRecord)
	http.HandleFunc("/delete-record/", deleteRecord)
	http.HandleFunc("/update-record/", updateRecord)
	http.HandleFunc("/createCategory-record/", addCategoryData)
	http.HandleFunc("/search-both-table/", searchBothTable)
	http.HandleFunc("/delete-data/", deleteDataInCategory)
	http.HandleFunc("/search-by-category", searchByCategory)
	http.HandleFunc("/update-data-category/", updateDataInCategory)
	http.HandleFunc("/convert-pdf-data/", convertClassifiedDataToPDF)
	fmt.Println("Excecuted Main Method")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

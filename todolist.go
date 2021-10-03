package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

// var dsn = "admin:26!73kxferHY*Pp@/tcp(gotest.cynjynn6hacb.us-east-1.rds.amazonaws.com:3306)/todolist?charset=utf8&parseTime=True&loc=Local"
// var db, _ = gorm.Open("mysql", dsn)
// var db, _ := gorm.Open("mysql", "root:root@/todolist?charset=utf8&parseTime=True&loc=Local")
var db *gorm.DB

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)
}

func GetItemByID(Id int) bool {
	todo := &TodoItemModel{}
	result := db.First(&todo, Id)
	if result.Error != nil {
		log.Warn("TodoItem not found in database")
		return false
	}
	return true
}

func GetTodoItems(completed bool) interface{} {
	var todos []TodoItemModel
	TodoItems := db.Where("completed = ?", completed).Find(&todos).Value
	return TodoItems
}

func databaseConnection() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error: " + err.Error())
	}

	authenticationToken, err := auth.BuildAuthToken(
		context.TODO(),
		"pepito-aurora-workshop-4.cluster-cynjynn6hacb.us-east-1.rds.amazonaws.com:3306", // Database Endpoint (With Port)
		"us-east-1", // AWS Region
		"root",      // Database Account
		cfg.Credentials,
	)
	if err != nil {
		panic("failed to create authentication token: " + err.Error())
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?tls=true&allowCleartextPasswords=true",
		"root", authenticationToken, "pepito-aurora-workshop-4.cluster-cynjynn6hacb.us-east-1.rds.amazonaws.com:3306", "todolist")

	gormDb, err := gorm.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	db = gormDb

}

func main() {

	databaseConnection()

	defer db.Close()
	db.Debug().DropTableIfExists(&TodoItemModel{})
	db.Debug().AutoMigrate(&TodoItemModel{})

	log.Info("Starting Todolist API server")
	router := mux.NewRouter()
	router.HandleFunc("/healthz", Health).Methods("GET")
	router.HandleFunc("/todo-completed", GetCompletedItems).Methods("GET")
	router.HandleFunc("/todo-incomplete", GetIncompleteItems).Methods("GET")
	router.HandleFunc("/todo", CreateItem).Methods("POST")
	router.HandleFunc("/todo/{id}", UpdateItem).Methods("POST")
	router.HandleFunc("/todo/{id}", DeleteItem).Methods("DELETE")

	handler := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"},
	}).Handler(router)
	http.ListenAndServe(":8000", handler)
}

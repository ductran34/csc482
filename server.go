package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/gorilla/mux"
)

type Movies struct {
	Movies []Movie `json:"results"`
}
type Movie struct {
	Id               int     `json:"Id" dynamodbav:"Id"`
	Title            string  `json:"title" dynamodbav:"Title"`
	VoteAverage      float64 `json:"vote_average"`
	OriginalLanguage string  `json:"original_language"`
	Overview         string  `json:"overview"`
	Popularity       float64 `json:"popularity"`
	MediaType        string  `json:"media_type"`
}

var movies []Movie
var count *int64

func getDynamodbTable() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		log.Fatal("Got error initializing AWS: %s", err)
	}
	svc := dynamodb.New(sess)

	//using scan api
	params := &dynamodb.ScanInput{
		TableName: aws.String("dtran4-tmdbmovie"),
	}
	result, err := svc.Scan(params)
	if err != nil {
		fmt.Println("failed to make Query API call", err)
	}
	fmt.Println(result.Count)
	count = result.Count

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &movies)
	if err != nil {
		fmt.Println("failed to unmarshal Query result items", err)
	}
	fmt.Println("Successfully get movies")

}

func getStatus(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	getDynamodbTable()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"table":       "dtran4-tmdbmovie",
		"recordCount": count})
}
func getAllData(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	getDynamodbTable()
	json.NewEncoder(w).Encode(movies)
}

func searchData(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	searchMediaType := mux.Vars(req)["mediaType"]
	proper, err := regexp.MatchString(`^[a-zA-Z][^0-9]+$`, searchMediaType)
	if err != nil {
		log.Fatal(err)
	}

	if proper {
		w.WriteHeader(http.StatusOK)

		sess, err := session.NewSession(&aws.Config{
			Region: aws.String("us-east-1")},
		)
		if err != nil {
			log.Fatal("Got error initializing AWS: %s", err)
		}
		svc := dynamodb.New(sess)

		filter := expression.Contains(expression.Name("media_type"), searchMediaType)

		expr, err := expression.NewBuilder().WithFilter(filter).Build()
		if err != nil {
			log.Fatal("Got error building expression: %s", err)
		}
		params := &dynamodb.ScanInput{
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			FilterExpression:          expr.Filter(),
			ProjectionExpression:      expr.Projection(),
			TableName:                 aws.String("dtran4-tmdbmovie"),
		}
		out, err := svc.Scan(params)
		if err != nil {
			log.Fatal("Query API call failed: %s", err)
		}
		searchResponse := []Movie{}
		err = dynamodbattribute.UnmarshalListOfMaps(out.Items, &searchResponse)
		if err != nil {
			panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
		}
		json.NewEncoder(w).Encode(searchResponse)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := "Search endpoint should be properly formed by /search?mediaType=string"
		json.NewEncoder(w).Encode(errorResponse)
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/dtran4/all", getAllData).Methods("GET")
	router.HandleFunc("/dtran4/status", getStatus).Methods("GET")
	router.HandleFunc("/dtran4/search", searchData).Queries("mediaType", "{mediaType:.*}")
	log.Fatal(http.ListenAndServe(":8080", router))
}

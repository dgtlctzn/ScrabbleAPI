package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"
)

type WordInfo struct {
	Valid bool `json:"validWord"`
	Word string `json:"word"`
	Message string `json:"message"`
}

type ScrabbleResponse struct{
	Success bool `json:"success"`
	Data WordInfo `json:"data"`
	Message string `json:"message"`
}

type Tag struct {
	ID         int    `json:"id"`
	Updated    string `json:"updated"`
	Word       string `json:"word"`
	WordLength int    `json:"word_length"`
}

func getAddress() string {
	address := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s", os.Getenv("USER"), os.Getenv("PASS"), os.Getenv("HOST"), os.Getenv("DB"),
	)
	return address
}

func makeQuery(query string, arg string) ([]Tag, error) {
	var tags []Tag

	address := getAddress()
	db, err := sql.Open("mysql", address)
	if err != nil {
		return tags, err
	}
	defer db.Close()

	results, err := db.Query(query, arg)
	if err != nil {
		return tags, err
	}

	for results.Next() {
		var tag Tag
		err = results.Scan(&tag.ID, &tag.Updated, &tag.Word, &tag.WordLength)
		if err != nil {
			return tags, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func isValidWord(word string) bool {
	query := "SELECT * FROM ValidWords WHERE word = ?"
	tags, err := makeQuery(query, word)
	if err != nil {
		panic(err.Error())
	}
	return len(tags) == 1
}

func HandleRequest(
	ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	word, ok := request.QueryStringParameters["word"]
	if !ok {
		invalidScrabbleRes, err := json.Marshal(
			ScrabbleResponse{
				Success: false,
				Message: "Must have 'word' as query param",
			},
		)
		if err != nil {
            panic(err)
        }
		return events.APIGatewayProxyResponse{
			Body: string(invalidScrabbleRes), StatusCode: 400,
		}, nil
	}

	isValid := isValidWord(word)
	var message string
	if isValid {
		message = fmt.Sprintf("'%s' is a valid Scrabble word!", word)
	} else {
		message = fmt.Sprintf("'%s' is not a valid Scrabble word :(", word)
	}
	validScrabbleRes, err := json.Marshal(
		ScrabbleResponse{
			Success: true,
			Data: WordInfo{Valid: isValid, Word: word, Message: message},
			Message: "success",
		},
	)
	if err != nil {
		panic(err)
	}
    return events.APIGatewayProxyResponse{
		Body: string(validScrabbleRes), 
		StatusCode: 200,
	}, nil
}

func main() {
    lambda.Start(HandleRequest)
}
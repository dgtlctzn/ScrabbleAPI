package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"math/rand"
	"strconv"
	"time"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"
)

type RandomWordInfo struct {
	Word string `json:"word"`
	Length int `json:"length"`
}

type ScrabbleResponse struct{
	Success bool `json:"success"`
	Data RandomWordInfo `json:"data"`
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

func getRandomWord(wordLength ...int) (Tag, error) {
	var tag Tag
	var count int = 0

	address := getAddress()
	db, err := sql.Open("mysql", address)
	if err != nil {
		return tag, err
	}
	defer db.Close()

	var results = &sql.Rows{}
	if len(wordLength) > 0 {
		results, err = db.Query("SELECT COUNT(*) FROM ValidWords WHERE word_length = ?", wordLength[0])
		if err != nil {
			return tag, err
		}
	} else {
		results, err = db.Query("SELECT COUNT(*) FROM ValidWords")
		if err != nil {
			return tag, err
		}
	}
	for results.Next() {
		err = results.Scan(&count)
		if err != nil {
			return tag, err
		}
	}

	var wordResults = &sql.Rows{}
	if len(wordLength) > 0 {
		wordResults, err = db.Query(
			"SELECT * FROM (SELECT * FROM ValidWords WHERE word_length = ?) AS dt LIMIT ?,1",  
			wordLength[0],
			getRandomNum(count),
		)
	} else {
		wordResults, err = db.Query(
			"SELECT * FROM ValidWords LIMIT ?,1", getRandomNum(count),
		)
	}
	if err != nil {
		return tag, err
	}
	if wordResults.Next() {
		err = wordResults.Scan(&tag.ID, &tag.Updated, &tag.Word, &tag.WordLength)
		if err != nil {
			return tag, err
		}
		return tag, nil
	}
	return tag, nil
}

func getRandomNum(limit int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(limit)
}

func HandleRequest(
		ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var tag Tag
		var err error
		length, ok := request.QueryStringParameters["length"]
		if !ok {
			tag, err = getRandomWord()
			if err != nil {
				panic(err.Error())
			}
		} else {
			lengthNum, err := strconv.Atoi(length)
			if lengthNum < 0 || lengthNum > 15 {
				invalidScrabbleRes, err := json.Marshal(
					ScrabbleResponse{Success: false, Message: "Length must be between 1 and 15 inclusive"},
				)
				if err != nil {
					panic(err.Error())
				}
				return events.APIGatewayProxyResponse{
					Body: string(invalidScrabbleRes), 
					StatusCode: 400,
				}, nil
			}
			if err != nil {
				panic(err.Error())
			}
			tag, err = getRandomWord(lengthNum)
			if err != nil {
				panic(err.Error())
			}
		}
		if err != nil {
			panic(err)
		}
		validScrabbleRes, err := json.Marshal(
			ScrabbleResponse{
				Success: true,
				Data: RandomWordInfo{Word: tag.Word, Length: len(tag.Word)},
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
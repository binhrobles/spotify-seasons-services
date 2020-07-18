package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type RequestBody struct {
	Code string `json:"code"`
}

// sends auth code to spotify token endpoint
// gets access/refresh tokens back
func getTokens(code string) string {
	response, err := http.PostForm("https://accounts.spotify.com/api/token", url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {os.Getenv("REDIRECT_URI")},
		"client_id":     {os.Getenv("CLIENT_ID")},
		"client_secret": {os.Getenv("CLIENT_SECRET")},
	})

	if err != nil {
		panic(err)
	}

	defer response.Body.Close()
	tokenBody, err := ioutil.ReadAll(response.Body)

	fmt.Println(string(tokenBody))

	return string(tokenBody)
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// unmarshal json request into RequestBody object
	requestBody := RequestBody{}
	if err := json.Unmarshal([]byte(req.Body), &requestBody); err != nil {
		panic(err)
	}

	tokenBody := getTokens(requestBody.Code)

	// parse access/refresh tokens
	// create/update user in DDB

	// return access code to user
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(tokenBody),
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*", // TODO: not this
		},
	}, nil
}

func main() {
	lambda.Start(handler)
}

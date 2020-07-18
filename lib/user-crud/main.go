package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/ssm"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type RequestBody struct {
	Code string `json:"code"`
}

type SpotifyTokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SpotifyUserInfoResponse struct {
	Name       string `json:"display_name"`
	ID         string `json:"id"`
	ProfileUrl string `json:"href"`
	Avatar     string `json:"images[0].url"`
}

type ResponseBody struct {
	AccessToken string `json:"access_token"`
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

var CLIENT_SECRET string
var ddb_client *dynamodb.DynamoDB

// performs one-time initializations for this lambda container
func coldstartInit() {
	// create AWS SDK session for AWS service clients
	sess := session.Must(session.NewSession())

	// grabs spotify client secret from SSM parameter store
	ssm_client := ssm.New(sess)
	ssm_result, err := ssm_client.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String("/spotifySeasons/clientSecret"),
		WithDecryption: aws.Bool(true),
	})
	handleError(err)
	CLIENT_SECRET = aws.StringValue(ssm_result.Parameter.Value)

	// initializes DDB interface
	ddb_client = dynamodb.New(sess)
}

// sends auth code to spotify token endpoint
// gets access/refresh tokens back
func getTokens(code string) SpotifyTokensResponse {
	response, err := http.PostForm("https://accounts.spotify.com/api/token", url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {os.Getenv("REDIRECT_URI")},
		"client_id":     {os.Getenv("CLIENT_ID")},
		"client_secret": {CLIENT_SECRET},
	})
	handleError(err)

	defer response.Body.Close()
	tokenBody, err := ioutil.ReadAll(response.Body)
	handleError(err)

	// parse out access and refresh tokens
	spotifyResponse := SpotifyTokensResponse{}
	err = json.Unmarshal([]byte(tokenBody), &spotifyResponse)
	handleError(err)

	return spotifyResponse
}

func getUserInfo(token string) SpotifyUserInfoResponse {
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me", nil)
	handleError(err)

	req.Header.Add("Authorization", "Bearer "+token)
	client := &http.Client{}
	response, err := client.Do(req)
	handleError(err)

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	handleError(err)

	// parse out relevant user info
	spotifyResponse := SpotifyUserInfoResponse{}
	err = json.Unmarshal([]byte(body), &spotifyResponse)
	handleError(err)

	return spotifyResponse
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if CLIENT_SECRET == "" {
		coldstartInit()
	}

	// unmarshal json request into RequestBody object
	requestBody := RequestBody{}
	err := json.Unmarshal([]byte(req.Body), &requestBody)
	handleError(err)

	tokens := getTokens(requestBody.Code)

	// get user info from spotify
	userInfo := getUserInfo(tokens.AccessToken)

	// create/update user in DDB

	// return info/credentials to client
	response, err := json.Marshal(ResponseBody{
		AccessToken: tokens.AccessToken,
	})

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(response),
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*", // TODO: not this
		},
	}, nil
}

func main() {
	lambda.Start(handler)
}

package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ssm"

	"encoding/json"
	"errors"
	"fmt"
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
	Avatar     string `json:"images[0].url"` // TODO: this doesn't actually work
}

type ErrorResponseBody struct {
	Message string `json:"message"`
}

type ResponseBody struct {
	AccessToken string                  `json:"access_token"`
	UserInfo    SpotifyUserInfoResponse `json:"user"`
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

var CLIENT_SECRET string
var SPOTIFY_ACCOUNTS_BASE_URI string
var SPOTIFY_WEB_BASE_URI string
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

	// sets spotify base URIs for local vs production env
	switch os.Getenv("STAGE") {
	case "local":
		SPOTIFY_ACCOUNTS_BASE_URI = "http://docker.for.mac.localhost:3001"
		SPOTIFY_WEB_BASE_URI = "http://docker.for.mac.localhost:3001"
	case "production":
		SPOTIFY_ACCOUNTS_BASE_URI = "https://accounts.spotify.com/api"
		SPOTIFY_WEB_BASE_URI = "https://api.spotify.com/v1"
	default:
		SPOTIFY_ACCOUNTS_BASE_URI = "https://accounts.spotify.com/api"
		SPOTIFY_WEB_BASE_URI = "https://api.spotify.com/v1"
	}

	fmt.Println("Spotify Accounts URI: " + SPOTIFY_ACCOUNTS_BASE_URI)
	fmt.Println("Spotify Web URI: " + SPOTIFY_WEB_BASE_URI)
	if CLIENT_SECRET == "" {
		panic("Client secret couldn't be retrieved")
	}
}

// sends auth code to spotify token endpoint
// gets access/refresh tokens back
func getTokens(code string) (SpotifyTokensResponse, error) {
	response, err := http.PostForm(SPOTIFY_ACCOUNTS_BASE_URI+"/token", url.Values{
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

	spotifyResponse := SpotifyTokensResponse{}
	if response.StatusCode != 200 {
		fmt.Println(string(tokenBody))
		return spotifyResponse, errors.New("Didn't receive a 200 for the oauth exchange")
	}

	// parse out access and refresh tokens
	err = json.Unmarshal([]byte(tokenBody), &spotifyResponse)
	handleError(err)

	if spotifyResponse.AccessToken == "" || spotifyResponse.RefreshToken == "" {
		return spotifyResponse, errors.New("Spotify didn't respond with a token")
	}

	return spotifyResponse, nil
}

func getUserInfo(token string) (SpotifyUserInfoResponse, error) {
	req, err := http.NewRequest("GET", SPOTIFY_WEB_BASE_URI+"/me", nil)
	handleError(err)

	req.Header.Add("Authorization", "Bearer "+token)
	client := &http.Client{}
	response, err := client.Do(req)
	handleError(err)

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	handleError(err)
	fmt.Println(string(body))

	// parse out relevant user info
	spotifyResponse := SpotifyUserInfoResponse{}
	err = json.Unmarshal([]byte(body), &spotifyResponse)
	handleError(err)

	return spotifyResponse, nil
}

func errorResponse(err error) events.APIGatewayProxyResponse {
	// return info/credentials to client
	response, err := json.Marshal(ErrorResponseBody{
		Message: err.Error(),
	})
	handleError(err)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusBadRequest,
		Body:       string(response),
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*", // TODO: not this
		},
	}
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// unmarshal json request into RequestBody object
	requestBody := RequestBody{}
	err := json.Unmarshal([]byte(req.Body), &requestBody)
	handleError(err)

	tokens, err := getTokens(requestBody.Code)
	if err != nil {
		return errorResponse(err), nil
	}

	// get user info from spotify
	userInfo, err := getUserInfo(tokens.AccessToken)
	if err != nil {
		return errorResponse(err), nil
	}

	// create/update user in DDB

	// return info/credentials to client
	response, err := json.Marshal(ResponseBody{
		AccessToken: tokens.AccessToken,
		UserInfo:    userInfo,
	})
	handleError(err)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(response),
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*", // TODO: not this
		},
	}, nil
}

func main() {
	coldstartInit()
	lambda.Start(handler)
}

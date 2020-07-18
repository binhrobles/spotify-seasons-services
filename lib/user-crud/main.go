package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

var CLIENT_SECRET string

type RequestBody struct {
	Code string `json:"code"`
}

type SpotifyTokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ResponseBody struct {
	AccessToken string `json:"access_token"`
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

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

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if CLIENT_SECRET == "" {
		coldstartInit()
	}

	// unmarshal json request into RequestBody object
	requestBody := RequestBody{}
	err := json.Unmarshal([]byte(req.Body), &requestBody)
	handleError(err)

	tokens := getTokens(requestBody.Code)

	// create/update user in DDB

	// return access code to user
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

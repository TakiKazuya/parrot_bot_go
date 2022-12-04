package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/line/line-bot-sdk-go/linebot"
	"log"
	"net/http"
	"os"
)

type Webhook struct {
	Destination string           `json:"destination"`
	Events      []*linebot.Event `json:"events"`
}

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	bot, err := linebot.New(
		os.Getenv("LINE_CHANNEL_SECRET"),
		os.Getenv("LINE_CHANNEL_TOKEN"),
	)

	if err != nil {
		log.Print(err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf(`{"message":"%s"}`+"n", http.StatusText(http.StatusInternalServerError)),
		}, nil
	}

	log.Print(request.Headers)
	log.Print(request.Body)

	if !validateSignature(os.Getenv("LINE_CHANNEL_SECRET"), request.Headers["x-line-signature"], []byte(request.Body)) {
		log.Print("invalid signature")
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf(`{"message":"%s"}`+"n", linebot.ErrInvalidSignature.Error()),
		}, nil
	}

	var webhook Webhook

	if err := json.Unmarshal([]byte(request.Body), &webhook); err != nil {
		log.Print(err)
		return events.APIGatewayProxyResponse{
			Body: fmt.Sprintf(`{"message":"%s"}`+"n", http.StatusText(http.StatusBadRequest)),
		}, nil
	}

	for _, event := range webhook.Events {
		if event.Type == linebot.EventTypeMessage {
			switch m := event.Message.(type) {
			case *linebot.TextMessage:
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(m.Text)).Do(); err != nil {
					log.Print(err)
					return events.APIGatewayProxyResponse{
						StatusCode: http.StatusInternalServerError,
						Body:       fmt.Sprintf(`{"message":"%s"}`+"\n", http.StatusText(http.StatusBadRequest)),
					}, nil
				}
			}
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
	}, nil
}

func validateSignature(channelSecret string, signature string, body []byte) bool {
	decoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}

	hash := hmac.New(sha256.New, []byte(channelSecret))
	_, err = hash.Write(body)
	if err != nil {
		return false
	}

	return hmac.Equal(decoded, hash.Sum(nil))
}

func main() {
	lambda.Start(HandleRequest)
}

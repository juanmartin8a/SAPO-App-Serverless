package main

import (
	"context"
	"encoding/json"
	"log"
	aiutils "sapopinguino/internal/ai"
	awsutils "sapopinguino/internal/aws"
	"sapopinguino/internal/config"
	dbutils "sapopinguino/internal/db"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/sashabaranov/go-openai"
)

func init() {
    awsutils.ConfigAWS()

	config.ReadConfig(config.ReadConfigOption{})

	aiutils.ConfigOpenAI()

    dbutils.ConfigDB()
}

func handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
    connectionID := event.RequestContext.ConnectionID

    tokens := []*aiutils.Token{}

    tokenStreamChannel := aiutils.ChatCompletion(ctx, openai.GPT4o, aiutils.SystemRoleContent, `{
        "input_language": "English",
        "target_language": "Spanish",
        "input": "abc, easy as do re mi, or as simple as 123, abc 123 baby you and me girl"
    }`)

    for res := range tokenStreamChannel {
		if res.Error != nil {
			log.Printf("\nError encountered: %v\n", res.Error)
            _, err := awsutils.APIGatewayClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
                ConnectionId: &connectionID,
                Data:         []byte("<error:/>"),
            })
            if err != nil {
                log.Println("Error sending error token to client: %s", err)
            }
			break
		}

        log.Println(res.Response.Type)

        tokens = append(tokens, res.Response)

        jsonData, err := json.Marshal(res.Response)
	    if err != nil {
		    log.Println("Error marshaling JSON:", err)
            break
	    }

        _, err = awsutils.APIGatewayClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: &connectionID,
			Data:         jsonData,
		})
        if err != nil {
		    log.Println("Error sending token to client: %s", err)
            break
        }
	}
    _, err := awsutils.APIGatewayClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
        ConnectionId: &connectionID,
        Data:         []byte("<end:)>"),
    })
    if err != nil {
        log.Println("Error sending <end:)> thingy to client: %s", err)
    }

	response := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "\"Hello from Lambda!\"",
	}

	return response, nil
}


func main() {
    lambda.Start(handler)
}

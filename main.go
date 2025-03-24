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
	"github.com/openai/openai-go"
)

func init() {
    awsutils.ConfigAWS()

    config.ReadConfig(config.ReadConfigOption{})

    awsutils.ConfigAWSGateway(&config.C.Websocket.Endpoint)

	aiutils.ConfigOpenAI()

    dbutils.ConfigDB()
}

func handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
    connectionID := event.RequestContext.ConnectionID

    tokens := []*aiutils.Token{}

    bodyBytes := []byte(event.Body)

    var bodyS awsutils.Body

    err := json.Unmarshal(bodyBytes, &bodyS)
    if err != nil {
        log.Printf("Failed to unmarshal request's body: %v", err)
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body:       `"Internal server error :/"`,
        }, err
    }

    tokenStreamChannel := aiutils.ChatCompletion(ctx, openai.ChatModelGPT4o, aiutils.SystemRoleContent, bodyS.Message)

    for res := range tokenStreamChannel {
		if res.Error != nil {
			log.Printf("Error while streaming LLM's response: %v", res.Error)
            _, err = awsutils.APIGatewayClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
                ConnectionId: &connectionID,
                Data:         []byte("<error:/>"),
            })
            if err != nil {
                log.Println("Error sending error token to client: %v", err)
                awsutils.HandleDeleteConnection(ctx, &connectionID, "sending \"<error:/>\" in PostConnection")
            }
			break
		}

        tokens = append(tokens, res.Response)

        jsonData, err := json.Marshal(res.Response)
	    if err != nil {
		    log.Println("Error marshaling JSON:", err)
            _, err := awsutils.APIGatewayClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
                ConnectionId: &connectionID,
                Data:         []byte("<error:/>"),
            })
            if err != nil {
                log.Println("Error sending error token to client: %s", err)
                awsutils.HandleDeleteConnection(ctx, &connectionID, "sending \"<error:/>\" in PostConnection")
            }
            break
	    }

        _, err = awsutils.APIGatewayClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: &connectionID,
			Data:         jsonData,
		})
        if err != nil {
		    log.Println("Error sending token to client: %s", err)
            awsutils.HandleDeleteConnection(ctx, &connectionID, "sending token in PostConnection")
            break
        }
	}

    if err != nil {
        _, err := awsutils.APIGatewayClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
            ConnectionId: &connectionID,
            Data:         []byte("<end:)>"),
        })
        if err != nil {
            log.Println("Error sending <end:)> thingy to client: %s", err)
            awsutils.HandleDeleteConnection(ctx, &connectionID, "sending \"<end:/>\" in PostConnection")
        }
    }

	response := events.APIGatewayProxyResponse{
		StatusCode: 200,
        Body:       `"SIIUUUUU! :D"`,
	}

	return response, nil
}


func main() {
    lambda.Start(handler)
}

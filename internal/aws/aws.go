package awsutils

import (
	"context"
	"log"
	"sapopinguino/internal/config"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

var (
	SSMClient *ssm.Client
	KMSClient *kms.Client
	APIGatewayClient *apigatewaymanagementapi.Client
)

func ConfigAWS() {
    cfg, err := awsConfig.LoadDefaultConfig(
        context.TODO(),
    )
	if err != nil {
        log.Fatalf("Error while loading the AWS config: %s", err)
	}    

    SSMClient = ssm.NewFromConfig(cfg)
    KMSClient = kms.NewFromConfig(cfg)
    APIGatewayClient = apigatewaymanagementapi.NewFromConfig(cfg, func(o *apigatewaymanagementapi.Options) {
		o.BaseEndpoint = &config.C.Websockets.Endpoint
    })
}

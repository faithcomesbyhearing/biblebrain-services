package sign

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func GetSsmParameter(ctx context.Context, ssmClient *ssm.Client, parameterName string) *string {
	input := &ssm.GetParameterInput{
		Name:           &parameterName,
		WithDecryption: NewTrue(),
	}

	output, err := ssmClient.GetParameter(ctx, input)
	if err != nil {
		log.Panicf("unable to retrieve value from SSM at parameter name: %s, error: %s", parameterName, err.Error())
	}

	return output.Parameter.Value
}

func NewTrue() *bool {
	b := true

	return &b
}

func GetSSMClient(ctx context.Context) *ssm.Client {
	// NOTE: typically, if IS_OFFLINE is true, we would configure a local endpoint for the service.
	// However, it does not appear that serverless_offline_ssm exposes an endpoint.
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithDefaultRegion("us-west-2"),
	)
	if err != nil {
		log.Panic(err)
	}

	return ssm.NewFromConfig(cfg)
}

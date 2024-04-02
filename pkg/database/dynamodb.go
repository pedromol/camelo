package database

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/ServiceWeaver/weaver"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pedromol/camelo/pkg/model"
)

type DynamoDB struct {
	weaver.Implements[Database]
	Client *dynamodb.Client
}

func (h *DynamoDB) Init(ctx context.Context) error {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				host := os.Getenv("DYNAMODB_HOST")
				if host == "" {
					log.Println("DYNAMODB_HOST not set")
					return aws.Endpoint{}, errors.New("DYNAMODB_HOST not set")
				}
				return aws.Endpoint{URL: host}, nil
			})),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "camelo", SecretAccessKey: "camelo", SessionToken: "camelo",
				Source: "values are irrelevant for local DynamoDB",
			},
		}),
	)
	if err != nil {
		return err
	}

	h.Client = dynamodb.NewFromConfig(cfg)
	return h.migrate(ctx)
}

func (h DynamoDB) migrate(ctx context.Context) error {
	h.Client.CreateTable(ctx, &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: aws.String("id"),
			AttributeType: types.ScalarAttributeTypeS,
		}},
		KeySchema: []types.KeySchemaElement{{
			AttributeName: aws.String("id"),
			KeyType:       types.KeyTypeHash,
		}},
		TableName: aws.String("camelo"),
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
	})
	waiter := dynamodb.NewTableExistsWaiter(h.Client)
	err := waiter.Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String("camelo")}, 5*time.Minute)

	return err
}

func (h DynamoDB) Get(ctx context.Context, id string) (model.Tag, error) {
	output, err := h.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("camelo"),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return model.Tag{}, err
	}

	if output.Item["device"] == nil {
		return model.Tag{}, errors.New("not found")
	}
	device := output.Item["device"].(*types.AttributeValueMemberS).Value
	return model.Tag{
		Id:     id,
		Device: device,
	}, nil
}

func (h DynamoDB) Upsert(ctx context.Context, model model.Tag) error {
	_, err := h.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("camelo"),
		Item: map[string]types.AttributeValue{
			"id":     &types.AttributeValueMemberS{Value: model.Id},
			"device": &types.AttributeValueMemberS{Value: model.Device},
		},
	})
	return err
}

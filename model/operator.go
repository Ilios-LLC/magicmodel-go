package model

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog"
	"time"
)

type Operator struct {
	Err          error
	IsWhereChain bool
}

type Deconstructed struct {
	FieldName  string
	FieldValue interface{}
	FieldType  string
}

var svc *dynamodb.Client
var dynamoDBTableName string

// region *string, provider *credentials.StaticCredentialsProvider
func NewMagicModelOperator(ctx context.Context, tableName string, endpoint *string, optFns ...func(options *config.LoadOptions) error) (*Operator, error) {
	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, fmt.Errorf("an error occurred when getting aws config %s", err)
	}

	var optFnsDynamodb []func(*dynamodb.Options)
	if endpoint != nil {
		optFnsDynamodb = append(optFnsDynamodb, func(o *dynamodb.Options) {
			o.BaseEndpoint = endpoint
		})
	}

	svc = dynamodb.NewFromConfig(cfg, optFnsDynamodb...)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	dynamoDBTableName = tableName

	err = createDynamoDBTable(ctx)
	if err != nil {
		//os.Exit(1)
		return nil, fmt.Errorf("encountered an error while creating DynamoDb table %s: %s", dynamoDBTableName, err)
	}
	return &Operator{
		Err: nil,
	}, nil
}

func createDynamoDBTable(ctx context.Context) error {
	// create DYNAMO DB table
	_, err := svc.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(dynamoDBTableName),
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: aws.String("Type"),
			AttributeType: types.ScalarAttributeTypeS,
		}, {
			AttributeName: aws.String("ID"),
			AttributeType: types.ScalarAttributeTypeS,
		}},
		KeySchema: []types.KeySchemaElement{{
			AttributeName: aws.String("Type"),
			KeyType:       types.KeyTypeHash,
		}, {
			AttributeName: aws.String("ID"),
			KeyType:       types.KeyTypeRange,
		}},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		if err.Error() != "ResourceInUseException: Cannot create preexisting table" {
			return nil
		}
		return fmt.Errorf("encountered an error during init operation: %v", err)
	}

	waiter := dynamodb.NewTableExistsWaiter(svc, func(o *dynamodb.TableExistsWaiterOptions) {
		o.MaxDelay = time.Second * 10
		o.MinDelay = time.Second * 5
	})
	_, err = waiter.WaitForOutput(ctx, &dynamodb.DescribeTableInput{TableName: &dynamoDBTableName}, 30*time.Second)
	if err != nil {
		return fmt.Errorf("error while waiting for table to be created: %s", err)
	}

	return nil
}

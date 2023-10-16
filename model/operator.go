package model

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog"
	"os"
	"time"
)

type Operator struct {
	Err error
}

type Deconstructed struct {
	FieldName  string
	FieldValue interface{}
	FieldType  string
}

var svc *dynamodb.Client
var dynamoDBTableName string

func Start() *Operator {
	return &Operator{
		Err: nil,
	}
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
		o.Region = "us-east-1"
		return nil
	})

	if err != nil {
		os.Exit(1)
	}

	svc = dynamodb.NewFromConfig(cfg)

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	dynamoDBTableName = os.Getenv("MM_DYNAMODB_TABLE_NAME")
	if dynamoDBTableName == "" {
		os.Exit(1)
	}

	err = createDynamoDBTable()
	if err != nil {
		os.Exit(1)
	}
}

func createDynamoDBTable() error {
	// create DYNAMO DB table
	_, err := svc.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
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
		} else {
			return fmt.Errorf("encountered an error during init operation: %v", err)
		}
	}

	time.Sleep(10 * time.Second)
	return nil
}

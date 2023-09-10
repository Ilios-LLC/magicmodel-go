package model

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

type Operator struct{}

type Deconstructed struct {
	FieldName  string
	FieldValue interface{}
	FieldType  string
}

var svc *dynamodb.Client
var dynamoDBTableName string

func Start() *Operator {
	return &Operator{}
}

func init() {
	log.Debug().Msg("Initializing MagicModel")
	cfg, err := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
		o.Region = "us-east-1"
		return nil
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	svc = dynamodb.NewFromConfig(cfg)

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	dynamoDBTableName = os.Getenv("MM_DYNAMODB_TABLE_NAME")
	log.Debug().Msg("DynamoDB table name: " + dynamoDBTableName)

	createDynamoDBTable()
}

func createDynamoDBTable() {
	log.Debug().Msg("Creating DynamoDB table")
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
			log.Debug().Msg("Table already exists, skipping")
			return
		} else {
			panic(err)
		}
	}

	log.Info().Msg("Waiting 10 seconds for table to be created, this is a one-time wait per new database")
	time.Sleep(10 * time.Second)
	log.Info().Msg("Table created")
}

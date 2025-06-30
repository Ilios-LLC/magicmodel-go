package model

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog"
	"time"
)

type Operator struct {
	Err               error
	IsWhereChain      bool
	PendingConditions []WhereV4Condition
	IsWhereV4Chain    bool
	db                DynamoDBAPI
	tableName         string
}

type WhereV4Condition struct {
	FieldName   string
	FieldValues []interface{}
}

type Deconstructed struct {
	FieldName  string
	FieldValue interface{}
	FieldType  string
}

// For backward compatibility
var svc DynamoDBAPI
var dynamoDBTableName string

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

	dbClient := dynamodb.NewFromConfig(cfg, optFnsDynamodb...)

	// For backward compatibility
	svc = dbClient
	dynamoDBTableName = tableName

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	operator := &Operator{
		Err:       nil,
		db:        dbClient,
		tableName: tableName,
	}

	err = operator.createDynamoDBTable(ctx)
	if err != nil {
		return nil, fmt.Errorf("encountered an error while creating DynamoDb table %s: %s", tableName, err)
	}

	return operator, nil
}

// NewMagicModelOperatorWithClient creates a new operator with a custom DynamoDB client
// This is useful for testing with mock clients
func NewMagicModelOperatorWithClient(dbClient DynamoDBAPI, tableName string) *Operator {
	// For backward compatibility
	svc = dbClient
	dynamoDBTableName = tableName

	return &Operator{
		Err:       nil,
		db:        dbClient,
		tableName: tableName,
	}
}

func (o *Operator) createDynamoDBTable(ctx context.Context) error {
	// create DYNAMO DB table
	_, err := o.db.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(o.tableName),
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
		var resourceInUse *types.ResourceInUseException
		if errors.As(err, &resourceInUse) {
			// Table already exists — that's fine, just continue
			return nil
		}
		// Unexpected error
		return fmt.Errorf("encountered an error during init operation: %w", err)
	}

	waiter := dynamodb.NewTableExistsWaiter(o.db, func(o *dynamodb.TableExistsWaiterOptions) {
		o.MaxDelay = time.Second * 10
		o.MinDelay = time.Second * 5
	})
	_, err = waiter.WaitForOutput(ctx, &dynamodb.DescribeTableInput{TableName: &o.tableName}, 30*time.Second)
	if err != nil {
		return fmt.Errorf("error while waiting for table to be created: %s", err)
	}

	return nil
}

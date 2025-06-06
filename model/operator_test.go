package model

import (
	"github.com/Ilios-LLC/magicmodel-go/mocks"
	"testing"
)

func TestNewMagicModelOperatorWithClient(t *testing.T) {
	// Create mock using Mockery
	mockDB := mocks.NewDynamoDBAPI(t)
	tableName := "test-table"

	// Create operator with mock
	op := NewMagicModelOperatorWithClient(mockDB, tableName)

	// Check if the operator was created correctly
	if op == nil {
		t.Errorf("Expected operator to be created, got nil")
		return
	}

	if op.Err != nil {
		t.Errorf("Expected Err to be nil, got %v", op.Err)
	}
}

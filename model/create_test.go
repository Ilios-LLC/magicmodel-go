package model

import (
	"errors"
	"github.com/Ilios-LLC/magicmodel-go/mocks"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOperator_Create(t *testing.T) {
	// Define test cases
	tests := []struct {
		name          string
		setupMock     func(dbMock *mocks.DynamoDBAPI)
		input         *TestUser
		presetID      string
		expectError   bool
		errorContains string
	}{
		{
			name: "success",
			setupMock: func(dbMock *mocks.DynamoDBAPI) {
				dbMock.On("PutItem", mock.Anything, mock.Anything, mock.Anything).Return(&dynamodb.PutItemOutput{}, nil)
			},
			input: &TestUser{
				Name:    "John Doe",
				Email:   "john@example.com",
				Age:     30,
				IsAdmin: true,
			},
			expectError: false,
		},
		{
			name: "item_already_exists",
			setupMock: func(dbMock *mocks.DynamoDBAPI) {
				// No need to set up mock as we won't reach the PutItem call
			},
			input: &TestUser{
				Name:    "John Doe",
				Email:   "john@example.com",
				Age:     30,
				IsAdmin: true,
			},
			presetID:      "existing-id",
			expectError:   true,
			errorContains: "item already exists",
		},
		{
			name: "other_failure",
			setupMock: func(dbMock *mocks.DynamoDBAPI) {
				dbMock.On("PutItem", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("dynamodb error"))
			},
			input: &TestUser{
				Name:    "John Doe",
				Email:   "john@example.com",
				Age:     30,
				IsAdmin: true,
			},
			expectError:   true,
			errorContains: "dynamodb error",
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock using Mockery
			mockDB := mocks.NewDynamoDBAPI(t)
			if tc.setupMock != nil {
				tc.setupMock(mockDB)
			}

			// Create operator with mock
			op := NewMagicModelOperatorWithClient(mockDB, "test-table")

			// Set ID if needed for the test
			if tc.presetID != "" {
				tc.input.ID = tc.presetID
			}

			// Call the Create method
			result := op.Create(tc.input)

			// Check for errors
			if tc.expectError {
				require.Error(t, result.Err)
				if tc.errorContains != "" {
					assert.Contains(t, result.Err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, result.Err)

				assert.NotEmpty(t, tc.input.ID, "ID was not set")
				assert.Equal(t, "test_user", tc.input.Type, "Type was not set correctly")
				assert.False(t, tc.input.CreatedAt.IsZero(), "CreatedAt was not set")
				assert.False(t, tc.input.UpdatedAt.IsZero(), "UpdatedAt was not set")
			}
			mockDB.AssertExpectations(t)
		})
	}
}

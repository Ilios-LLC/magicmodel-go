# MagicModel-Go

[![Go Reference](https://pkg.go.dev/badge/github.com/Ilios-LLC/magicmodel-go.svg)](https://pkg.go.dev/github.com/Ilios-LLC/magicmodel-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/Ilios-LLC/magicmodel-go)](https://goreportcard.com/report/github.com/Ilios-LLC/magicmodel-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

MagicModel-Go is a lightweight ORM-like library for Go applications that simplifies working with Amazon DynamoDB. It provides a clean, intuitive interface for common database operations while abstracting away the complexity of DynamoDB's API.

## Features

- **Simple Model Definition**: Define your data models as Go structs with embedded `model.Model`
- **Automatic Table Creation**: Tables are automatically created if they don't exist
- **CRUD Operations**: Easy-to-use Create, Read, Update, and Delete operations
- **Query Building**: Fluent interface for building complex queries
- **Soft Delete Support**: Built-in support for soft deletion of records
- **Type Safety**: Leverages Go's type system for compile-time safety

## Installation

```bash
go get github.com/Ilios-LLC/magicmodel-go
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"github.com/Ilios-LLC/magicmodel-go/model"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/rs/zerolog/log"
	"os"
)

// Define your model by embedding model.Model
type Dog struct {
	Name  string
	Breed string
	model.Model
}

func main() {
	// Initialize MagicModel with your DynamoDB table name and region
	mm, err := model.NewMagicModelOperator(context.Background(), "my-table", nil, config.WithRegion("us-east-1"))
	if err != nil || mm.Err != nil {
		log.Error().Str("InstantiateMagicModelOperator", "ApplyApp").Msg(fmt.Sprintf("Encountered an err: %s", err))
		os.Exit(1)
	}

	// Create a new dog
	dog := Dog{
		Name:  "Buddy",
		Breed: "Dalmatian",
	}

	// Save to DynamoDB
	o := mm.Create(&dog)
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	
	log.Info().Msg(fmt.Sprintf("Dog created successfully: %+v", dog))
	
	// Find a dog by ID
	var foundDog Dog
	o = mm.Find(&foundDog, dog.ID)
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	
	// Update a dog
	foundDog.Breed = "Labrador"
	o = mm.Update(&foundDog)
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	
	// Delete a dog
	o = mm.Delete(&foundDog)
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
}
```

## Advanced Usage

### Query with Where Clauses

```go
var dogs []Dog
o := mm.Where("Breed", "=", "Dalmatian").All(&dogs)
if o.Err != nil {
	fmt.Println(o.Err)
	os.Exit(1)
}
```

### Soft Delete

```go
// Soft delete (record remains in database but marked as deleted)
o := mm.SoftDelete(&dog)
if o.Err != nil {
	fmt.Println(o.Err)
	os.Exit(1)
}

// Find including soft-deleted items
var allDogs []Dog
o = mm.IncludeSoftDeleted().All(&allDogs)
if o.Err != nil {
	fmt.Println(o.Err)
	os.Exit(1)
}
```

## Local Development

For local development and testing, you can use a local DynamoDB instance:

```go
endpoint := "http://localhost:8000"
mm, err := model.NewMagicModelOperator(context.Background(), "my-table", &endpoint, config.WithRegion("local"))
```

## Testing

MagicModel-Go includes mocks for easy testing:

```go
// Create a mock DynamoDB client
mockDB := &mocks.DynamoDBAPI{}

// Initialize MagicModel with the mock
mm := model.NewMagicModelOperatorWithClient(mockDB, "test-table")
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2)
- Inspired by the simplicity of ActiveRecord and other ORMs
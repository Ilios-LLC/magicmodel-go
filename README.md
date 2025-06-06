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

	buddy := Dog{
		Name:  "Buddy",
		Breed: "Dalmatian",
	}

	fido := Dog{
		Name:  "Fido",
		Breed: "Labrador",
	}
	
	// Create dogs in DynamoDB
	o := mm.Create(&buddy)
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	log.Info().Msg(fmt.Sprintf("Dog created successfully: %+v", buddy))
	
	o = mm.Create(&fido)
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	log.Info().Msg(fmt.Sprintf("Dog created successfully: %+v", fido))
	
	// Find a dog by ID
	var foundDog Dog
	o = mm.Find(&foundDog, buddy.ID)
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	log.Info().Msg(fmt.Sprintf("Dog created successfully: %+v", foundDog))

	// Update a dog
	foundDog.Breed = "Labrador"
	o = mm.Save(&foundDog)
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	log.Info().Msg(fmt.Sprintf("Found dog successfully: %+v", foundDog))

	
	// Find all dogs
	var allDogs []Dog
	o = mm.All(&allDogs)
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	log.Info().Msg(fmt.Sprintf("Found all dogs successfully: %+v", allDogs))
	
	// Delete a dog
	o = mm.Delete(&foundDog)
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	log.Info().Msg(fmt.Sprintf("Dog deleted successfully: %+v", foundDog))

}
```

## Advanced Usage

### Query with Where Clauses

```go
// Find dogs with specific attribute
var labradors []Dog
o = mm.WhereV3(false, &labradors, "Breed", "Labrador")
if o.Err != nil {
    fmt.Println(o.Err)
    os.Exit(1)
}
log.Info().Msg(fmt.Sprintf("Found labradors successfully: %+v", labradors))

// Find dogs with chained attributes
var labradorsNamedFido []Dog
o = mm.WhereV3(true, &labradorsNamedFido, "Breed", "Labrador").WhereV3(false, &labradorsNamedFido, "Name", "Fido")
if o.Err != nil {
    fmt.Println(o.Err)
    os.Exit(1)
}
log.Info().Msg(fmt.Sprintf("Found labradors with name Fido successfully: %+v", labradorsNamedFido))
```

### Soft Delete

```go
// Soft delete (record remains in database but marked as deleted and is not returned in queries)
o := mm.SoftDelete(&dog)
if o.Err != nil {
	fmt.Println(o.Err)
	os.Exit(1)
}
```

## Local Development

Simply create a test file and run magic model commands against a DynamoDB table.

The function `NewMagicModelOperator` will create the table if it does not exist, so you can run tests without needing to manually set up the DynamoDB table.

Alternatively, you can use [localstack](https://docs.localstack.cloud/user-guide/aws/dynamodb/) to run a local DynamoDB instance for testing.

## Testing
MagicModel-Go utilizes `mockery` to generate mocks for the DynamoDB client, allowing you to write unit tests without needing a live DynamoDB instance.

To run tests, use the following command:

```bash
go test ./...
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
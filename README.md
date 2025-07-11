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

> **NOTE**: In order to implement chaining (WhereV3), you MUST check the error after each operation (if o.Err != nil) before proceeding to the next operation. This is because the chaining will not work if the previous operation fails.

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

#### WhereV3 (Legacy - In-Memory Filtering for Chained Queries)

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

#### WhereV4 (Recommended - Efficient Single Query with OR Support)

WhereV4 provides improved performance by building a single DynamoDB query instead of multiple queries. It supports both single values and arrays for OR conditions within the same field.

```go
// Find dogs with single condition
var labradors []Dog
o = mm.WhereV4(false, &labradors, "Breed", "Labrador")
if o.Err != nil {
    fmt.Println(o.Err)
    os.Exit(1)
}
log.Info().Msg(fmt.Sprintf("Found labradors successfully: %+v", labradors))

// Find dogs with OR conditions (multiple breeds)
var multiBreedDogs []Dog
o = mm.WhereV4(false, &multiBreedDogs, "Breed", []string{"Labrador", "Dalmatian", "Beagle"})
if o.Err != nil {
    fmt.Println(o.Err)
    os.Exit(1)
}
log.Info().Msg(fmt.Sprintf("Found dogs with multiple breeds successfully: %+v", multiBreedDogs))

// Find dogs with chained AND conditions
var labradorsNamedFido []Dog
o = mm.WhereV4(true, &labradorsNamedFido, "Breed", "Labrador").WhereV4(false, &labradorsNamedFido, "Name", "Fido")
if o.Err != nil {
    fmt.Println(o.Err)
    os.Exit(1)
}
log.Info().Msg(fmt.Sprintf("Found labradors with name Fido successfully: %+v", labradorsNamedFido))

// Complex example: Find dogs that are (Labrador OR Dalmatian) AND (Name is Buddy OR Fido) AND Age is 3
var complexQuery []Dog
o = mm.WhereV4(true, &complexQuery, "Breed", []string{"Labrador", "Dalmatian"}).
     WhereV4(true, &complexQuery, "Name", []string{"Buddy", "Fido"}).
     WhereV4(false, &complexQuery, "Age", 3)
if o.Err != nil {
    fmt.Println(o.Err)
    os.Exit(1)
}
log.Info().Msg(fmt.Sprintf("Found dogs matching complex criteria successfully: %+v", complexQuery))
```

**WhereV4 Key Features:**
- **Single Query Execution**: All conditions are combined into one DynamoDB query for better performance
- **OR Support**: Pass arrays for OR conditions within the same field (e.g., `[]string{"value1", "value2"}`)
- **Flexible Input**: Accepts both single values and arrays - single values are automatically converted to single-element arrays
- **Backward Compatible**: Same chaining syntax as WhereV3 with `isChain` parameter
- **Deferred Execution**: Query only executes when `isChain=false`, allowing efficient condition accumulation

### Soft Delete

```go
// Soft delete (record remains in database but marked as deleted and is not returned in list queries)
o := mm.SoftDelete(&dog)
if o.Err != nil {
	fmt.Println(o.Err)
	os.Exit(1)
}
// SoftDelete still returns the object if you use Find
var foundDog Dog
o = mm.Find(&foundDog, buddy.ID)
if o.Err != nil {
    fmt.Println(o.Err)
    os.Exit(1)
}
```

## Local Development and Testing

MagicModel-Go includes comprehensive integration tests in `integration_test.go` that demonstrate all the key features of the library and verify they work correctly against a real DynamoDB instance.

MagicModel-Go utilizes `mockery` to generate mocks for the DynamoDB client, allowing you to write unit tests without needing a live DynamoDB instance.

### Run Tests
Run tests with:
```bash
go test ./...
```

### Running Integration Tests with LocalStack

As a part of `go test ./...` an integration test again a LocalStack DynamoDB table will run. This is the most comprehensive way to test the library.

This script requires Docker and Docker Compose to be installed on your system.

The function `NewMagicModelOperator` will create the table if it does not exist, so you can run tests without needing to manually set up the DynamoDB table.

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
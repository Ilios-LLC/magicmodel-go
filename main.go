package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Ilios-LLC/magicmodel-go/model"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/rs/zerolog/log"
	"os"
)

// Command line flags
var dynamoDBEndpoint string

type Dog struct {
	Name  string
	Breed string
	model.Model
}

func init() {
	// Parse command line flags
	flag.StringVar(&dynamoDBEndpoint, "endpoint", "", "DynamoDB endpoint URL (for local testing)")
	flag.Parse()
}

func main() {
	// Set up the DynamoDB endpoint
	var endpoint *string
	if dynamoDBEndpoint != "" {
		endpoint = &dynamoDBEndpoint
		fmt.Println("Using DynamoDB endpoint:", dynamoDBEndpoint)
	}

	// Initialize MagicModel with the specified endpoint
	mm, err := model.NewMagicModelOperator(context.Background(), "my-table", endpoint, config.WithRegion("us-east-1"))
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

	spike := Dog{
		Name:  "Spike",
		Breed: "Retriever",
	}

	// Create dogs in DynamoDB
	o := mm.Create(&buddy)
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	fmt.Println()
	fmt.Println(fmt.Sprintf("Dog created successfully: %+v", buddy))

	o = mm.Create(&fido)
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	fmt.Println()
	fmt.Println(fmt.Sprintf("Dog created successfully: %+v", fido))

	o = mm.Create(&spike)
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	fmt.Println()
	fmt.Println(fmt.Sprintf("Dog created successfully: %+v", spike))

	// Find a dog by ID
	var foundDog Dog
	o = mm.Find(&foundDog, buddy.ID)
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	fmt.Println()
	fmt.Println(fmt.Sprintf("Dog found successfully: %+v", foundDog))

	// Update a dog
	foundDog.Breed = "Labrador"
	o = mm.Save(&foundDog)
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	fmt.Println()
	fmt.Println(fmt.Sprintf("Found dog successfully: %+v", foundDog))

	// Find all dogs
	var allDogs []Dog
	o = mm.All(&allDogs)
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	fmt.Println()
	fmt.Println(fmt.Sprintf("Found all dogs successfully: %+v", allDogs))

	// Find dogs with specific attribute
	var labradors []Dog
	o = mm.WhereV3(false, &labradors, "Breed", "Labrador")
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	fmt.Println()
	fmt.Println(fmt.Sprintf("Found labradors successfully: %+v", labradors))

	// Find dogs with chained attributes
	var labradorsNamedFido []Dog
	o = mm.WhereV3(true, &labradorsNamedFido, "Breed", "Labrador").WhereV3(false, &labradorsNamedFido, "Name", "Fido")
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	fmt.Println()
	fmt.Println(fmt.Sprintf("Found labradors with name Fido successfully: %+v", labradorsNamedFido))

	// Find dogs with chained attributes
	var noDalmatians []Dog
	o = mm.WhereV3(true, &noDalmatians, "Breed", "Dalmatian").WhereV3(false, &noDalmatians, "Name", "Fido")
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	fmt.Println()
	fmt.Println(fmt.Sprintf("No dalmatians with found: %+v", noDalmatians))

	// Find dogs with chained attributes
	var labradorWithNameSpike []Dog
	o = mm.WhereV3(true, &labradorWithNameSpike, "Breed", "Labrador").WhereV3(false, &labradorWithNameSpike, "Name", "Spike")
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	fmt.Println()
	fmt.Println(fmt.Sprintf("No labradors with name Spike: %+v", labradorWithNameSpike))

	// Cleanup
	var allDogsToDelete []Dog
	o = mm.All(&allDogsToDelete)
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	fmt.Println()
	fmt.Println("Beginning cleanup of all dogs...")
	for _, dog := range allDogsToDelete {
		o = mm.Delete(&dog)
		if o.Err != nil {
			fmt.Println(o.Err)
			os.Exit(1)
		}
		fmt.Println(fmt.Sprintf("Deleted dog successfully: %+v", dog))
	}
}

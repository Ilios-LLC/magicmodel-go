package integration

import (
	"context"
	"flag"
	"fmt"
	"github.com/Ilios-LLC/magicmodel-go/model"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"os"
	"testing"
)

// Command line flags
var dynamoDBEndpoint string

// Dog is our test model
type Dog struct {
	Name  string
	Breed string
	Home  Home
	model.Model
}

type Home struct {
	FamilyName string
	Address    Address
}
type Address struct {
	Street string
	City   string
}

var mm *model.Operator
var localstack testcontainers.Container

func TestMain(m *testing.M) {
	flag.StringVar(&dynamoDBEndpoint, "endpoint", "", "DynamoDB endpoint URL (for local testing)")
	flag.Parse()

	ctx := context.Background()
	setupTest(ctx)

	// Run tests
	code := m.Run()
	// Teardown
	_ = localstack.Terminate(ctx)
	os.Exit(code)
}

// setupTest initializes the MagicModel operator
func setupTest(ctx context.Context) {

	// Start LocalStack container
	req := testcontainers.ContainerRequest{
		Image:        "localstack/localstack:latest",
		ExposedPorts: []string{"4566/tcp"},
		Env: map[string]string{
			"SERVICES": "dynamodb",
		},
		WaitingFor: wait.ForLog("Ready."),
	}
	var err error
	localstack, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("failed to start localstack: %v", err)
	}

	// Get endpoint
	host, err := localstack.Host(ctx)
	if err != nil {
		log.Fatalf("error getting localstack host: %v", err)
	}
	port, err := localstack.MappedPort(ctx, "4566")
	if err != nil {
		log.Fatalf("error getting localstack port: %v", err)
	}
	endpoint := fmt.Sprintf("http://%s:%s", host, port.Port())

	mm, err = model.NewMagicModelOperator(ctx, "MyTable", &endpoint,
		func(o *config.LoadOptions) error {
			o.Region = "us-east-1"
			o.Credentials = aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test", "test", ""))
			return nil
		},
	)
	if err != nil {
		log.Fatalf("failed to init operator: %v", err)
	}
	return
}

// cleanupDogs removes all dogs from the database
func cleanupDogs(t *testing.T) {
	var allDogs []Dog
	o := mm.All(&allDogs)
	if o.Err != nil {
		t.Logf("Warning: Failed to fetch dogs for cleanup: %v", o.Err)
		o.Err = nil
		return
	}

	t.Logf("Cleaning up %d dogs", len(allDogs))
	for _, dog := range allDogs {
		o = mm.Delete(&dog)
		if o.Err != nil {
			t.Logf("Warning: Failed to delete dog %s: %v", dog.ID, o.Err)
			o.Err = nil
		}
	}
}

// cleanupDogs removes all dogs from the database
func cleanupDogsSoftDelete(t *testing.T, dog *Dog) {
	o := mm.Delete(dog)
	if o.Err != nil {
		t.Fatalf("Warning: Failed to delete dog: %v", o.Err)
	}
}

func TestCreateAndFind(t *testing.T) {
	defer cleanupDogs(t)

	// Create a dog
	buddy := Dog{
		Name:  "Buddy",
		Breed: "Dalmatian",
	}

	o := mm.Create(&buddy)
	if o.Err != nil {
		t.Fatalf("Failed to create dog: %v", o.Err)
	}

	// Verify the dog was created with an ID
	if buddy.ID == "" {
		t.Errorf("Expected dog to have an ID, but it was empty")
	}
	t.Logf("Created dog with ID: %s", buddy.ID)

	// Find the dog by ID
	var foundDog Dog
	o = mm.Find(&foundDog, buddy.ID)
	if o.Err != nil {
		t.Fatalf("Failed to find dog: %v", o.Err)
	}

	// Verify the found dog matches the created dog
	if foundDog.ID != buddy.ID {
		t.Errorf("Expected found dog ID to be %s, got %s", buddy.ID, foundDog.ID)
	}
	if foundDog.Name != buddy.Name {
		t.Errorf("Expected found dog name to be %s, got %s", buddy.Name, foundDog.Name)
	}
	if foundDog.Breed != buddy.Breed {
		t.Errorf("Expected found dog breed to be %s, got %s", buddy.Breed, foundDog.Breed)
	}
}

// TestUpdate tests updating a dog
func TestUpdate(t *testing.T) {
	defer cleanupDogs(t)

	// Create a dog
	dog := Dog{
		Name:  "Rex",
		Breed: "German Shepherd",
	}

	o := mm.Create(&dog)
	if o.Err != nil {
		t.Fatalf("Failed to create dog: %v", o.Err)
	}

	// Update the dog
	dog.Breed = "Belgian Shepherd"
	o = mm.Save(&dog)
	if o.Err != nil {
		t.Fatalf("Failed to update dog: %v", o.Err)
	}

	// Find the dog to verify the update
	var updatedDog Dog
	o = mm.Find(&updatedDog, dog.ID)
	if o.Err != nil {
		t.Fatalf("Failed to find updated dog: %v", o.Err)
	}

	// Verify the update was successful
	if updatedDog.Breed != "Belgian Shepherd" {
		t.Errorf("Expected updated dog breed to be 'Belgian Shepherd', got '%s'", updatedDog.Breed)
	}
}

// TestDelete tests deleting a dog
func TestDelete(t *testing.T) {
	// Create a dog
	dog := Dog{
		Name:  "Spot",
		Breed: "Beagle",
	}

	o := mm.Create(&dog)
	if o.Err != nil {
		t.Fatalf("Failed to create dog: %v", o.Err)
	}

	// Delete the dog
	o = mm.Delete(&dog)
	if o.Err != nil {
		t.Fatalf("Failed to delete dog: %v", o.Err)
	}

	// Try to find the deleted dog
	var deletedDog Dog
	o = mm.Find(&deletedDog, dog.ID)

	// Verify the dog was deleted
	if o.Err == nil {
		t.Errorf("Expected error when finding deleted dog, but got nil")
	}
	o.Err = nil
}

func TestSoftDelete(t *testing.T) {

	// Create a dog
	dog := Dog{
		Name:  "Spot",
		Breed: "Beagle",
	}

	o := mm.Create(&dog)
	if o.Err != nil {
		t.Fatalf("Failed to create dog: %v", o.Err)
	}

	defer cleanupDogsSoftDelete(t, &dog)

	// Delete the dog
	o = mm.SoftDelete(&dog)
	if o.Err != nil {
		t.Fatalf("Failed to delete dog: %v", o.Err)
	}

	// Find the dog by ID
	var foundDog Dog
	o = mm.Find(&foundDog, dog.ID)
	if o.Err != nil {
		t.Fatalf("Failed to find dog: %v", o.Err)
	}

	// Verify the found dog matches the created dog
	if foundDog.ID != dog.ID {
		t.Errorf("Expected found dog ID to be %s, got %s", dog.ID, foundDog.ID)
	}
	if foundDog.Name != dog.Name {
		t.Errorf("Expected found dog name to be %s, got %s", dog.Name, foundDog.Name)
	}
	if foundDog.Breed != dog.Breed {
		t.Errorf("Expected found dog breed to be %s, got %s", dog.Breed, foundDog.Breed)
	}
}

// TestAll tests retrieving all dogs
func TestAll(t *testing.T) {

	defer cleanupDogs(t)

	// Create multiple dogs
	dogs := []Dog{
		{Name: "Buddy", Breed: "Dalmatian"},
		{Name: "Fido", Breed: "Labrador"},
		{Name: "Spike", Breed: "Retriever"},
	}

	for i := range dogs {
		o := mm.Create(&dogs[i])
		if o.Err != nil {
			t.Fatalf("Failed to create dog %s: %v", dogs[i].Name, o.Err)
		}
	}

	// Retrieve all dogs
	var allDogs []Dog
	o := mm.All(&allDogs)
	if o.Err != nil {
		t.Fatalf("Failed to retrieve all dogs: %v", o.Err)
	}

	// Verify we got at least the number of dogs we created
	if len(allDogs) != len(dogs) {
		t.Errorf("Expected at least %d dogs, got %d", len(dogs), len(allDogs))
	}
}

// TestWhereV3 tests filtering dogs with WhereV3
func TestWhereV3(t *testing.T) {

	defer cleanupDogs(t)

	// Create multiple dogs with different breeds and nested fields
	dogs := []Dog{
		{Name: "Buddy", Breed: "Dalmatian", Home: Home{
			FamilyName: "Miller",
			Address: Address{
				Street: "123 Bark St",
			},
		}},
		{Name: "Fido", Breed: "Labrador", Home: Home{
			FamilyName: "Smith",
			Address: Address{
				Street: "9723 Bark St",
				City:   "Cattown",
			},
		}},
		{Name: "Rex", Breed: "Labrador", Home: Home{
			FamilyName: "Miller",
			Address: Address{
				Street: "12344 Bark St",
				City:   "Dogtown",
			},
		}},
		{Name: "Spike", Breed: "Retriever"},
	}

	for i := range dogs {
		o := mm.Create(&dogs[i])
		if o.Err != nil {
			t.Fatalf("Failed to create dog %s: %v", dogs[i].Name, o.Err)
		}
	}

	// Test 1: Find all Labradors
	t.Run("FindLabradors", func(t *testing.T) {
		var labradors []Dog
		o := mm.WhereV3(false, &labradors, "Breed", "Labrador")
		if o.Err != nil {
			t.Fatalf("Failed to find Labradors: %v", o.Err)
		}

		if len(labradors) != 2 {
			t.Errorf("Expected 2 Labradors, got %d", len(labradors))
		}

		// Verify all dogs are Labradors
		for _, dog := range labradors {
			if dog.Breed != "Labrador" {
				t.Errorf("Expected dog breed to be 'Labrador', got '%s'", dog.Breed)
			}
		}
	})

	// Test 2: Find Labradors named Fido
	t.Run("FindLabradorsNamedFido", func(t *testing.T) {
		var labradorsNamedFido []Dog
		o := mm.WhereV3(true, &labradorsNamedFido, "Breed", "Labrador").WhereV3(false, &labradorsNamedFido, "Name", "Fido")
		if o.Err != nil {
			t.Fatalf("Failed to find Labradors named Fido: %v", o.Err)
		}

		if len(labradorsNamedFido) != 1 {
			t.Errorf("Expected 1 Labrador named Fido, got %d", len(labradorsNamedFido))
		}

		if len(labradorsNamedFido) > 0 {
			if labradorsNamedFido[0].Name != "Fido" || labradorsNamedFido[0].Breed != "Labrador" {
				t.Errorf("Expected dog to be a Labrador named Fido, got %+v", labradorsNamedFido[0])
			}
		}
	})

	// Test 3: Find Dalmatians named Fido (should be empty)
	t.Run("FindDalmatiansNamedFido", func(t *testing.T) {
		var dalmatiansNamedFido []Dog
		o := mm.WhereV3(true, &dalmatiansNamedFido, "Breed", "Dalmatian").WhereV3(false, &dalmatiansNamedFido, "Name", "Fido")
		if o.Err != nil {
			t.Fatalf("Failed to find Dalmatians named Fido: %v", o.Err)
		}

		if len(dalmatiansNamedFido) != 0 {
			t.Errorf("Expected 0 Dalmatians named Fido, got %d", len(dalmatiansNamedFido))
		}
	})

	// Test 4: Find Labradors named Spike (should be empty)
	t.Run("FindLabradorsNamedSpike", func(t *testing.T) {
		var labradorsNamedSpike []Dog
		o := mm.WhereV3(true, &labradorsNamedSpike, "Breed", "Labrador").WhereV3(false, &labradorsNamedSpike, "Name", "Spike")
		if o.Err != nil {
			t.Fatalf("Failed to find Labradors named Spike: %v", o.Err)
		}

		if len(labradorsNamedSpike) != 0 {
			t.Errorf("Expected 0 Labradors named Spike, got %d", len(labradorsNamedSpike))
		}
	})

	// Test 5: Find all dogs with a specific family name in their home
	t.Run("FindDogsNestedFilters", func(t *testing.T) {
		var labradorsWithMillers []Dog
		o := mm.WhereV3(true, &labradorsWithMillers, "Breed", "Labrador").WhereV3(false, &labradorsWithMillers, "Home.FamilyName", "Miller")
		if o.Err != nil {
			t.Fatalf("Failed to find Labradors: %v", o.Err)
		}

		if len(labradorsWithMillers) != 1 {
			t.Errorf("Expected 1 Labradors living with the Millers, got %d", len(labradorsWithMillers))
		}

		// Verify all dogs are Labradors
		for _, dog := range labradorsWithMillers {
			if dog.Breed != "Labrador" {
				t.Errorf("Expected dog breed to be 'Labrador', got '%s'", dog.Breed)
			}
			if dog.Home.FamilyName != "Miller" {
				t.Errorf("Expected dog family name to be 'Miller', got '%s'", dog.Home.FamilyName)
			}
		}

		var dalmatiansWithSmiths []Dog
		o = mm.WhereV3(true, &dalmatiansWithSmiths, "Breed", "Dalmatian").WhereV3(false, &dalmatiansWithSmiths, "Home.FamilyName", "Smith")
		if o.Err != nil {
			t.Fatalf("Failed to find Labradors: %v", o.Err)
		}

		if len(dalmatiansWithSmiths) != 0 {
			t.Errorf("Expected 1 Labradors living with the Millers, got %d", len(dalmatiansWithSmiths))
		}

		var dalmatiansWithSmithsV2 []Dog
		o = mm.WhereV3(true, &dalmatiansWithSmithsV2, "Home.FamilyName", "Smith").WhereV3(false, &dalmatiansWithSmithsV2, "Breed", "Dalmatian")
		if o.Err != nil {
			t.Fatalf("Failed to find Labradors: %v", o.Err)
		}

		if len(dalmatiansWithSmiths) != 0 {
			t.Errorf("Expected 1 Labradors living with the Millers, got %d", len(dalmatiansWithSmithsV2))
		}
	})

}

// TestWhereV2 tests filtering dogs with WhereV2
func TestWhereV2(t *testing.T) {

	defer cleanupDogs(t)

	// Create multiple dogs with different breeds
	dogs := []Dog{
		{Name: "Buddy", Breed: "Dalmatian"},
		{Name: "Fido", Breed: "Labrador"},
		{Name: "Rex", Breed: "Labrador"},
		{Name: "Spike", Breed: "Retriever"},
	}

	for i := range dogs {
		o := mm.Create(&dogs[i])
		if o.Err != nil {
			t.Fatalf("Failed to create dog %s: %v", dogs[i].Name, o.Err)
		}
	}

	// Test: Find all Labradors using WhereV2
	var labradors []Dog
	o := mm.WhereV2(false, &labradors, "Breed", "Labrador")
	if o.Err != nil {
		t.Fatalf("Failed to find Labradors with WhereV2: %v", o.Err)
	}

	if len(labradors) != 2 {
		t.Errorf("Expected 2 Labradors, got %d", len(labradors))
	}

	// Verify all dogs are Labradors
	for _, dog := range labradors {
		if dog.Breed != "Labrador" {
			t.Errorf("Expected dog breed to be 'Labrador', got '%s'", dog.Breed)
		}
	}
}

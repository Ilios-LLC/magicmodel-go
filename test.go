package main

import (
	"context"
	"fmt"
	"github.com/Ilios-LLC/magicmodel-go/model"
	"github.com/aws/aws-sdk-go-v2/aws"
	"os"
)

type MyStruct struct {
	model.Model
	Name   string
	Secret string `magicmodel:"sensitive"`
}

func main() {
	// instantiate magic model with kms key
	os.Setenv("AWS_REGION", "us-east-1")
	o, err := model.NewMagicModelOperator(context.Background(), "kias-table", aws.String("arn:aws:kms:us-east-1:677323703079:key/6f2cc367-206d-4706-b705-90d827549bb9"))
	if err != nil {
		fmt.Printf("\nerror creating magic model operator: %s\n", err.Error())
		os.Exit(1)
	}
	//test encrypt
	myStruct := MyStruct{
		Name:   "John Doe",
		Secret: "thisIsASecret",
	}
	// create a new record
	o = o.Create(&myStruct)
	if o.Err != nil {
		fmt.Printf("\nerror creating magic model operator: %s\n", o.Err.Error())
		os.Exit(1)
	}
	// test decrypt
	//myStructs := []MyStruct{}
	//// create a new record
	//o = o.All(&myStructs)
	//if o.Err != nil {
	//	fmt.Printf("\nerror creating magic model operator: %s\n", o.Err.Error())
	//	os.Exit(1)
	//}
}

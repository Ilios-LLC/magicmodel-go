package main

import (
	"context"
	"fmt"
	"github.com/Ilios-LLC/magicmodel-go/model"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/rs/zerolog/log"
	"os"
)

type Dog struct {
	Name  string
	Breed string
	model.Model
}

func main() {
	mm, err := model.NewMagicModelOperator(context.Background(), "my-table", nil, config.WithRegion("us-east-1"))
	if err != nil || mm.Err != nil {
		log.Error().Str("InstantiateMagicModelOperator", "ApplyApp").Msg(fmt.Sprintf("Encountered an err: %s", err))
		os.Exit(1)
	}
	dog := Dog{

		Name:  "Buddy",
		Breed: "Dalmatian",
	}

	o := mm.Create(&dog)
	if o.Err != nil {
		fmt.Println(o.Err)
		os.Exit(1)
	}
	log.Info().Msg(fmt.Sprintf("Dog created successfully: %+v", dog))
}

package main

import (
	"fmt"

	"github.com/nullexp/sanaz.crm.api/configs"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/log"
)

func init() {
	log.Initialize()
}

func main() {
	log.Info.Println("initializing server")

	conf := configs.ReadConfig()
	fmt.Printf("%+v", conf)
	fmt.Println("running")
	initializeApi(conf)
}

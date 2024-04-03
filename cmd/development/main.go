package main

import (
	"fmt"

	"git.omidgolestani.ir/clinic/crm.api/configs"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/log"
)

func init() {
	log.Initialize()
}

func main() {

	log.Info.Println("initializing server")

	conf := configs.ReadConfig()
	fmt.Printf("%+v", conf)
	fmt.Println("running")
}

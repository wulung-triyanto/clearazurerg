package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/wulung-triyanto/clean-azure-rg/pkg/app"
	"github.com/wulung-triyanto/clean-azure-rg/pkg/auto"
)

func main() {
	//Berfungsi untuk membaca environment variabel
	p, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT")
	if !ok {
		p = "8080"
	}

	a := app.New(fmt.Sprintf(":%s", p)) //Set router port sesuai dengan env value
	a.Router.Use(app.ContentTypeJson)   //Set router untuk menerima dan memberikan JSON value

	//Set method auto.HandleTick sebagai logic handler dari endpoint /CleanUpResourceGroups
	a.Router.HandleFunc("/CleanUpResourceGroups", auto.HandleTick).Methods(http.MethodPost)
	a.ListenAndServe() //Start router
}

package main

import (
	"log"
	"os"
)

const disapprovalLook string = "ಠ_ಠ"

func ExitWithErr(err error) {
	log.Fatal(err)
	os.Exit(-1)
}

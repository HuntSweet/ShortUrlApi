package main

import (
	"log"
	"testing"
)

func TestGetEnv(t *testing.T) {
	r := GetEnv()
	log.Print(r)
}

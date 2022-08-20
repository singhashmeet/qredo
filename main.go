package main

import (
	"fmt"
)

func main() {
	// var data = `{"name":{"first":"Janet","last":"Prichard"},"age":-47.3, "pool": 102, "numbers": [1,2,3,4]}`
	var data = `"ashmeet"`
	err := Validate(data)
	fmt.Println(err)
	fmt.Println(globalNumbers)
}

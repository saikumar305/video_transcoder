package main

import (
	"fmt"
)



var name string = "sai"

func add(a int , b int) int {
		return a+b
	}


func main() {
	age := 7
	fmt.Println("Hello, World!" , name)

	if age > 18 {
		fmt.Println("adult")
	}else{
		fmt.Println("minor")
	}

	// loops

	for i := 0; i<5; i++ {
		fmt.Println(i*i)
	}

	

	fmt.Println(add(1,3))

	x := 10
    p := &x // p is a pointer to x

    fmt.Println("Value of x:", x)
    fmt.Println("Address of x:", &x)
    fmt.Println("Value stored in p (address):", p)
    fmt.Println("Value pointed by p:", *p)
}



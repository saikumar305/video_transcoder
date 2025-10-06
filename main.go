package main

import (
	"fmt"
	"video_transcoder/config"
)


func main(){
	fmt.Println("Hello")

	db := config.DatabaseConnection()
	fmt.Println(db)
}
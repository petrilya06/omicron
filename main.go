package main

import "github.com/joho/godotenv"

func main() {
	if err := godotenv.Load(); err != nil {
		panic("can`t load .env, " + err.Error())
	}

}

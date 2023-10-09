package main

func main() {
	if err := loadConfig(); err != nil {
		panic(err)
	}

	if err := StartServer(); err != nil {
		panic(err)
	}
}

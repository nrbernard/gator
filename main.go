package main

import (
	"fmt"
	"github.com/nrbernard/gator/internal/config"
)

func main() {
	configFile, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}

	if err := configFile.SetUser("arthur"); err != nil {
		fmt.Println("Error setting user:", err)
		return
	}

	configFile, err = config.Read()
	if err != nil {
		fmt.Println("Error reading updated config:", err)
		return
	}

	fmt.Printf("Config: %+v\n", configFile)
}

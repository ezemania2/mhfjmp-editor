package main

import (
	"log"
	"mhfjmp-editor/extractor"
	"mhfjmp-editor/injector"
	"os"
)

func main() {
	log.Println("Starting MHF data tool")

	command := ""
	if len(os.Args) >= 2 {
		command = os.Args[1]
	} else {
		log.Fatalf("No command provided. Use 'extract(e)','inject(i)' or 'generate folders(gf)'")
	}

	log.Printf("Command received: '%s'. Processing...", command)

	switch command {
	case "gf":
		log.Println("Generating necessary folders for the program")

		// Define the necessary folders
		folders := []string{
			"input",
			"output",
		}

		// Create the folders if they do not exist
		for _, folder := range folders {
			if _, err := os.Stat(folder); os.IsNotExist(err) {
				err := os.Mkdir(folder, os.ModePerm)
				if err != nil {
					log.Fatalf("Failed to create folder '%s': %v", folder, err)
				}
				log.Printf("Folder '%s' created successfully", folder)
			} else {
				log.Printf("Folder '%s' already exists", folder)
			}
		}
	case "e":
		extractor.ExtractData()
		log.Println("Data extraction done!")
	case "i":
		injector.Start()
		log.Println("Data generation done!")
	default:
		log.Fatalf("Invalid command: '%s'. Use 'extract' or 'generate'", command)
	}
}

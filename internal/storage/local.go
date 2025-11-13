package storage

import (
	"fmt"
	"log"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

func DecodeFile(file *os.File, out any) error {
	// Decode the YAML file into a document object
	decoder := yaml.NewDecoder(file)
	// Out should be a pointer to a struct
	if err := decoder.Decode(out); err != nil {
		// Consider using a lower logging level or removing log output
		log.Printf("failed to decode file: %v", err)
		return fmt.Errorf("decode error: %w", err)
	}
	return nil
}

// WriteYAMLDocument writes a YAML document to local storage
func WriteYAMLDocument(objectKey string, formData map[string]any) (string, error) {

	fileName := path.Base(objectKey)
	localFilePath := path.Join("s3", fileName)

	if err := os.MkdirAll("s3", 0755); err != nil {
		log.Printf("Couldn't create s3 directory. Here's why: %v\n", err)
		return "", err
	}

	file, err := os.Create(localFilePath)
	if err != nil {
		log.Printf("Couldn't create file %v. Here's why: %v\n", localFilePath, err)
		return "", err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error closing file: %v", err)
		}
	}()

	enc := yaml.NewEncoder(file)
	defer func() {
		if err := enc.Close(); err != nil {
			log.Printf("error closing encoder: %v", err)
		}
	}()

	if err := enc.Encode(formData); err != nil {
		log.Printf("Couldn't encode document to YAML. Here's why: %v\n", err)
		return "", err
	}

	log.Printf("Successfully created document: %s", objectKey)
	return localFilePath, nil
}

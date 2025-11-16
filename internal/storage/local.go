package storage

import (
	"fmt"
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// Decodes a YAML file into the provided output structure
func DecodeFile(file io.Reader, out any) error {
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(out); err != nil {
		log.Printf("failed to decode file: %v", err)
		return fmt.Errorf("decode error: %w", err)
	}
	return nil
}

// WriteYAMLDocument writes a YAML document to local storage
func WriteYAMLDocument(localFilePath string, formData map[string]any) error {

	file, err := os.Create(localFilePath)
	if err != nil {
		log.Printf("Couldn't create file %v. Here's why: %v\n", localFilePath, err)
		return err
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
		return err
	}

	return nil
}

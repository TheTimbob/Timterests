package storage

import (
	"fmt"
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// DecodeFile decodes a YAML file into the provided output structure.
func DecodeFile(file io.Reader, out any) error {
	decoder := yaml.NewDecoder(file)

	err := decoder.Decode(out)
	if err != nil {
		log.Printf("failed to decode file: %v", err)

		return fmt.Errorf("decode error: %w", err)
	}

	return nil
}

// WriteYAMLDocument writes a YAML document to local storage.
func WriteYAMLDocument(localFilePath string, formData map[string]any) error {
	// #nosec G304 -- localFilePath is constructed from sanitized user input and restricted to s3/ directory
	file, err := os.Create(localFilePath)
	if err != nil {
		log.Printf("Couldn't create file %v. Here's why: %v\n", localFilePath, err)

		return fmt.Errorf("failed to create file: %w", err)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			log.Printf("error closing file: %v", err)
		}
	}()

	enc := yaml.NewEncoder(file)

	defer func() {
		err := enc.Close()
		if err != nil {
			log.Printf("error closing encoder: %v", err)
		}
	}()

	err = enc.Encode(formData)
	if err != nil {
		log.Printf("Couldn't encode document to YAML. Here's why: %v\n", err)

		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	return nil
}

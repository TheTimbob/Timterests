package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"timterests/internal/model"
	"timterests/internal/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// ListProjects retrieves all projects from storage, optionally filtering by tag.
// Pass tag="" or tag="all" to retrieve all projects.
func ListProjects(ctx context.Context, s storage.Storage, tag string) ([]model.Project, error) {
	var projects []model.Project

	prefix := "projects/"

	projectFiles, err := s.ListObjects(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	mdKeys := make(map[string]bool, len(projectFiles))
	for _, obj := range projectFiles {
		key := aws.ToString(obj.Key)
		if strings.HasSuffix(key, ".md") {
			mdKeys[key] = true
		}
	}

	docIdx := 0

	for _, obj := range projectFiles {
		key := aws.ToString(obj.Key)

		if key == prefix || !strings.HasSuffix(key, ".yaml") {
			continue
		}

		if !mdKeys[strings.TrimSuffix(key, ".yaml")+".md"] {
			log.Printf("ListProjects: skipping %s — no paired .md body file", key)

			continue
		}

		project, err := GetProject(ctx, s, key, docIdx)
		if err != nil {
			log.Printf("Failed to get project: %v", err)

			return nil, err
		}

		docIdx++

		if tag == "" || tag == "all" || slices.Contains(project.Tags, tag) {
			projects = append(projects, *project)
		}
	}

	return projects, nil
}

// GetProject retrieves a single project by its storage key and numeric ID,
// including downloading and resolving its associated image.
func GetProject(ctx context.Context, s storage.Storage, key string, id int) (*model.Project, error) {
	project, err := getDoc[model.Project](ctx, s, key, id)
	if err != nil {
		return nil, err
	}

	imagePath, err := s.GetImage(ctx, project.Image)
	if err != nil {
		log.Printf("Failed to download image: %v", err)

		return nil, fmt.Errorf("failed to resolve image %q: %w", project.Image, err)
	}

	project.Image = imagePath

	return project, nil
}

// GetFeaturedProject retrieves the project whose title matches featuredProjectTitle.
// Returns an error if no match is found.
func GetFeaturedProject(ctx context.Context, s storage.Storage, featuredProjectTitle string) (*model.Project, error) {
	projects, err := ListProjects(ctx, s, "all")
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return nil, errors.New("no projects found")
	}

	for _, project := range projects {
		if project.Title == featuredProjectTitle {
			p := project

			return &p, nil
		}
	}

	return nil, fmt.Errorf("no project matched the featured project title %q", featuredProjectTitle)
}

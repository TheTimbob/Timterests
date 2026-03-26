package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
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
		return nil, fmt.Errorf("failed to list S3 objects: %w", err)
	}

	for id, obj := range projectFiles {
		key := aws.ToString(obj.Key)

		if key == prefix {
			continue
		}

		project, err := GetProject(ctx, s, key, id)
		if err != nil {
			log.Printf("Failed to get project: %v", err)

			return nil, err
		}

		if slices.Contains(project.Tags, tag) || tag == "all" || tag == "" {
			projects = append(projects, *project)
		}
	}

	return projects, nil
}

// GetProject retrieves a single project by its storage key and numeric ID,
// including downloading and resolving its associated image.
func GetProject(ctx context.Context, s storage.Storage, key string, id int) (*model.Project, error) {
	var project model.Project

	project.ID = strconv.Itoa(id)
	project.S3Key = key

	err := s.GetPreparedFile(ctx, key, &project)
	if err != nil {
		return nil, fmt.Errorf("failed to get prepared file: %w", err)
	}

	localImagePath, err := s.GetImage(ctx, project.Image)
	if err != nil {
		log.Printf("Failed to download image: %v", err)

		return nil, fmt.Errorf("failed to get image from S3: %w", err)
	}

	project.Image = localImagePath

	return &project, nil
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
			p.Body = storage.RemoveHTMLTags(p.Body)

			return &p, nil
		}
	}

	return nil, fmt.Errorf("no project matched the featured project title %q", featuredProjectTitle)
}

package web

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"timterests/cmd/web/components"
	"timterests/internal/auth"
	"timterests/internal/model"
	"timterests/internal/storage"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/aws"
)

// Project represents a personal software project.
type Project struct {
	model.Document `yaml:",inline"`

	Repository string `yaml:"repository"`
	Image      string `yaml:"imagePath"`
}

// ProjectsPageHandler handles requests to the projects page,
// ensuring authentication and rendering the appropriate content.
func ProjectsPageHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, currentTag, design string) {
	var (
		component templ.Component
		tags      []string
	)

	projects, err := ListProjects(r.Context(), s, currentTag)
	if err != nil {
		message := "Failed to fetch projects"
		http.Error(w, fmt.Sprintf("%s: %v", message, err), http.StatusInternalServerError)

		return
	}

	for i := range projects {
		projects[i].Body = storage.RemoveHTMLTags(projects[i].Body)
		v := reflect.ValueOf(projects[i])
		tags = storage.GetTags(v, tags)
	}

	if r.Header.Get("Hx-Request") == "true" {
		component = ProjectsList(projects, design)
	} else {
		component = ProjectsListPage(projects, tags, design)
	}

	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error rendering in ProjectPosts: %e", err)
	}
}

// GetProjectHandler handles requests to get a specific project by its ID.
func GetProjectHandler(w http.ResponseWriter, r *http.Request, s storage.Storage, projectID string, a *auth.Auth) {
	projects, err := ListProjects(r.Context(), s, "all")
	if err != nil {
		http.Error(w, "Failed to fetch projects", http.StatusInternalServerError)

		return
	}

	for _, project := range projects {
		if project.ID == projectID {
			var component templ.Component

			authenticated := a.IsAuthenticated(r)

			if r.Header.Get("Hx-Request") == "true" {
				component = ProjectDisplay(project, authenticated)
			} else {
				component = ProjectPage(project, authenticated)
			}

			err = component.Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Printf("Error rendering in GetProjectsHandler: %e", err)
			}
		}
	}
}

// ListProjects retrieves a list of projects from storage,
// optionally filtering by tag.
func ListProjects(ctx context.Context, s storage.Storage, tag string) ([]Project, error) {
	var projects []Project

	// Get all projects from the storage
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

		project, err := GetProject(ctx, key, id, s)
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

// GetProject retrieves a single project from storage based on its S3 key and ID.
func GetProject(ctx context.Context, key string, id int, s storage.Storage) (*Project, error) {
	var project Project

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

// GetFeaturedProject retrieves the Timterests Project.
func GetFeaturedProject(ctx context.Context, s storage.Storage, featuredProjectTitle string) (*Project, error) {
	projects, err := ListProjects(ctx, s, "all")
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return nil, errors.New("no projects found")
	}

	var featuredProject Project

	for _, project := range projects {
		if project.Title == featuredProjectTitle {
			featuredProject = project
		}
	}

	if featuredProject.Title == "" {
		return nil, errors.New("no projects matched the featured project title")
	}

	featuredProject.Body = storage.RemoveHTMLTags(featuredProject.Body)

	return &featuredProject, nil
}

// ToCard converts a Project to a Card component for display in lists.
func (p Project) ToCard(i int) components.Card {
	return components.Card{
		Title:     p.Title,
		Subtitle:  p.Subtitle,
		Date:      "",
		Body:      p.Body,
		ImagePath: p.Image,
		Get:       "/project?id=" + p.ID,
		Tags:      p.Tags,
		Index:     i,
	}
}

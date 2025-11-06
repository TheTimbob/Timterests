package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"timterests/internal/storage"
	"timterests/internal/types"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/aws"
)

type Project struct {
	types.Document `yaml:",inline"`
	Repository     string `yaml:"repository"`
	Image          string `yaml:"image-path"`
}

func ProjectsPageHandler(w http.ResponseWriter, r *http.Request, storageInstance storage.Storage, currentTag, design string) {
	var component templ.Component
	var tags []string

	projects, err := ListProjects(storageInstance, currentTag)
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

	if r.Header.Get("HX-Request") == "true" {
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

func GetProjectHandler(w http.ResponseWriter, r *http.Request, storageInstance storage.Storage, projectID string) {
	projects, err := ListProjects(storageInstance, "all")
	if err != nil {
		http.Error(w, "Failed to fetch projects", http.StatusInternalServerError)
		return
	}

	for _, project := range projects {
		if project.ID == projectID {
			var component templ.Component
			isAuthenticated := IsAuthenticated(r)

			if r.Header.Get("HX-Request") == "true" {
				component = ProjectDisplay(project, isAuthenticated)
			} else {
				component = ProjectPage(project, isAuthenticated)
			}
			err = component.Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Printf("Error rendering in GetProjectsHandler: %e", err)
			}
		}
	}

}

func ListProjects(storageInstance storage.Storage, tag string) ([]Project, error) {
	var projects []Project

	// Get all projects from the storage
	prefix := "projects/"
	projectFiles, err := storage.ListObjects(context.Background(), storageInstance, prefix)
	if err != nil {
		return nil, err
	}

	for id, obj := range projectFiles {
		key := aws.ToString(obj.Key)

		if key == prefix {
			continue
		}

		project, err := GetProject(key, id, storageInstance)
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

func GetProject(key string, id int, storageInstance storage.Storage) (*Project, error) {
	var project Project
	project.ID = strconv.Itoa(id)
	project.S3Key = key
	err := storage.GetPreparedFile(key, &project, storageInstance)
	if err != nil {
		return nil, err
	}

	localImagePath, err := storage.GetImageFromS3(storageInstance, project.Image)
	if err != nil {
		log.Printf("Failed to download image: %v", err)
		return nil, err
	}
	project.Image = localImagePath

	return &project, nil
}

func GetFeaturedProject(storageInstance storage.Storage) (*Project, error) {

	projects, err := ListProjects(storageInstance, "all")
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("no projects found")
	}

	featuredProjectTitle := "Timterests"
	var featuredProject Project
	for _, project := range projects {
		if project.Title == featuredProjectTitle {
			featuredProject = project
		}
	}
	featuredProject.Body = storage.RemoveHTMLTags(featuredProject.Body)
	return &featuredProject, nil
}

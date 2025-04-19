package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
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

	if currentTag != "" || design != "" {
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
			component := ProjectPage(project)
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
	fileName := path.Base(key)
	localFilePath := path.Join("s3", fileName)

	file, err := storage.GetFile(key, localFilePath, storageInstance)
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		return nil, err
	}

	if err := storage.DecodeFile(file, &project); err != nil {
		log.Printf("Failed to decode file: %v", err)
		return nil, err
	}

	body, err := storage.BodyToHTML(project.Body)
	if err != nil {
		log.Printf("Failed to convert body to HTML: %v", err)
		return nil, err
	}

	localImagePath, err := storage.GetImageFromS3(storageInstance, project.Image)
	if err != nil {
		log.Printf("Failed to download image: %v", err)
		return nil, err
	}

	project.Image = localImagePath
	project.Body = body
	project.ID = strconv.Itoa(id)
	return &project, nil
}

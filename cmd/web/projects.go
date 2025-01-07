package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"timterests/internal/models"
	"timterests/internal/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func ProjectsListHandler(w http.ResponseWriter, r *http.Request, storageInstance models.Storage) {
	projects, err := ListProjects(storageInstance)
	if err != nil {
		message := "Failed to fetch projects"
		http.Error(w, fmt.Sprintf("%s: %v", message, err), http.StatusInternalServerError)
		return
	}

	for i := range projects {
		projects[i].Document.Body = storage.RemoveHTMLTags(projects[i].Body)
	}
	component := ProjectsList(projects)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Fatalf("Error rendering in ProjectPosts: %e", err)
	}
}

func GetProjectsHandler(w http.ResponseWriter, r *http.Request, storageInstance models.Storage, projectID string) {

	projects, err := ListProjects(storageInstance)
	if err != nil {
		http.Error(w, "Failed to fetch projects", http.StatusInternalServerError)
		return
	}

	for _, project := range projects {
		if project.Document.ID == projectID {
			component := ProjectPage(project)
			err = component.Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Fatalf("Error rendering in GetProjectsHandler: %e", err)
			}
		}
	}

}

func ListProjects(storageInstance models.Storage) ([]models.Project, error) {
	var projects []models.Project

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

		fileName := path.Base(key)
		localFilePath := path.Join("s3", fileName)

		project, err := storage.ReadFile[models.Project](key, localFilePath, storageInstance)
		if err != nil {
			log.Printf("Failed to read file: %v", err)
			return nil, err
		}

        localImagePath, err := ProjectImage(storageInstance, project.Image)
        if err != nil {
            log.Printf("Failed to download image: %v", err)
            return nil, err
        }

        project.Image = localImagePath
		project.Document.ID = strconv.Itoa(id)
		projects = append(projects, project)
	}

	return projects, nil
}

func ProjectImage(storageInstance models.Storage, imagePath string) (string, error) {
    localImagePath := path.Join("s3", filepath.Base(imagePath))
    err := storage.DownloadFile(context.Background(), storageInstance, imagePath, localImagePath)
    if err != nil {
		log.Println("Failed to download image: ", err)
		return localImagePath, err
	}

    localImagePath = path.Join("/", localImagePath)

    return localImagePath, nil
}
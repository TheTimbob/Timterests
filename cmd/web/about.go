package web

import (
	"net/http"
	"strings"

	apperrors "timterests/internal/errors"
	"timterests/internal/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/a-h/templ"
)

type Experience struct {
	Company     string `yaml:"company"`
	Role        string `yaml:"role"`
	StartDate   string `yaml:"startDate"`
	EndDate     string `yaml:"endDate"`
	Description string `yaml:"description"`
	Location    string `yaml:"location"`
}

type Education struct {
	Institution string `yaml:"institution"`
	Degree      string `yaml:"degree"`
	StartDate   string `yaml:"startDate"`
	EndDate     string `yaml:"endDate"`
	Description string `yaml:"description"`
	Location    string `yaml:"location"`
}

type Skill struct {
	Name        string   `yaml:"name"`
	Items       []string `yaml:"items"`
	Description string   `yaml:"description"`
}

type About struct {
	Title      string       `yaml:"title"`
	Subtitle   string       `yaml:"subtitle"`
	Body       string       `yaml:"-"`
	Name       string       `yaml:"name"`
	Specialty  string       `yaml:"specialty"`
	Location   string       `yaml:"location"`
	GitHub     string       `yaml:"github"`
	Email      string       `yaml:"email"`
	Experience []Experience `yaml:"experience"`
	Education  []Education  `yaml:"education"`
	Skills     []Skill      `yaml:"skills"`
}

func AboutHandler(w http.ResponseWriter, r *http.Request, s storage.Storage) {
	var about About

	prefix := "about/"

	aboutFile, err := s.ListObjects(r.Context(), prefix)
	if err != nil {
		HandleError(w, r, apperrors.StorageFailed(err), "AboutHandler", "listObjects")

		return
	}

	if len(aboutFile) == 0 {
		HandleError(w, r, apperrors.NotFound(nil), "AboutHandler", "findAbout")

		return
	}

	var key string

	for _, obj := range aboutFile {
		k := aws.ToString(obj.Key)
		if strings.HasSuffix(k, ".yaml") {
			key = k

			break
		}
	}

	if key == "" {
		HandleError(w, r, apperrors.NotFound(nil), "AboutHandler", "getKey")

		return
	}

	err = s.GetPreparedFile(r.Context(), key, &about)
	if err != nil {
		HandleError(w, r, apperrors.StorageFailed(err), "AboutHandler", "getPreparedFile")

		return
	}

	body, err := s.GetDocumentBody(r.Context(), key)
	if err != nil {
		HandleError(w, r, apperrors.StorageFailed(err), "AboutHandler", "getDocumentBody")

		return
	}

	about.Body = body
	about.GitHub = strings.TrimSpace(about.GitHub)
	about.Email = strings.TrimSpace(about.Email)

	var component templ.Component

	switch r.URL.Query().Get("tab") {
	case "bio":
		component = BioTab(about)
	case "education":
		component = EducationTab(about.Education)
	case "work":
		component = ExperienceTab(about.Experience)
	case "skills":
		component = SkillsTab(about.Skills)
	default:
		component = AboutForm(about)
	}

	err = renderHTML(w, r, http.StatusOK, component)
	if err != nil {
		HandleError(w, r, apperrors.RenderFailed(err), "AboutHandler", "render")
	}
}

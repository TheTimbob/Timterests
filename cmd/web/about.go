package web

import (
	"log"
	"net/http"
	"timterests/internal/storage"

	"github.com/a-h/templ"
)

type Experience struct {
	Company     string `yaml:"company"`
	Role        string `yaml:"role"`
	StartDate   string `yaml:"startDate"`
	EndDate     string `yaml:"endDate"`
	Description string `yaml:"description"`
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
	Name  string   `yaml:"name"`
	Items []string `yaml:"items"`
}

// About represents About page content.
type About struct {
	Title      string       `yaml:"title"`
	Name       string       `yaml:"name"`
	Location   string       `yaml:"location"`
	Focus      string       `yaml:"focus"`
	Github     string       `yaml:"github"`
	Email      string       `yaml:"email"`
	Body       string       `yaml:"body"`
	Experience []Experience `yaml:"experience"`
	Education  []Education  `yaml:"education"`
	Skills     []Skill      `yaml:"skills"`
}

// AboutHandler handles requests to the About page and serves its content.
func AboutHandler(w http.ResponseWriter, r *http.Request, s storage.Storage) {
	var about About

	prefix := "about/"

	aboutFile, err := s.ListObjects(r.Context(), prefix)
	if err != nil {
		http.Error(w, "Failed to fetch about info", http.StatusInternalServerError)

		return
	}

	key := *aboutFile[0].Key

	err = s.GetPreparedFile(r.Context(), key, &about)
	if err != nil {
		http.Error(w, "Failed to prepare about info", http.StatusInternalServerError)
		log.Printf("Error fetching about info: %v", err)

		return
	}

	// Check if this is a tab request
	tab := r.URL.Query().Get("tab")

	var component templ.Component

	switch tab {
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

	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error rendering in AboutHandler: %e", err)

		return
	}
}

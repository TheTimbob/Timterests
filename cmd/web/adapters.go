package web

import (
	"timterests/cmd/web/components"
	"timterests/internal/model"
)

// ArticleCard converts a model.Article to a Card component for display in lists.
func ArticleCard(a model.Article, i int) components.Card {
	return components.Card{
		Title:     a.Title,
		Subtitle:  a.Subtitle,
		Date:      a.Date,
		Preview:   a.Preview,
		ImagePath: "",
		Get:       "/article?id=" + a.ID,
		Tags:      a.Tags,
		Index:     i,
	}
}

// ProjectCard converts a model.Project to a Card component for display in lists.
func ProjectCard(p model.Project, i int) components.Card {
	return components.Card{
		Title:     p.Title,
		Subtitle:  p.Subtitle,
		Date:      p.Timespan(),
		Preview:   p.Preview,
		ImagePath: p.Image,
		Get:       "/project?id=" + p.ID,
		Tags:      p.Tags,
		Index:     i,
	}
}

// LetterCard converts a model.Letter to a Card component for display in lists.
func LetterCard(l model.Letter, i int) components.Card {
	return components.Card{
		Title:     l.Title,
		Subtitle:  l.Subtitle,
		Date:      l.Date,
		Preview:   l.Preview,
		ImagePath: "",
		Get:       "/letter?id=" + l.ID,
		Tags:      l.Tags,
		Index:     i,
	}
}

// BookCard converts a model.ReadingList to a Card component for display in lists.
func BookCard(r model.ReadingList, i int) components.Card {
	return components.Card{
		Title:     r.Title,
		Subtitle:  r.Subtitle,
		Date:      "",
		Preview:   r.Preview,
		ImagePath: r.Image,
		Get:       "/book?id=" + r.ID,
		Tags:      r.Tags,
		Index:     i,
	}
}

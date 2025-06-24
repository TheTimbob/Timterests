package web

import (
	"context"
	"fmt"
	"log"
	"path"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"timterests/internal/storage"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/aws"
)

type Item interface {
	Article | Letter | Project | ReadingList
	GetID() string
	GetBody() string
	GetTitle() string
	GetSubtitle() string
	GetTags() []string
}

func GetListPageComponent[T Item](storageInstance storage.Storage, currentTag, typeStr, design string) (templ.Component, error) {
	var component templ.Component
	var tags []string

	items, err := GetItemsList[T](storageInstance, currentTag, typeStr)
	if err != nil {
		return nil, err
	}

	for i := range items {
		body := storage.RemoveHTMLTags(items[i].GetBody())
		SetField(items[i], "Body", body)
		v := reflect.ValueOf(&items[i]).Elem()
		tags = storage.GetTags(v, tags)
	}

	title := GenerateTitle(typeStr)
	get := "/" + typeStr
	if currentTag != "" || design != "" {
		component = List(items, design, get)
	} else {
		component = ListPage(items, design, get, title)
	}

	return component, nil
}

func GetItemComponent[T Item](storageInstance storage.Storage, itemID, typeStr string, page bool) (templ.Component, error) {
	items, err := GetItemsList[T](storageInstance, "all", typeStr)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if item.GetID() == itemID {
			var component templ.Component
			if page {
				component = Page(item)
			} else {
				component = Display(item)
			}
			return component, nil
		}
	}
	return nil, fmt.Errorf("item with ID %s not found", itemID)
}

func GetItemsList[T Item](storageInstance storage.Storage, tag, typeStr string) ([]T, error) {
	var items []T
	prefix := typeStr + "s/"

	itemFiles, err := storage.ListObjects(context.Background(), storageInstance, prefix)
	if err != nil {
		return nil, err
	}

	for id, obj := range itemFiles {
		key := aws.ToString(obj.Key)
		if key == prefix {
			continue
		}

		item, err := GetItem[T](key, id, storageInstance)
		if err != nil {
			return nil, err
		}

		if slices.Contains(item.GetTags(), tag) || tag == "all" || tag == "" {
			items = append(items, item)
		}
	}

	return items, nil
}

func GetItem[T Item](key string, id int, storageInstance storage.Storage) (T, error) {
	var item T
	fileName := path.Base(key)
	localFilePath := path.Join("s3", fileName)

	file, err := storage.GetFile(key, localFilePath, storageInstance)
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		return item, err
	}

	if err := storage.DecodeFile(file, &item); err != nil {
		log.Printf("Failed to decode file: %v", err)
		return item, err
	}

	body, err := storage.BodyToHTML(item.GetBody())
	if err != nil {
		log.Printf("Failed to parse the body text into HTML: %v", err)
		return item, err
	}

	SetField(item, "Body", body)
	SetField(item, "ID", strconv.Itoa(id))
	return item, nil
}

func GenerateTitle(typeStr string) string {
	if len(typeStr) == 0 {
		return typeStr
	}
	firstChar := string(typeStr[0])
	remainingChars := typeStr[1:]
	return strings.ToUpper(firstChar) + remainingChars
}

func SetField[T Item](item T, fieldName, value string) {
	v := reflect.ValueOf(&item).Elem()
	field := v.FieldByName(fieldName)
	if field.IsValid() && field.CanSet() {
		field.SetString(value)
	}
}

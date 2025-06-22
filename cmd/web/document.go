package web

import (
	"context"
	"fmt"
	"log"
	"path"
	"reflect"
	"slices"
	"strconv"
	"timterests/internal/storage"
	"timterests/internal/types"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/aws"
)

func ListPageHandler[T types.DocumentItem](storageInstance storage.Storage, currentTag, design, prefix string) (templ.Component, error) {
	var component templ.Component
	var tags []string

	items, err := GetItemsList[T](storageInstance, currentTag, prefix)
	if err != nil {
		return nil, err
	}

	for i := range items {
		items[i].SetBody(storage.RemoveHTMLTags(items[i].GetBody()))
		v := reflect.ValueOf(items[i])
		tags = storage.GetTags(v, tags)
	}

	if currentTag != "" || design != "" {
		component = List(items, design, prefix)
	} else {
		component = ListPage(items, design, prefix, "Articles")
	}

	return component, nil
}

func ItemPageHandler[T types.DocumentItem](storageInstance storage.Storage, itemID, prefix string, page bool) (templ.Component, error) {
	items, err := GetItemsList[T](storageInstance, "all", prefix)
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

func GetItemsList[T types.DocumentItem](storageInstance storage.Storage, tag, prefix string) ([]T, error) {
	var items []T

	itemFiles, err := storage.ListObjects(context.Background(), storageInstance, prefix)
	if err != nil {
		return nil, err
	}

	for id, obj := range itemFiles {
		key := aws.ToString(obj.Key)
		if key == prefix {
			continue
		}

		i, err := GetItem[T](key, id, storageInstance)
		if err != nil {
			return nil, err
		}

		item := *i
		if slices.Contains(item.GetTags(), tag) || tag == "all" || tag == "" {
			items = append(items, item)
		}
	}

	return items, nil
}

func GetItem[T types.DocumentItem](key string, id int, storageInstance storage.Storage) (*T, error) {
	var item T
	fileName := path.Base(key)
	localFilePath := path.Join("s3", fileName)

	file, err := storage.GetFile(key, localFilePath, storageInstance)
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		return nil, err
	}

	if err := storage.DecodeFile(file, &item); err != nil {
		log.Printf("Failed to decode file: %v", err)
		return nil, err
	}

	body, err := storage.BodyToHTML(item.GetBody())
	if err != nil {
		log.Printf("Failed to parse the body text into HTML: %v", err)
		return nil, err
	}

	item.SetBody(body)
	item.SetID(strconv.Itoa(id))
	return &item, nil
}

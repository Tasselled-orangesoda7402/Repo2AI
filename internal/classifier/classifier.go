package classifier

import (
	"os"
	"path/filepath"
	"strings"
)

type Category string

const (
	CategoryController Category = "controllers"
	CategoryService    Category = "services"
	CategoryMapper     Category = "mappers"
	CategoryEntity     Category = "entities"
	CategorySQL        Category = "sql"
	CategoryConfig     Category = "configs"
	CategoryBuild      Category = "build"
	CategoryOther      Category = "others"
)

func Classify(path string) Category {
	normalizedPath := filepath.ToSlash(strings.ToLower(path))
	fileName := strings.ToLower(filepath.Base(path))

	if fileName == "pom.xml" {
		return CategoryBuild
	}

	if fileName == "application.yml" ||
		fileName == "application.yaml" ||
		fileName == "application.properties" {
		return CategoryConfig
	}

	if strings.HasSuffix(fileName, ".xml") &&
		(strings.Contains(normalizedPath, "/mapper/") ||
			strings.Contains(normalizedPath, "/mappers/") ||
			strings.Contains(normalizedPath, "/mapping/") ||
			strings.Contains(normalizedPath, "/mappings/")) {
		return CategorySQL
	}

	if strings.HasSuffix(fileName, ".java") {
		content, err := os.ReadFile(path)
		if err != nil {
			return CategoryOther
		}

		text := string(content)

		if containsAny(text, "@RestController", "@Controller") {
			return CategoryController
		}

		if containsAny(text, "@Service") {
			return CategoryService
		}

		if containsAny(text, "@Mapper", "@Repository") {
			return CategoryMapper
		}

		if containsAny(text, "@Entity") {
			return CategoryEntity
		}
	}

	return CategoryOther
}

func containsAny(text string, keywords ...string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fichil/Repo2AI/internal/classifier"
)

type Manifest struct {
	ProjectName  string     `json:"projectName"`
	RootPath     string     `json:"rootPath"`
	TotalFiles   int        `json:"totalFiles"`
	JavaFiles    int        `json:"javaFiles"`
	XmlFiles     int        `json:"xmlFiles"`
	IgnoredFiles int        `json:"ignoredFiles"`
	Files        []FileInfo `json:"files"`
}

type FileInfo struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Type     string `json:"type"`
	Category string `json:"category"`
}

func Scan(rootPath string) (*Manifest, error) {
	absRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	projectName := filepath.Base(absRootPath)

	manifest := &Manifest{
		ProjectName: projectName,
		RootPath:    absRootPath,
		Files:       make([]FileInfo, 0),
	}

	err = filepath.Walk(absRootPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relativePath, err := filepath.Rel(absRootPath, path)
		if err != nil {
			return err
		}

		relativePath = filepath.ToSlash(relativePath)

		if info.IsDir() {
			if shouldIgnore(relativePath, true) {
				if relativePath != "." {
					manifest.IgnoredFiles++
					return filepath.SkipDir
				}
			}
			return nil
		}

		if shouldIgnore(relativePath, false) {
			manifest.IgnoredFiles++
			return nil
		}

		fileType := detectType(relativePath)
		category := classifier.Classify(path)

		if fileType == "java" {
			manifest.JavaFiles++
		}

		if fileType == "xml" {
			manifest.XmlFiles++
		}

		manifest.Files = append(manifest.Files, FileInfo{
			Path:     relativePath,
			Size:     info.Size(),
			Type:     fileType,
			Category: string(category),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	manifest.TotalFiles = len(manifest.Files)

	return manifest, nil
}

func WriteManifest(manifest *Manifest, outputPath string) error {
	err := os.MkdirAll(outputPath, os.ModePerm)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	manifestPath := filepath.Join(outputPath, "manifest.json")

	err = os.WriteFile(manifestPath, data, 0644)
	if err != nil {
		return err
	}

	fmt.Println("Manifest generated:", filepath.ToSlash(manifestPath))
	return nil
}

func detectType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".java":
		return "java"
	case ".xml":
		return "xml"
	case ".md":
		return "markdown"
	case ".yml", ".yaml":
		return "yaml"
	case ".properties":
		return "properties"
	case ".json":
		return "json"
	case ".sql":
		return "sql"
	case ".go":
		return "go"
	case ".txt":
		return "text"
	default:
		return "other"
	}
}

func shouldIgnore(path string, isDir bool) bool {
	normalizedPath := filepath.ToSlash(path)
	lowerPath := strings.ToLower(normalizedPath)

	ignoreDirs := []string{
		".git",
		".idea",
		".vscode",
		"target",
		"build",
		"dist",
		"node_modules",
		"logs",
		"out",
		"output",
	}

	ignoreFiles := []string{
		".class",
		".jar",
		".war",
		".ear",
		".exe",
		".dll",
		".so",
		".dylib",
		".log",
		".zip",
		".tar",
		".gz",
		".rar",
		".7z",
	}

	if isDir {
		baseName := strings.ToLower(filepath.Base(lowerPath))
		for _, dir := range ignoreDirs {
			if baseName == dir {
				return true
			}
		}
		return false
	}

	for _, suffix := range ignoreFiles {
		if strings.HasSuffix(lowerPath, suffix) {
			return true
		}
	}

	for _, dir := range ignoreDirs {
		if strings.Contains(lowerPath, "/"+dir+"/") {
			return true
		}
	}

	return false
}

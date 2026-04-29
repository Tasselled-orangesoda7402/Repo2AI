package packer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fichil/Repo2AI/internal/scanner"
)

const outputDir = "output"

var maxPackSizeBytes = 10 * 1024 * 1024

func Generate(manifest *scanner.Manifest) error {
	if manifest == nil {
		return fmt.Errorf("manifest is nil")
	}

	groups := map[string][]scanner.FileInfo{}

	for _, file := range manifest.Files {
		category := file.Category
		if category == "" {
			category = "others"
		}
		groups[category] = append(groups[category], file)
	}

	err := writeProjectSummary(manifest, groups)
	if err != nil {
		return err
	}

	for category, files := range groups {
		if len(files) == 0 {
			continue
		}

		err := writeCategoryPack(manifest.RootPath, category, files)
		if err != nil {
			return err
		}
	}

	fmt.Println("AI Context Pack generated:", outputDir)
	return nil
}

func writeCategoryPack(rootPath string, category string, files []scanner.FileInfo) error {
	partNumber := 1

	var builder strings.Builder
	writePackHeader(&builder, category, partNumber)
	headerSize := builder.Len()

	for _, file := range files {
		block, err := buildFileBlock(rootPath, file)
		if err != nil {
			return err
		}

		if builder.Len() > headerSize && builder.Len()+len(block) > maxPackSizeBytes {
			err = writePackFile(category, partNumber, builder.String())
			if err != nil {
				return err
			}

			partNumber++
			builder.Reset()
			writePackHeader(&builder, category, partNumber)
			headerSize = builder.Len()
		}

		builder.WriteString(block)
	}

	if builder.Len() > headerSize {
		err := writePackFile(category, partNumber, builder.String())
		if err != nil {
			return err
		}
	}

	return nil
}

func languageByType(fileType string) string {
	switch fileType {
	case "java":
		return "java"
	case "xml":
		return "xml"
	case "json":
		return "json"
	case "yaml":
		return "yaml"
	case "properties":
		return "properties"
	case "sql":
		return "sql"
	case "go":
		return "go"
	case "markdown":
		return "markdown"
	default:
		return ""
	}
}

func writeProjectSummary(manifest *scanner.Manifest, groups map[string][]scanner.FileInfo) error {
	outputPath := filepath.Join(outputDir, "project-summary.md")

	var builder strings.Builder

	builder.WriteString("# Repo2AI Project Summary\n\n")

	builder.WriteString("## Project\n\n")
	builder.WriteString("- Project Name: ")
	builder.WriteString(manifest.ProjectName)
	builder.WriteString("\n")
	builder.WriteString("- Root Path: ")
	builder.WriteString(manifest.RootPath)
	builder.WriteString("\n\n")

	builder.WriteString("## Statistics\n\n")
	builder.WriteString(fmt.Sprintf("- Total Files: %d\n", manifest.TotalFiles))
	builder.WriteString(fmt.Sprintf("- Java Files: %d\n", manifest.JavaFiles))
	builder.WriteString(fmt.Sprintf("- XML Files: %d\n", manifest.XmlFiles))
	builder.WriteString(fmt.Sprintf("- Ignored Files: %d\n\n", manifest.IgnoredFiles))

	builder.WriteString("## Components\n\n")
	writeComponentCount(&builder, "Controllers", groups["controllers"])
	writeComponentCount(&builder, "Services", groups["services"])
	writeComponentCount(&builder, "Entities", groups["entities"])
	writeComponentCount(&builder, "Mappers", groups["mappers"])
	writeComponentCount(&builder, "SQL", groups["sql"])
	writeComponentCount(&builder, "Configs", groups["configs"])
	writeComponentCount(&builder, "Others", groups["others"])
	builder.WriteString("\n")

	builder.WriteString("## Generated Context Packs\n\n")
	writeGeneratedPack(&builder, "controllers", groups["controllers"])
	writeGeneratedPack(&builder, "services", groups["services"])
	writeGeneratedPack(&builder, "entities", groups["entities"])
	writeGeneratedPack(&builder, "mappers", groups["mappers"])
	writeGeneratedPack(&builder, "sql", groups["sql"])
	writeGeneratedPack(&builder, "configs", groups["configs"])
	writeGeneratedPack(&builder, "others", groups["others"])
	builder.WriteString("\n")

	builder.WriteString("## Recommended Reading Order\n\n")

	index := 1

	builder.WriteString(fmt.Sprintf("%d. project-summary.md\n", index))
	index++

	index = writeReadingOrder(&builder, index, "controllers", groups["controllers"])
	index = writeReadingOrder(&builder, index, "services", groups["services"])
	index = writeReadingOrder(&builder, index, "entities", groups["entities"])
	index = writeReadingOrder(&builder, index, "mappers", groups["mappers"])
	index = writeReadingOrder(&builder, index, "sql", groups["sql"])
	index = writeReadingOrder(&builder, index, "configs", groups["configs"])
	index = writeReadingOrder(&builder, index, "others", groups["others"])

	builder.WriteString("\n")

	builder.WriteString("## Files\n\n")
	for _, file := range manifest.Files {
		builder.WriteString("- ")
		builder.WriteString(file.Path)
		builder.WriteString(" [")
		builder.WriteString(file.Category)
		builder.WriteString("]\n")
	}

	err := os.WriteFile(outputPath, []byte(builder.String()), 0644)
	if err != nil {
		return err
	}

	fmt.Println("Project summary generated:", filepath.ToSlash(outputPath))
	return nil
}

func writeComponentCount(builder *strings.Builder, name string, files []scanner.FileInfo) {
	builder.WriteString(fmt.Sprintf("- %s: %d\n", name, len(files)))
}

func writeGeneratedPack(builder *strings.Builder, category string, files []scanner.FileInfo) {
	if len(files) == 0 {
		return
	}
	builder.WriteString("- ")
	builder.WriteString(category)
	builder.WriteString("_01.md\n")
}

func writeReadingOrder(builder *strings.Builder, index int, category string, files []scanner.FileInfo) int {
	if len(files) == 0 {
		return index
	}

	builder.WriteString(fmt.Sprintf("%d. %s_01.md\n", index, category))
	return index + 1
}

func writePackHeader(builder *strings.Builder, category string, partNumber int) {
	builder.WriteString("# Repo2AI Context Pack\n\n")
	builder.WriteString("Category: ")
	builder.WriteString(category)
	builder.WriteString("\n\n")
	builder.WriteString("Part: ")
	builder.WriteString(fmt.Sprintf("%02d", partNumber))
	builder.WriteString("\n\n")
}

func buildFileBlock(rootPath string, file scanner.FileInfo) (string, error) {
	fullPath := filepath.Join(rootPath, filepath.FromSlash(file.Path))

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("read file failed: %s, error: %w", file.Path, err)
	}

	var builder strings.Builder

	builder.WriteString("---\n\n")
	builder.WriteString("## ")
	builder.WriteString(file.Path)
	builder.WriteString("\n\n")

	builder.WriteString("```")
	builder.WriteString(languageByType(file.Type))
	builder.WriteString("\n")
	builder.Write(content)
	builder.WriteString("\n```\n\n")

	return builder.String(), nil
}

func writePackFile(category string, partNumber int, content string) error {
	fileName := fmt.Sprintf("%s_%02d.md", category, partNumber)
	outputPath := filepath.Join(outputDir, fileName)

	err := os.WriteFile(outputPath, []byte(content), 0644)
	if err != nil {
		return err
	}

	fmt.Println("Context pack generated:", filepath.ToSlash(outputPath))
	return nil
}

func SetMaxPackSize(size int) {
	if size > 0 {
		maxPackSizeBytes = size
	}
}

func CleanOutput() error {
	err := os.RemoveAll(outputDir)
	if err != nil {
		return err
	}

	return os.MkdirAll(outputDir, os.ModePerm)
}

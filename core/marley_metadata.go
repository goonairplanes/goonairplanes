package core

import (
	"regexp"
	"strings"
)


type PageMetadata struct {
	Title       string
	Description string
	MetaTags    map[string]string
	RenderMode  string
}

var metaTagRegex = regexp.MustCompile(`<!--meta:([a-zA-Z0-9_:,\-\s]+)-->`)
var renderModeRegex = regexp.MustCompile(`<!--render:([a-zA-Z]+)-->`)
var titleRegex = regexp.MustCompile(`<!--title:([^-]+)-->`)
var descRegex = regexp.MustCompile(`<!--description:([^-]+)-->`)

var htmlCommentMetaTagRegex = regexp.MustCompile(`<!---meta:([a-zA-Z0-9_:,\-\s]+)(?:-->|--->)`)
var htmlCommentRenderModeRegex = regexp.MustCompile(`<!---render:([a-zA-Z]+)(?:-->|--->)`)
var htmlCommentTitleRegex = regexp.MustCompile(`<!---title:([^-]+)(?:-->|--->)`)
var htmlCommentDescRegex = regexp.MustCompile(`<!---description:([^-]+)(?:-->|--->)`)


func extractPageMetadata(content, _ string) *PageMetadata {
	metadata := &PageMetadata{
		Title:       "Go on Airplanes",
		Description: AppConfig.DefaultMetaTags["description"],
		MetaTags:    make(map[string]string),
		RenderMode:  AppConfig.DefaultRenderMode,
	}

	for k, v := range AppConfig.DefaultMetaTags {
		metadata.MetaTags[k] = v
	}

	foundTitle := false
	foundDesc := false
	foundRenderMode := false

	htmlTitleMatch := htmlCommentTitleRegex.FindStringSubmatch(content)
	if len(htmlTitleMatch) > 1 {
		titleText := strings.TrimSpace(htmlTitleMatch[1])

		titleText = strings.TrimSuffix(titleText, "-")
		titleText = strings.TrimSpace(titleText)
		metadata.Title = titleText
		metadata.MetaTags["og:title"] = titleText
		foundTitle = true
	}

	htmlDescMatch := htmlCommentDescRegex.FindStringSubmatch(content)
	if len(htmlDescMatch) > 1 {
		descText := strings.TrimSpace(htmlDescMatch[1])

		descText = strings.TrimSuffix(descText, "-")
		descText = strings.TrimSpace(descText)
		metadata.Description = descText
		metadata.MetaTags["description"] = descText
		metadata.MetaTags["og:description"] = descText
		foundDesc = true
	}

	htmlMetaMatches := htmlCommentMetaTagRegex.FindAllStringSubmatch(content, -1)
	for _, match := range htmlMetaMatches {
		if len(match) > 1 {
			metaText := strings.TrimSpace(match[1])

			metaText = strings.TrimSuffix(metaText, "-")
			metaText = strings.TrimSpace(metaText)
			parts := strings.SplitN(metaText, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				metadata.MetaTags[key] = value
			}
		}
	}

	htmlRenderMatch := htmlCommentRenderModeRegex.FindStringSubmatch(content)
	if len(htmlRenderMatch) > 1 {
		renderText := strings.TrimSpace(htmlRenderMatch[1])

		renderText = strings.TrimSuffix(renderText, "-")
		renderText = strings.TrimSpace(renderText)
		mode := strings.ToLower(renderText)
		if mode == "ssg" {
			metadata.RenderMode = mode
			foundRenderMode = true
		}
	}

	if !foundTitle {
		titleMatch := titleRegex.FindStringSubmatch(content)
		if len(titleMatch) > 1 {
			metadata.Title = strings.TrimSpace(titleMatch[1])
			metadata.MetaTags["og:title"] = metadata.Title
		}
	}

	if !foundDesc {
		descMatch := descRegex.FindStringSubmatch(content)
		if len(descMatch) > 1 {
			metadata.Description = strings.TrimSpace(descMatch[1])
			metadata.MetaTags["description"] = metadata.Description
			metadata.MetaTags["og:description"] = metadata.Description
		}
	}

	metaMatches := metaTagRegex.FindAllStringSubmatch(content, -1)
	for _, match := range metaMatches {
		if len(match) > 1 {
			parts := strings.SplitN(match[1], ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				if _, exists := metadata.MetaTags[key]; !exists {
					metadata.MetaTags[key] = value
				}
			}
		}
	}

	if !foundRenderMode {
		renderMatch := renderModeRegex.FindStringSubmatch(content)
		if len(renderMatch) > 1 {
			mode := strings.ToLower(renderMatch[1])
			if mode == "ssg" {
				metadata.RenderMode = mode
			}
		}
	}

	return metadata
}


func processPageContent(content string, _ *PageMetadata) string {
	contentBefore := len(content)

	content = titleRegex.ReplaceAllString(content, "")
	content = descRegex.ReplaceAllString(content, "")
	content = metaTagRegex.ReplaceAllString(content, "")
	content = renderModeRegex.ReplaceAllString(content, "")

	content = htmlCommentTitleRegex.ReplaceAllString(content, "")
	content = htmlCommentDescRegex.ReplaceAllString(content, "")
	content = htmlCommentMetaTagRegex.ReplaceAllString(content, "")
	content = htmlCommentRenderModeRegex.ReplaceAllString(content, "")

	content = strings.TrimLeft(content, "\r\n")

	contentAfter := len(content)
	_ = contentBefore - contentAfter

	return content
}


func (m *Marley) mergeMetadata(routePath string, pageMetadata *PageMetadata) *PageMetadata {
	result := &PageMetadata{
		Title:       "Go on Airplanes",
		Description: AppConfig.DefaultMetaTags["description"],
		MetaTags:    make(map[string]string),
		RenderMode:  AppConfig.DefaultRenderMode,
	}

	for k, v := range AppConfig.DefaultMetaTags {
		if k != "description" && k != "og:description" && k != "og:title" {
			result.MetaTags[k] = v
		}
	}

	if m.LayoutMetadata != nil {
		if m.LayoutMetadata.Title != "Go on Airplanes" {
			result.Title = m.LayoutMetadata.Title
		}

		if m.LayoutMetadata.Description != AppConfig.DefaultMetaTags["description"] {
			result.Description = m.LayoutMetadata.Description
		}

		for k, v := range m.LayoutMetadata.MetaTags {
			if k != "description" && k != "og:description" && k != "og:title" {
				result.MetaTags[k] = v
			}
		}

		if m.LayoutMetadata.RenderMode != AppConfig.DefaultRenderMode {
			result.RenderMode = m.LayoutMetadata.RenderMode
		}
	}

	if pageMetadata.Title != "Go on Airplanes" {
		result.Title = pageMetadata.Title
	}

	if pageMetadata.Description != AppConfig.DefaultMetaTags["description"] {
		result.Description = pageMetadata.Description
	}

	for k, v := range pageMetadata.MetaTags {
		if k != "description" && k != "og:description" && k != "og:title" {
			result.MetaTags[k] = v
		}
	}

	if pageMetadata.RenderMode != AppConfig.DefaultRenderMode {
		result.RenderMode = pageMetadata.RenderMode
	}

	result.MetaTags["og:title"] = result.Title

	if result.Description != "" {
		result.MetaTags["description"] = result.Description
		result.MetaTags["og:description"] = result.Description
	}

	m.Logger.InfoLog.Printf("Merged metadata for %s: Title='%s', Description='%s...', Mode='%s'",
		routePath,
		result.Title,
		truncateString(result.Description, 30),
		result.RenderMode)

	return result
}

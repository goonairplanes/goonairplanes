package core

import (
	"strings"
)


func extractPageMetadataInternal(content, _ string) *PageMetadata {
	metadata := &PageMetadata{
		Title:       defaultTitle,
		Description: AppConfig.DefaultMetaTags["description"],
		MetaTags:    make(map[string]string, len(AppConfig.DefaultMetaTags)+4),
		RenderMode:  AppConfig.DefaultRenderMode,
		JSLibrary:   defaultJSLibrary,
	}

	for k, v := range AppConfig.DefaultMetaTags {
		metadata.MetaTags[k] = v
	}

	if match := htmlCommentTitleRegex.FindStringSubmatch(content); len(match) > 1 {
		title := strings.TrimSpace(match[1])
		title = strings.TrimSuffix(title, "-")
		title = strings.TrimSpace(title)
		metadata.Title = title
		metadata.MetaTags["og:title"] = title
	} else if match := titleRegex.FindStringSubmatch(content); len(match) > 1 {
		title := strings.TrimSpace(match[1])
		metadata.Title = title
		metadata.MetaTags["og:title"] = title
	}

	if match := htmlCommentDescRegex.FindStringSubmatch(content); len(match) > 1 {
		desc := strings.TrimSpace(match[1])
		desc = strings.TrimSuffix(desc, "-")
		desc = strings.TrimSpace(desc)
		metadata.Description = desc
		metadata.MetaTags["description"] = desc
		metadata.MetaTags["og:description"] = desc
	} else if match := descRegex.FindStringSubmatch(content); len(match) > 1 {
		desc := strings.TrimSpace(match[1])
		metadata.Description = desc
		metadata.MetaTags["description"] = desc
		metadata.MetaTags["og:description"] = desc
	}

	if match := htmlCommentRenderModeRegex.FindStringSubmatch(content); len(match) > 1 {
		mode := strings.ToLower(strings.TrimSpace(strings.TrimSuffix(match[1], "-")))
		if mode == "ssg" {
			metadata.RenderMode = mode
		}
	} else if match := renderModeRegex.FindStringSubmatch(content); len(match) > 1 {
		mode := strings.ToLower(match[1])
		if mode == "ssg" {
			metadata.RenderMode = mode
		}
	}

	for _, match := range htmlCommentMetaTagRegex.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			metaText := strings.TrimSpace(match[1])
			metaText = strings.TrimSuffix(metaText, "-")
			metaText = strings.TrimSpace(metaText)
			if idx := strings.IndexByte(metaText, ':'); idx > 0 {
				key := strings.TrimSpace(metaText[:idx])
				value := strings.TrimSpace(metaText[idx+1:])
				metadata.MetaTags[key] = value
			}
		}
	}

	for _, match := range metaTagRegex.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			metaText := match[1]
			if idx := strings.IndexByte(metaText, ':'); idx > 0 {
				key := strings.TrimSpace(metaText[:idx])
				value := strings.TrimSpace(metaText[idx+1:])
				if _, exists := metadata.MetaTags[key]; !exists {
					metadata.MetaTags[key] = value
				}
			}
		}
	}

	if match := htmlCommentJSLibraryRegex.FindStringSubmatch(content); len(match) > 1 {
		jsLibrary := strings.TrimSpace(match[1])
		jsLibrary = strings.ToLower(jsLibrary)

		switch jsLibrary {
		case "alpine", "jquery", "vanilla", "pvue":
			metadata.JSLibrary = jsLibrary
		}
	} else if match := jsLibraryRegex.FindStringSubmatch(content); len(match) > 1 {
		jsLibrary := strings.ToLower(strings.TrimSpace(match[1]))

		switch jsLibrary {
		case "alpine", "jquery", "vanilla", "pvue":
			metadata.JSLibrary = jsLibrary
		}
	}

	return metadata
}


func extractPageMetadata(content, filePath string) *PageMetadata {
	cacheKey := filePath

	if metadata, found := metadataCache.Get(cacheKey); found {
		return metadata
	}

	if len(content) < 1024 {
		metadata := extractPageMetadataInternal(content, filePath)
		metadataCache.Set(cacheKey, metadata)
		return metadata
	}

	resultChan := make(chan *PageMetadata, 1)
	metadataCache.extractC <- extractRequest{
		content:  content,
		filePath: filePath,
		result:   resultChan,
	}

	metadata := <-resultChan
	metadataCache.Set(cacheKey, metadata)

	return metadata
}


func processPageContent(content string, _ *PageMetadata) string {
	contentBefore := len(content)

	content = titleRegex.ReplaceAllString(content, "")
	content = descRegex.ReplaceAllString(content, "")
	content = metaTagRegex.ReplaceAllString(content, "")
	content = renderModeRegex.ReplaceAllString(content, "")
	content = jsLibraryRegex.ReplaceAllString(content, "")

	content = htmlCommentTitleRegex.ReplaceAllString(content, "")
	content = htmlCommentDescRegex.ReplaceAllString(content, "")
	content = htmlCommentMetaTagRegex.ReplaceAllString(content, "")
	content = htmlCommentRenderModeRegex.ReplaceAllString(content, "")
	content = htmlCommentJSLibraryRegex.ReplaceAllString(content, "")

	content = strings.TrimLeft(content, "\r\n")

	contentAfter := len(content)
	_ = contentBefore - contentAfter

	return content
}

package core


func (m *Marley) mergeMetadata(routePath string, pageMetadata *PageMetadata) *PageMetadata {
	cacheKey := "merge:" + routePath

	if metadata, found := metadataCache.Get(cacheKey); found {
		return metadata
	}

	result := &PageMetadata{
		Title:       defaultTitle,
		Description: AppConfig.DefaultMetaTags["description"],
		MetaTags:    make(map[string]string, len(AppConfig.DefaultMetaTags)+4),
		RenderMode:  AppConfig.DefaultRenderMode,
		JSLibrary:   defaultJSLibrary,
	}

	for k, v := range AppConfig.DefaultMetaTags {
		if k != "description" && k != "og:description" && k != "og:title" {
			result.MetaTags[k] = v
		}
	}

	if m.LayoutMetadata != nil {
		if m.LayoutMetadata.Title != defaultTitle {
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

		if m.LayoutMetadata.JSLibrary != defaultJSLibrary {
			result.JSLibrary = m.LayoutMetadata.JSLibrary
		}
	}

	if pageMetadata.Title != defaultTitle {
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

	if pageMetadata.JSLibrary != defaultJSLibrary {
		result.JSLibrary = pageMetadata.JSLibrary
	}

	result.MetaTags["og:title"] = result.Title

	if result.Description != "" {
		result.MetaTags["description"] = result.Description
		result.MetaTags["og:description"] = result.Description
	}

	if AppConfig.LogLevel == "debug" {
		m.Logger.InfoLog.Printf("Merged metadata for %s: Title='%s', Description='%s...', Mode='%s', JS='%s'",
			routePath,
			result.Title,
			truncateString(result.Description, 30),
			result.RenderMode,
			result.JSLibrary)
	}

	metadataCache.Set(cacheKey, result)

	return result
}

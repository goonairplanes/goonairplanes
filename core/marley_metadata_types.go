package core

import (
	"regexp"
)


type PageMetadata struct {
	Title       string
	Description string
	MetaTags    map[string]string
	RenderMode  string
	JSLibrary   string
}


var (
	
	metaTagRegex    = regexp.MustCompile(`<!--meta:([a-zA-Z0-9_:,\-\s]+)-->`)
	renderModeRegex = regexp.MustCompile(`<!--render:([a-zA-Z]+)-->`)
	titleRegex      = regexp.MustCompile(`<!--title:([^-]+)-->`)
	descRegex       = regexp.MustCompile(`<!--description:([^-]+)-->`)
	jsLibraryRegex  = regexp.MustCompile(`<!--js:\s*([a-zA-Z]+)\s*-->`)

	
	htmlCommentMetaTagRegex    = regexp.MustCompile(`<!---meta:([a-zA-Z0-9_:,\-\s]+)(?:-->|--->)`)
	htmlCommentRenderModeRegex = regexp.MustCompile(`<!---render:([a-zA-Z]+)(?:-->|--->)`)
	htmlCommentTitleRegex      = regexp.MustCompile(`<!---title:([^-]+)(?:-->|--->)`)
	htmlCommentDescRegex       = regexp.MustCompile(`<!---description:([^-]+)(?:-->|--->)`)
	htmlCommentJSLibraryRegex  = regexp.MustCompile(`<!---js:\s*([a-zA-Z]+)\s*(?:-->|--->)`)

	
	defaultTitle      = "Go on Airplanes"
	defaultRenderMode = "ssr"
	defaultJSLibrary  = "alpine"
)

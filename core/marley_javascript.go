package core

import (
	"fmt"
	"strings"
)


func injectJavaScriptLibraries(html, jsLibrary string) string {
	if jsLibrary == "vanilla" {
		if !AppConfig.DevMode {
			return html
		}
	}

	scriptContent, inMemory, cdnURL := GetJSLibraryContent(jsLibrary)

	var scriptTag string
	if jsLibrary == "alpine" {
		if inMemory {
			scriptTag = fmt.Sprintf("<script defer>%s</script>", scriptContent)
		} else {
			scriptTag = fmt.Sprintf("<script defer src=\"%s\"></script>", cdnURL)
		}
	} else if jsLibrary == "jquery" {
		if inMemory {
			scriptTag = fmt.Sprintf("<script>%s</script>", scriptContent)
		} else {
			scriptTag = fmt.Sprintf("<script src=\"%s\"></script>", cdnURL)
		}
	} else if jsLibrary == "pvue" {
		if inMemory {
			scriptTag = fmt.Sprintf("<script defer>%s</script>", scriptContent)
		} else {
			scriptTag = fmt.Sprintf("<script defer src=\"%s\"></script>", cdnURL)
		}
	}

	if AppConfig.DevMode {
		wsClientJS := GetWebSocketClientJS()
		if scriptTag != "" {
			scriptTag = scriptTag + "\n" + wsClientJS
		} else {
			scriptTag = wsClientJS
		}
	}

	if scriptTag != "" {
		return strings.Replace(html, "</head>", scriptTag+"\n</head>", 1)
	}

	return html
}

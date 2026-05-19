package openapi

import (
	"fmt"
	"strconv"
)

const scalarHTMLTemplate = `<!doctype html>
<html>
	<head>
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />
		<title>Netstamp API Docs</title>
		<style>
			body {
				margin: 0;
			}
		</style>
	</head>
	<body>
		<div id="app"></div>
		<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
		<script>
			Scalar.createApiReference("#app", {
				"url": %s,
				"theme": "elysiajs",
				"layout": "modern",
				"defaultOpenAllTags": true,
				"expandAllModelSections": true,
				"expandAllResponses": true,
				"hideClientButton": false,
				"showSidebar": true,
				"showDeveloperTools": "localhost",
				"showToolbar": "localhost",
				"operationTitleSource": "summary",
				"persistAuth": false,
				"telemetry": true,
				"externalUrls": {
					"dashboardUrl": "https://dashboard.scalar.com",
					"registryUrl": "https://registry.scalar.com",
					"proxyUrl": "https://proxy.scalar.com",
					"apiBaseUrl": "https://api.scalar.com"
				},
				"default": false,
				"isEditable": false,
				"isLoading": false,
				"hideModels": false,
				"documentDownloadType": "both",
				"hideTestRequestButton": false,
				"hideSearch": false,
				"showOperationId": false,
				"hideDarkModeToggle": false,
				"withDefaultFonts": true,
				"defaultOpenFirstTag": true,
				"orderSchemaPropertiesBy": "alpha",
				"orderRequiredPropertiesFirst": true,
				"slug": "api-1",
				"title": "API #1"
			});
		</script>
	</body>
</html>
`

func ScalarHTML(openAPIURL string) []byte {
	return []byte(fmt.Sprintf(scalarHTMLTemplate, strconv.Quote(openAPIURL)))
}

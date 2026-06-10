package handler

// scalarHTML returns the HTML page that loads the Scalar API reference UI
// pointed at the swagger spec served by gin-swagger at /swagger/doc.json.
func scalarHTML() []byte {
	return []byte(`<!doctype html>
<html>
  <head>
    <title>FinanceOS API — Docs</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>
      body { margin: 0; padding: 0; }
    </style>
  </head>
  <body>
    <script
      id="api-reference"
      data-url="/swagger/doc.json"
      data-configuration='{
        "theme": "purple",
        "layout": "modern",
        "defaultHttpClient": {
          "targetKey": "shell",
          "clientKey": "curl"
        },
        "metaData": {
          "title": "FinanceOS API"
        }
      }'
    ></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`)
}

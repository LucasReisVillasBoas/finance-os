package entity

// ImportResult holds the result of an import operation.
type ImportResult struct {
	Imported   int      `json:"imported"`
	Duplicates int      `json:"duplicates"`
	Errors     int      `json:"errors"`
	Messages   []string `json:"messages"`
}

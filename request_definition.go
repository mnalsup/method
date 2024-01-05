package method

type FileDefinition struct {
	RequestBodyPath string `yaml:"requestBodyPath"`
	FilePath        string `yaml:"filePath"`
}

type RequestDefinition struct {
	Method  string            `yaml:"method"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
	BodyStr string            `yaml:"bodyStr"`
	Body    interface{}       `yaml:"body"`
	Files   []FileDefinition  `yaml:"files"`
}

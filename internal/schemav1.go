package internal

type RequestSchema struct {
	FileName                          string
	RequestDefinition         `yaml:",inline"`
	AuthenticationHook `yaml:"authenticationHook"`
}

type RequestDefinition struct {
	Method  string            `yaml:"method"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
	BodyStr string            `yaml:"bodyStr"`
	Body    interface{}       `yaml:"body"`
	Files   []FileDefinition  `yaml:"files"`
}

type FileDefinition struct {
	RequestBodyPath string `yaml:"requestBodyPath"`
	FilePath        string `yaml:"filePath"`
}

type AuthenticationHook struct {
	Triggers            []AuthenticationTrigger  `yaml:"triggers"`
	RequestPath         string                   `yaml:"requestPath"`
	JsonParseBodyPath   string                   `yaml:"jsonParseBodyPath"`
	BearerToken         *BearerToken             `yaml:"bearerToken"`
	AuthHeader          *AuthHeader              `yaml:"authHeader"`
	EnvironmentVariable *AuthEnvironmentVariable `yaml:"environmentVariable"`
}

type AuthHeader struct {
	Header       string `yaml:"header"`
	FormatString string `yaml:"formatString"`
}

type AuthEnvironmentVariable struct {
	Variable string `yaml:"variable"`
}

type BearerToken struct{}

type AuthenticationTrigger struct {
	OnHttpStatus []int        `yaml:"onHttpStatus"`
	OnJsonValue  *OnJsonValue `yaml:"onJsonValue"`
}

type OnJsonValue struct {
	Path  string `yaml:"path"`
	Value string `yaml:"value"`
}

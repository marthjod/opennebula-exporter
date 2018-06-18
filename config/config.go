package config

type exporterConfig struct {
	Namespace string `yaml:"namespace"`
}

type apiConfig struct {
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	Endpoint    string `yaml:"endpoint"`
	InsecureSSL bool   `yaml:"insecure_ssl"`
}

// VMNameRegexpLabel represents a label name/VM name-oriented regexp combination.
type VMNameRegexpLabel struct {
	Name   string `yaml:"name"`
	Regexp string `yaml:"regexp"`
}

// UserTemplateLabel represents a label name/VM user template field combination.
type UserTemplateLabel struct {
	Name          string `yaml:"name"`
	TemplateField string `yaml:"field"`
}

// Config represents a complete configuration set.
type Config struct {
	Exporter           exporterConfig      `yaml:"exporter"`
	API                apiConfig           `yaml:"api"`
	VMNameRegexpLabels []VMNameRegexpLabel `yaml:"labels_vm_name"`
	UserTemplateLabels []UserTemplateLabel `yaml:"labels_user_template"`
}

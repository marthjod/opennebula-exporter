package config

type exporterConfig struct {
	Namespace     string `yaml:"namespace"`
	ListenAddress string `yaml:"listen_address"`
	MetricsPath   string `yaml:"metrics_path"`
	WriteStdout   bool   `yaml:"write_stdout"`
}

type apiConfig struct {
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	Endpoint    string `yaml:"endpoint"`
	InsecureSSL bool   `yaml:"insecure_ssl"`
}

type VMNameRegexpLabel struct {
	Name   string `yaml:"name"`
	Regexp string `yaml:"regexp"`
}

type UserTemplateLabel struct {
	Name          string `yaml:"name"`
	TemplateField string `yaml:"field"`
}

type Config struct {
	Exporter           exporterConfig      `yaml:"exporter"`
	API                apiConfig           `yaml:"api"`
	VMNameRegexpLabels []VMNameRegexpLabel `yaml:"labels_vm_name"`
	UserTemplateLabels []UserTemplateLabel `yaml:"labels_user_template"`
}

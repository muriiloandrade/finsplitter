package config

type Config struct {
	App Application
	Env Environment
}

type Application struct {
	Name    string
	Version string
}

type Environment struct {
	Name      string
	LogFormat string
}

package config

type Config struct {
	ProjectName string `koanf:"name" validate:"omitnil,max=255"`
}

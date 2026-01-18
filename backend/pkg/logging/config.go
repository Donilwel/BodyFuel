package logging

type Config struct {
	Level                 string `yaml:"level" env:"LOG_LEVEL" env-default:"debug"`
	AppID                 int    `yaml:"app_id" env:"APP_ID"`
	DeploymentEnvironment string `yaml:"deployment_environment" env:"DEPLOYMENT_ENVIRONMENT"`
	SageSystem            string `yaml:"system" env:"APP_NAME"`
	Namespace             string `yaml:"namespace" env:"NAMESPACE"`
	StandType             string `yaml:"stand_type" env:"STAND_TYPE"`
	PodName               string `yaml:"pod_name" env:"POD_NAME"`
}

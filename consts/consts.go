package consts

const (
	ConfigDir                       string = "./config"
	ServerYamlPath                  string = ConfigDir + "/server.yaml"
	LoggerYamlPath                  string = ConfigDir + "/logger.yaml"
	MgoYamlPath                     string = ConfigDir + "/mgo.yaml"
	TimeFormatLayout                string = "2006-01-02 15:04:05.000"
	ErrorMessageInternalServerError string = "An internal server error occurred."
)

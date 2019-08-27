package dockercreds

type DockerCreds map[string]entry

type entry struct {
	Auth string `json:"auth"`
}

func (c DockerCreds) append(dockerConfigJsonPath string ) error {

}

func parseDockerCfg(path string) (DockerCreds, error) {

}

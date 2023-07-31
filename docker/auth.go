package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

type DockerAuth struct {
	Auth string `json:"auth"`
}

type DockerAuths struct {
	Auths map[string]DockerAuth `json:"auths"`
}

type UsernamePassword struct {
	UserName string
	Password string
}

func (up *UsernamePassword) ToDockerAuth() DockerAuth {
	return DockerAuth{
		Auth: base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", up.UserName, up.Password))),
	}
}

func LoadAuthConfigFromCICDVar(jsonStringOrPath string) *DockerAuths {

	var jsonString string

	if _, err := os.Stat(jsonStringOrPath); err == nil {
		jsonStringBytes, err := os.ReadFile(jsonStringOrPath) // #nosec G304
		if err != nil {
			panic(fmt.Errorf("unexpected error occurred while trying to read the file path (from DOCKER_AUTH_CONFIG) ('%w')", err))
		}
		jsonString = string(jsonStringBytes)
	} else {
		jsonString = jsonStringOrPath
	}

	dockerAuths := DockerAuths{}
	err := json.Unmarshal([]byte(jsonString), &dockerAuths)
	if err != nil {
		panic(fmt.Errorf("unexpected error occurred while trying to parse the DOCKER_AUTH_CONFIG json env var ('%w')", err))
	}
	return &dockerAuths
}

func CreateAuth(authMap map[string]UsernamePassword) string {
	authsMap := make(map[string]DockerAuth)
	for registry, userpass := range authMap {
		patchedRegistry := registry
		if patchedRegistry == "index.docker.io" {
			patchedRegistry = "https://index.docker.io/v1/" // docker central registry compatability quirks
		}
		authsMap[patchedRegistry] = userpass.ToDockerAuth()
	}
	auths := DockerAuths{
		Auths: authsMap,
	}
	bytes, err := json.Marshal(auths)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func (da *DockerAuths) ToJsonString() string {
	jsonBytes, err := json.Marshal(da)
	if err != nil {
		panic(fmt.Errorf("unexpected error occurred while marshalling the docker auth to json: '%w'", err))
	}
	return string(jsonBytes)
}

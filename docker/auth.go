package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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

func CreateAuth(authMap map[string]UsernamePassword) string {
	authsMap := make(map[string]DockerAuth)
	for registry, userpass := range authMap {
		patchedRegistry := registry
		if patchedRegistry == "index.docker.io" {
			patchedRegistry = "https://index.docker.io/v1/" // docker central registry compatability quirks
		}
		authsMap[patchedRegistry] = DockerAuth{
			Auth: base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", userpass.UserName, userpass.Password))),
		}
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

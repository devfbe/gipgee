package docker

import "testing"

func TestDockerAuthGeneration(t *testing.T) {
	authMap := map[string]UsernamePassword{
		"foo": {
			"testuser",
			"testpassword",
		},
	}

	expected := `{"auths":{"foo":{"auth":"dGVzdHVzZXI6dGVzdHBhc3N3b3Jk"}}}`
	given := CreateAuth(authMap)
	if expected != given {
		t.Errorf("auth expected: '%v' auth given: '%v'", expected, given)
	}
}

package pipelinemodel

import "testing"

func TestContainerImageCoordinatesParsing(t *testing.T) {
	c1, err1 := ContainerImageCoordinatesFromString("docker.io/foo/bar:latest")
	c2, err2 := ContainerImageCoordinatesFromString("docker.io/foo/bar")
	c3, err3 := ContainerImageCoordinatesFromString("docker.io:443/foobar:foobar")
	c4, err4 := ContainerImageCoordinatesFromString("docker.io:443/foobar")
	for idx, e := range []error{err1, err2, err3, err4} {
		if e != nil {
			t.Errorf("Got error while parsing valid container image coordinates c%d: err: %s", idx, e)
		}
	}

	for _, coordinate := range []*ContainerImageCoordinates{c1, c2} {

		if coordinate.Registry != "docker.io" {
			t.Errorf("Expected registry is docker.io, parsed registry is: %s", coordinate.Registry)
		}

		if coordinate.Repository != "foo/bar" {
			t.Errorf("Expected repository is foo/bar, parsed repository is: %s", coordinate.Repository)
		}

		if coordinate.Tag != "latest" {
			t.Errorf("Expected tag is latest, parsed tag is: %s", coordinate.Tag)
		}
	}

	for _, coordinate := range []*ContainerImageCoordinates{c3, c4} {
		if coordinate.Registry != "docker.io:443" {
			t.Errorf("Expected registry is docker.io:443, parsed registry is: %s", coordinate.Registry)
		}

		if coordinate.Repository != "foobar" {
			t.Errorf("Expected repository is foo/bar, parsed repository is: %s", coordinate.Repository)
		}
	}

	if c3.Tag != "foobar" {
		t.Errorf("Expected tag is foobar, parsed tag is: %s", c3.Tag)
	}
	if c4.Tag != "latest" {
		t.Errorf("Expected tag is latest, parsed tag is: %s", c4.Tag)
	}

	c5, err := ContainerImageCoordinatesFromString("foobar")
	if c5 != nil {
		t.Error("unexpected non nil value for c5")
	}
	if err == nil {
		t.Error("expected error but error is nil")
	} else {
		expectedMsg := `didn't find / in container coordinates 'foobar' - cannot extract registry. Given coordinates must contain at least registry and repository`
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s' but got '%s'", expectedMsg, err.Error())
		}
	}

}

package config

import (
	"testing"

	yaml "gopkg.in/yaml.v3"
)

func TestConfig(t *testing.T) {
	testYaml := `
version: 1
images:
  hanswurst:
  hurzwanzt:
`
	config := Config{}
	err := yaml.Unmarshal([]byte(testYaml), &config)
	if err != nil {
		t.Fatal(err)
	}

	if config.Defaults != nil {
		t.Errorf("defaults it not nil")
	}

	if config.Version != 1 {
		t.Errorf("Expected version is != 1")
	}

}

func assertStringEquals(given string, expected string, t *testing.T) {
	if given != expected {
		t.Errorf("Given '%v' doesn't match expected '%v'", given, expected)
	}
}

func assertIntEquals(given int, expected int, t *testing.T) {
	if given != expected {
		t.Errorf("Given '%v' doesn't match expected '%v'", given, expected)
	}
}

func stringSliceEquals(given, expected []string, t *testing.T) {
	if len(given) != len(expected) {
		t.Errorf("Lenght of given slice (%v) doesn't equal length of expected slice (%v)", len(given), len(expected))
	}
	for i := range given {
		if given[i] != expected[i] {
			t.Errorf("Element at index %v doesn't match (elem in given: '%v', elem in expected: '%v')", i, given[i], expected[i])
		}
	}
}

func TestVersion(t *testing.T) {
	c, err := LoadConfiguration("testconfig.yml")

	if err != nil {
		panic(err)
	}

	assertIntEquals(c.Version, []int{1}[0], t)
}

func TestDefaults(t *testing.T) {
	c, err := LoadConfiguration("testconfig.yml")

	if err != nil {
		panic(err)
	}

	assertStringEquals(*c.Defaults.DefaultStagingRegistry, "staging.example.com", t)
	assertStringEquals(*c.Defaults.DefaultReleaseRegistry, "release.example.com", t)
	assertStringEquals(*c.Defaults.DefaultBaseImageRegistry, "baseImages.example.com", t)
	assertStringEquals(*c.Defaults.DefaultContainerFile, "Containerfile", t)
	stringSliceEquals(*c.Defaults.DefaultUpdateCheckCommand, []string{"test", "updates"}, t)
	stringSliceEquals(*c.Defaults.DefaultTestCommand, []string{"test.sh"}, t)
	stringSliceEquals(*c.Defaults.DefaultAssetsToWatch, []string{"test-assets/*"}, t)
	assertStringEquals(*c.Defaults.DefaultBaseImage.Registry, "thebaseimageregistry.example.com", t)
	assertStringEquals(*c.Defaults.DefaultBaseImage.Repository, "thebaseimage", t)
	assertStringEquals(*c.Defaults.DefaultBaseImage.Tag, "latest", t)
	assertIntEquals(len(*c.Defaults.DefaultBuildArgs), 2, t)
	assertStringEquals((*c.Defaults.DefaultBuildArgs)[0].Key, "default-build-arg-a", t)
	assertStringEquals((*c.Defaults.DefaultBuildArgs)[0].Value, "default-build-arg-value-a", t)
	assertStringEquals((*c.Defaults.DefaultBuildArgs)[1].Key, "default-build-arg-b", t)
	assertStringEquals((*c.Defaults.DefaultBuildArgs)[1].Value, "default-build-arg-value-b", t)
}
func TestImageConfigWithNoDefaults(t *testing.T) {
	c, err := LoadConfiguration("testconfig.yml")

	if err != nil {
		panic(err)
	}

	if image, exists := c.Images["imageWithoutDefaults"]; exists {
		assertStringEquals(*image.ContainerFile, "Containerfile.withoutDefaults", t)

		assertStringEquals(*image.StagingLocation.Tag, "nodefaultstagingtag", t)
		assertStringEquals(*image.StagingLocation.Registry, "nodefaultstagingregistry.example.com", t)
		assertStringEquals(*image.StagingLocation.Repository, "nodefaultstagingimage", t)

		assertIntEquals(len(image.ReleaseLocations), 2, t)
		assertStringEquals(*image.ReleaseLocations[0].Registry, "nodefaultregistry-a.example.com", t)
		assertStringEquals(*image.ReleaseLocations[0].Repository, "nodefaultimage-a", t)
		assertStringEquals(*image.ReleaseLocations[0].Tag, "nodefaulttag-a", t)

		assertStringEquals(*image.ReleaseLocations[1].Registry, "nodefaultregistry-b.example.com", t)
		assertStringEquals(*image.ReleaseLocations[1].Repository, "nodefaultimage-b", t)
		assertStringEquals(*image.ReleaseLocations[1].Tag, "nodefaulttag-b", t)

		assertStringEquals(*image.BaseImage.Registry, "nodefaultregistry-base.example.com", t)
		assertStringEquals(*image.BaseImage.Repository, "nodefaultbaseimage", t)
		assertStringEquals(*image.BaseImage.Tag, "nodefaulttag", t)

		assertIntEquals(len(*image.UpdateCheckCommand), 2, t)
		assertStringEquals((*image.UpdateCheckCommand)[0], "nonDefaultUpdateCheckCommand.sh", t)
		assertStringEquals((*image.UpdateCheckCommand)[1], "updatecheck", t)

		assertIntEquals(len(*image.TestCommand), 2, t)
		assertStringEquals((*image.TestCommand)[0], "test-image.sh", t)
		assertStringEquals((*image.TestCommand)[1], "imageWithoutDefaults", t)

		assertIntEquals(len(*image.AssetsToWatch), 2, t)
		assertStringEquals((*image.AssetsToWatch)[0], "test-assets/*", t)
		assertStringEquals((*image.AssetsToWatch)[1], "build-assets/nodefaultimage/*", t)

		assertIntEquals(len(*image.BuildArgs), 2, t)
		assertStringEquals((*image.BuildArgs)[0].Key, "nondefault-build-arg-a", t)
		assertStringEquals((*image.BuildArgs)[0].Value, "nondefault-build-arg-value-a", t)
		assertStringEquals((*image.BuildArgs)[1].Key, "nondefault-build-arg-b", t)
		assertStringEquals((*image.BuildArgs)[1].Value, "nondefault-build-arg-value-b", t)
	} else {
		t.Errorf("imageWithoutDefaults does not exist, but is expected to exist")
	}
}

func TestImageMinimalConfigWithAllDefaults(t *testing.T) {
	c, err := LoadConfiguration("testconfig.yml")

	if err != nil {
		panic(err)
	}

	if image, exists := c.Images["imageWithDefaults"]; exists {
		assertStringEquals(*image.ContainerFile, "Containerfile", t)

		if image.StagingLocation != nil {
			assertStringEquals(*image.StagingLocation.Registry, "staging.example.com", t)
		} else {
			t.Fatal("staging location exists in imageWithDefaults but is not defined in model")
		}

		// FIXME TAG and REPOSITORY from env vars
		if image.StagingLocation.Repository != nil {
			t.Fatalf("staging repository is not nil for imageWithDefaults")
		}
		if image.StagingLocation.Tag != nil {
			t.Fatalf("staging tag is not nil for imageWithDefaults")
		}

		assertIntEquals(len(image.ReleaseLocations), 1, t)
		assertStringEquals(*image.ReleaseLocations[0].Repository, "imageWithDefault", t)
		assertStringEquals(*image.ReleaseLocations[0].Tag, "latest", t)
		assertStringEquals(*image.ReleaseLocations[0].Registry, "release.example.com", t)

		assertStringEquals(*image.BaseImage.Registry, "thebaseimageregistry.example.com", t)
		assertStringEquals(*image.BaseImage.Repository, "thebaseimage", t)
		assertStringEquals(*image.BaseImage.Tag, "latest", t)

		assertIntEquals(len(*image.UpdateCheckCommand), 2, t)
		assertStringEquals((*image.UpdateCheckCommand)[0], "test", t)
		assertStringEquals((*image.UpdateCheckCommand)[1], "updates", t)

		assertIntEquals(len(*image.TestCommand), 1, t)
		assertStringEquals((*image.TestCommand)[0], "test.sh", t)

		assertIntEquals(len(*image.AssetsToWatch), 1, t)
		assertStringEquals((*image.AssetsToWatch)[0], "test-assets/*", t)

		assertIntEquals(len(*image.BuildArgs), 2, t)
		assertStringEquals((*image.BuildArgs)[0].Key, "default-build-arg-a", t)
		assertStringEquals((*image.BuildArgs)[0].Value, "default-build-arg-value-a", t)
		assertStringEquals((*image.BuildArgs)[1].Key, "default-build-arg-b", t)
		assertStringEquals((*image.BuildArgs)[1].Value, "default-build-arg-value-b", t)
	} else {
		t.Errorf("imageWithoutDefaults does not exist, but is expected to exist")
	}
}

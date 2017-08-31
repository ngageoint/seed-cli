package dockerhub

import (
	"fmt"
	"net/http"
	"strings"
)

//DockerHubRegistry type representing a Docker Hub registry
type DockerHubRegistry struct {
	URL    string
	Client *http.Client
}

//New creates a new docker hub registry from the given URL
func New(registryUrl string) (*DockerHubRegistry, error) {
	url := strings.TrimSuffix(registryUrl, "/")
	registry := &DockerHubRegistry{
		URL:    url,
		Client: &http.Client{},
	}

	return registry, nil
}

func (r *DockerHubRegistry) url(pathTemplate string, args ...interface{}) string {
	pathSuffix := fmt.Sprintf(pathTemplate, args...)
	url := fmt.Sprintf("%s%s", r.URL, pathSuffix)
	return url
}

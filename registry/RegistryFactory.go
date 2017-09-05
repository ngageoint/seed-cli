package registry

import (
	"github.com/heroku/docker-registry-client/registry"

	"strings"

	"github.com/ngageoint/seed-cli/registry/v2"
	"github.com/ngageoint/seed-cli/registry/dockerhub"
	"github.com/ngageoint/seed-cli/registry/containeryard"
)

type RepositoryRegistry interface {
	Name() string
	Ping() error
	Repositories(org string) ([]string, error)
	Tags(repository, org string) ([]string, error)
	Images(org string) ([]string, error)
}

type RepoRegistryFactory func(url, username, password string) (RepositoryRegistry, error)

func NewV2Registry(url, username, password string) (RepositoryRegistry, error) {
	v2, err := registry.New(url, username, password)
	if err != nil {
		if strings.Contains(url, "https://") {
			httpFallback := strings.Replace(url, "https://", "http://", 1)
			v2, err = registry.New(httpFallback, username, password)
		}
	}

	return &v2registry{v2}, err
}

func NewDockerHubRegistry(url, username, password string) (RepositoryRegistry, error) {
	hub, err := dockerhub.New(url)
	if err != nil {
		if strings.Contains(url, "https://") {
			httpFallback := strings.Replace(url, "https://", "http://", 1)
			hub, err = dockerhub.New(httpFallback)
		}
	}

	return hub, err
}

func NewContainerYardRegistry(url, username, password string) (RepositoryRegistry, error) {
	yard, err := containeryard.New(url)
	if err != nil {
		if strings.Contains(url, "https://") {
			httpFallback := strings.Replace(url, "https://", "http://", 1)
			yard, err = containeryard.New(httpFallback)
		}
	}

	return yard, err
}

func CreateRegistry(url, username, password string) (RepositoryRegistry, error) {
	v2, err1 := NewV2Registry(url, username, password)
	if v2 != nil && v2.Ping() == nil {
		return v2, nil
	}

	hub, err2 := NewDockerHubRegistry(url, username, password)
	if hub != nil && hub.Ping() == nil {
		return hub, nil
	}

	yard, err3 := NewContainerYardRegistry(url, username, password)
	if yard != nil && yard.Ping() == nil {
		return yard, nil
	}

	return nil, err1
}


package v2

import (
	"github.com/heroku/docker-registry-client/registry"
	"fmt"
	"os"
	"strings"
)

type v2registry struct {
	r *registry.Registry
}

func (r *v2registry) Name() string {
	return "V2"
}

func (r *v2registry) Ping() error {
	return r.Ping()
}

func (r *v2registry) Repositories(org string) ([]string, error) {
	return r.r.Repositories()
}

func (r *v2registry) Tags(repository, org string) ([]string, error) {
	return r.r.Tags(repository)
}

func (r *v2registry) Images(org string) ([]string, error) {
	repositories, err := r.r.Repositories()

	var images []string
	for _, repo := range repositories {
		if !strings.HasSuffix(repo, "-seed") {
			continue
		}
		tags, err := r.Tags(repo)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			continue
		}
		for _, tag := range tags {
			images = append(images, repo+":"+tag)
		}
	}

	return images, err
}

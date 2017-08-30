package dockerhub

type repositoriesResponse struct {
	Count    int
	Next     string
	Previous string
	Results  []Result
}

//Result struct representing JSON result
type Result struct {
	Name string
}

//UserRepositories Returns repositories for the given user
func (registry *DockerHubRegistry) UserRepositories(user string) ([]string, error) {
	url := registry.url("/v2/repositories/%s/", user)
	repos := make([]string, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response repositoriesResponse
	for err == nil {
		response.Next = ""
		url, err = registry.getDockerHubPaginatedJson(url, &response)
		for _, r := range response.Results {
			// Add all tags if found
			if rs, _ := registry.UserRepositoriesTags(user, r.Name); len(rs) > 0 {
				repos = append(repos, rs...)

				// No tags found - so just add the repo name
			} else {
				repos = append(repos, r.Name)
			}
		}
	}
	if err != ErrNoMorePages {
		return nil, err
	}
	return repos, nil
}

//UserRepositoriesTags Returns repositories along with their tags
func (registry *DockerHubRegistry) UserRepositoriesTags(user string, repository string) ([]string, error) {
	url := registry.url("/v2/repositories/%s/%s/tags", user, repository)
	repos := make([]string, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response repositoriesResponse
	for err == nil {
		response.Next = ""
		url, err = registry.getDockerHubPaginatedJson(url, &response)
		for _, r := range response.Results {
			repos = append(repos, repository+":"+r.Name)
		}
	}
	if err != ErrNoMorePages {
		return nil, err
	}
	return repos, nil
}

package containeryard

type Results struct {
	Community map[string]Image
	Imports map[string]Image
}

type Image struct {
	Author    string
	Compliant     bool
	Error bool
	Labels  map[string]string
	Obsolete bool
	Pulls string
	Stars int
	Tags map[string]Tag
}

type Tag struct {
	Age int
	Created string
	Digest string
	Size string
}

//Result struct representing JSON result
type Result struct {
	Name string
}

//UserRepositories Returns repositories for the given user
func (registry *ContainerYardRegistry) Repositories() ([]string, error) {
	url := registry.url("/search?q=%s&t=json", "-seed")
	repos := make([]string, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response Results

	err = registry.getContainerYardJson(url, &response)
	if err == nil {
		for repoName, image := range response.Community {
			for tagName, _ := range image.Tags {
				imageStr := repoName + ":" + tagName
				repos = append(repos, imageStr)
			}
		}
		for repoName, image := range response.Imports {
			for tagName, _ := range image.Tags {
				imageStr := repoName + ":" + tagName
				repos = append(repos, imageStr)
			}
		}
	}
	return repos, nil
}

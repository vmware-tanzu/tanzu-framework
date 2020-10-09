package cli

import "fmt"

// KnownRepositories is a list of known repositories.
var KnownRepositories = []Repository{
	CommunityGCPBucketRepository,
	TMCGCPBucketRepository,
}

// DefaultMultiRepo is the default multirepo with the known repositories.
var DefaultMultiRepo = NewMultiRepo(KnownRepositories...)

// MultiRepo is a multiple
type MultiRepo struct {
	repositories []Repository
}

// NewMultiRepo returns a new multirepo.
func NewMultiRepo(repositories ...Repository) *MultiRepo {
	return &MultiRepo{
		repositories: repositories,
	}
}

// AddRepository to known.
func (m *MultiRepo) AddRepository(repo Repository) {
	m.repositories = append(m.repositories, repo)
}

// RemoveRepository removes a repo.
func (m *MultiRepo) RemoveRepository(name string) {
	newRepos := []Repository{}
	for _, repo := range m.repositories {
		if name != repo.Name() {
			newRepos = append(newRepos, repo)
		}
	}
	m.repositories = newRepos
}

// GetRepository returns a repository.
func (m *MultiRepo) GetRepository(name string) (Repository, error) {
	for _, repo := range m.repositories {
		if name == repo.Name() {
			return repo, nil
		}
	}
	return nil, fmt.Errorf("could not find repository %q", name)
}

// ListPlugins across the repositories.
func (m *MultiRepo) ListPlugins() (mp map[string][]PluginDescriptor, err error) {
	mp = map[string][]PluginDescriptor{}
	for _, repo := range m.repositories {
		descriptors, err := repo.List()
		if err != nil {
			return mp, err
		}
		mp[repo.Name()] = descriptors
	}
	return
}

// Find a repository for the given plugin name.
// TODO: check for duplicates.
func (m *MultiRepo) Find(name string) (r Repository, err error) {
	for _, repo := range m.repositories {
		descriptors, err := repo.List()
		if err != nil {
			return r, err
		}
		for _, desc := range descriptors {
			if desc.Name == name {
				return repo, nil
			}
		}
	}
	return nil, fmt.Errorf("could not find plugin %q", name)
}

package registry

//go:generate counterfeiter -o ../../fakes/registy.go --fake-name Registry . Registry

type Registry interface {
	// ListImageTags lists all tags of the given image.
	ListImageTags(imageName string) ([]string, error)
	// GetFile gets the file content bundled in the given image:tag.
	// If filename is empty, it will get the first file.
	GetFile(image string, tag string, filename string) ([]byte, error)
}

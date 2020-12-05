package maps

// FilesGroupsMap is a mapping between files and groups
type FilesGroupsMap struct {
	Filenames []string `json:"filenames"`
	Groups    []string `json:"groups"`
}

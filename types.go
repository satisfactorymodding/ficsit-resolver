package resolver

import "context"

type Provider interface {
	SMLVersions(context context.Context) ([]SMLVersion, error)
	ModVersionsWithDependencies(context context.Context, modID string) ([]ModVersion, error)
	GetModName(context context.Context, modReference string) (*ModName, error)
}

type TargetName string

const (
	TargetNameLinuxServer   TargetName = "LinuxServer"
	TargetNameWindows       TargetName = "Windows"
	TargetNameWindowsServer TargetName = "WindowsServer"
)

type ModName struct {
	ID           string `json:"id"`
	ModReference string `json:"mod_reference"`
	Name         string `json:"name"`
}

type VersionDependency struct {
	ModReference string
	Constraint   string
	Optional     bool
}

type SMLVersion struct {
	ID                  string             `json:"id"`
	Version             string             `json:"version"`
	Targets             []SMLVersionTarget `json:"targets"`
	SatisfactoryVersion int                `json:"satisfactory_version"`
}

type SMLVersionTarget struct {
	TargetName TargetName `json:"targetName"`
	Link       string     `json:"link"`
}

type ModVersion struct {
	ID           string       `json:"id"`
	Version      string       `json:"version"`
	Dependencies []Dependency `json:"dependencies"`
	Targets      []Target     `json:"targets"`
}

type Dependency struct {
	ModID     string `json:"mod_id"`
	Condition string `json:"condition"`
	Optional  bool   `json:"optional"`
}

type Target struct {
	VersionID  string     `json:"version_id"`
	TargetName TargetName `json:"target_name"`
	Hash       string     `json:"hash"`
	Size       int64      `json:"size"`
}

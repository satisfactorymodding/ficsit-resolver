package resolver

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/mircearoata/pubgrub-go/pubgrub"
	"github.com/mircearoata/pubgrub-go/pubgrub/helpers"
	"github.com/mircearoata/pubgrub-go/pubgrub/semver"
	"github.com/puzpuzpuz/xsync/v3"
)

type ficsitAPISource struct {
	provider        Provider
	lockfile        *LockFile
	toInstall       map[string]semver.Constraint
	requiredTargets map[TargetName]bool
	modVersionInfo  *xsync.MapOf[string, []ModVersion]
	gameVersion     semver.Version
}

func (f *ficsitAPISource) GetPackageVersions(pkg string) ([]pubgrub.PackageVersion, error) {
	// If root package, return the base list of dependencies
	if pkg == rootPkg {
		return []pubgrub.PackageVersion{{Version: semver.Version{}, Dependencies: f.toInstall}}, nil
	}

	// Ignore game dependency
	if pkg == factoryGamePkg {
		return []pubgrub.PackageVersion{{Version: f.gameVersion}}, nil
	}

	response, err := f.provider.ModVersionsWithDependencies(context.TODO(), pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch mod %s: %w", pkg, err)
	}

	f.modVersionInfo.Store(pkg, response)

	versions := make([]pubgrub.PackageVersion, 0)
	for _, modVersion := range response {
		v, err := semver.NewVersion(modVersion.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to parse version %s: %w", modVersion.Version, err)
		}

		foundTargets := maps.Clone(f.requiredTargets)
		for _, target := range modVersion.Targets {
			delete(foundTargets, target.TargetName)
		}

		if len(foundTargets) > 0 {
			continue
		}

		dependencies := make(map[string]semver.Constraint)
		optionalDependencies := make(map[string]semver.Constraint)
		for _, dependency := range modVersion.Dependencies {
			c, err := semver.NewConstraint(dependency.Condition)
			if err != nil {
				return nil, fmt.Errorf("failed to parse constraint %s: %w", dependency.Condition, err)
			}

			if dependency.Optional {
				optionalDependencies[dependency.ModID] = c
			} else {
				dependencies[dependency.ModID] = c
			}
		}

		versions = append(versions, pubgrub.PackageVersion{
			Version:              v,
			Dependencies:         dependencies,
			OptionalDependencies: optionalDependencies,
		})
	}

	return versions, nil
}

func (f *ficsitAPISource) PickVersion(pkg string, versions []semver.Version) semver.Version {
	if f.lockfile != nil {
		if existing, ok := f.lockfile.Mods[pkg]; ok {
			v, err := semver.NewVersion(existing.Version)
			if err == nil {
				if slices.ContainsFunc(versions, func(version semver.Version) bool {
					return v.Compare(version) == 0
				}) {
					return v
				}
			}
		}
	}

	return helpers.StandardVersionPriority(versions)
}

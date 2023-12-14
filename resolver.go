package resolver

import (
	"context"
	"errors"
	"fmt"

	"github.com/mircearoata/pubgrub-go/pubgrub"
	"github.com/mircearoata/pubgrub-go/pubgrub/helpers"
	"github.com/mircearoata/pubgrub-go/pubgrub/semver"
	"github.com/puzpuzpuz/xsync/v3"
)

const (
	rootPkg        = "$$root$$"
	smlPkg         = "SML"
	factoryGamePkg = "FactoryGame"
)

type DependencyResolver struct {
	provider Provider
	apiBase  string
}

func NewDependencyResolver(provider Provider, apiBase string) DependencyResolver {
	return DependencyResolver{
		provider: provider,
		apiBase:  apiBase,
	}
}

func (d DependencyResolver) ResolveModDependencies(ctx context.Context, constraints map[string]string, lockFile *LockFile, gameVersion int, requiredTargets []TargetName) (*LockFile, error) {
	smlVersionsDB, err := d.provider.SMLVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed fetching SML versions: %w", err)
	}

	gameVersionSemver, err := semver.NewVersion(fmt.Sprintf("%d", gameVersion))
	if err != nil {
		return nil, fmt.Errorf("failed parsing game version: %w", err)
	}

	toInstall := make(map[string]semver.Constraint, len(constraints))
	for k, v := range constraints {
		c, err := semver.NewConstraint(v)
		if err != nil {
			return nil, fmt.Errorf("failed to parse constraint %s: %w", v, err)
		}
		toInstall[k] = c
	}

	mappedTargets := make(map[TargetName]bool, len(requiredTargets))
	for _, target := range requiredTargets {
		mappedTargets[target] = true
	}

	ficsitSource := &ficsitAPISource{
		provider:        d.provider,
		smlVersions:     smlVersionsDB,
		gameVersion:     gameVersionSemver,
		lockfile:        lockFile,
		toInstall:       toInstall,
		modVersionInfo:  xsync.NewMapOf[string, []ModVersion](),
		requiredTargets: mappedTargets,
	}

	result, err := pubgrub.Solve(helpers.NewCachingSource(ficsitSource), rootPkg)
	if err != nil {
		finalError := err
		var solverErr pubgrub.SolvingError
		if errors.As(err, &solverErr) {
			finalError = DependencyResolverError{SolvingError: solverErr, provider: d.provider, smlVersions: smlVersionsDB, gameVersion: gameVersion}
		}
		return nil, fmt.Errorf("failed to solve dependencies: %w", finalError)
	}

	delete(result, rootPkg)
	delete(result, factoryGamePkg)

	outputLock := NewLockfile()
	for k, v := range result {
		if k == smlPkg {
			for _, version := range ficsitSource.smlVersions {
				if version.Version == v.String() {
					targets := make(map[string]LockedModTarget)
					for _, target := range version.Targets {
						targets[string(target.TargetName)] = LockedModTarget{
							Link: target.Link,
						}
					}

					outputLock.Mods[k] = LockedMod{
						Version: v.String(),
						Targets: targets,
					}
					break
				}
			}
			continue
		}

		value, _ := ficsitSource.modVersionInfo.Load(k)
		for _, ver := range value {
			if ver.Version == v.RawString() {
				targets := make(map[string]LockedModTarget)
				for _, target := range ver.Targets {
					targets[string(target.TargetName)] = LockedModTarget{
						Link: d.apiBase + "/v1/version/" + ver.ID + "/" + string(target.TargetName) + "/download",
						Hash: target.Hash,
					}
				}

				outputLock.Mods[k] = LockedMod{
					Version: v.String(),
					Targets: targets,
				}

				break
			}
		}
	}

	return outputLock, nil
}

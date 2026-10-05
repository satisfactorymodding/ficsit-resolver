package resolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/mircearoata/pubgrub-go/pubgrub"
	"github.com/mircearoata/pubgrub-go/pubgrub/semver"
)

type DependencyResolverError struct {
	provider Provider
	pubgrub.SolvingError
	gameVersion int
}

func (e DependencyResolverError) Error() string {
	stringer := e.Stringer()
	return pubgrub.NewStandardTextReporter().
		WithIncompatibilityStringer(stringer).
		WithTermStringer(
			pubgrub.NewStandardTermStringer().
				WithPackageFormatter(stringer).
				WithConstraintFormatter(stringer),
		).
		Render(e.Report())
}

func (e DependencyResolverError) Stringer() *DependencyResolverErrorStringer {
	return MakeDependencyResolverErrorStringer(e.provider, e.gameVersion)
}

type DependencyResolverErrorStringer struct {
	provider     Provider
	packageNames map[string]string
	pubgrub.StandardIncompatibilityStringer
	gameVersion int
}

func MakeDependencyResolverErrorStringer(provider Provider, gameVersion int) *DependencyResolverErrorStringer {
	return &DependencyResolverErrorStringer{
		StandardIncompatibilityStringer: pubgrub.NewStandardIncompatibilityStringer(),
		provider:                        provider,
		gameVersion:                     gameVersion,
		packageNames:                    map[string]string{},
	}
}

func (w *DependencyResolverErrorStringer) GetPackageName(pkg string) string {
	if name, ok := w.packageNames[pkg]; ok {
		return name
	}

	result, err := w.provider.GetModName(context.TODO(), pkg)
	if err != nil {
		return pkg
	}

	w.packageNames[pkg] = result.Name

	return result.Name
}

func (w *DependencyResolverErrorStringer) FormatPackage(pkg string) string {
	if pkg == FactoryGamePkg {
		return "Satisfactory"
	}

	name := w.GetPackageName(pkg)
	if name == pkg {
		return name
	}
	return fmt.Sprintf("%s (%s)", name, pkg)
}

func (w *DependencyResolverErrorStringer) FormatConstraint(pkg string, constraint semver.Constraint) string {
	if pkg == FactoryGamePkg {
		return strings.ReplaceAll(constraint.String(), ".0.0", "")
	}

	res, err := w.provider.ModVersionsWithDependencies(context.TODO(), pkg)
	if err != nil {
		return constraint.String()
	}

	var matched []semver.Version
	for _, v := range res {
		ver, err := semver.NewVersion(v.Version)
		if err != nil {
			// Assume it is contained in the constraint
			matched = append(matched, semver.Version{})
			continue
		}

		if constraint.Contains(ver) {
			matched = append(matched, ver)
		}
	}

	if len(matched) == 1 {
		return matched[0].RawString()
	}

	return constraint.String()
}

func (w *DependencyResolverErrorStringer) IncompatibilityString(incompatibility *pubgrub.Incompatibility, ts pubgrub.TermStringer, rootPkg string) string {
	if env, ok := incompatibility.Cause().(pubgrub.EnvironmentPackageCause); ok {
		if env.Pkg == FactoryGamePkg {
			return fmt.Sprintf("Satisfactory CL%d is installed", w.gameVersion)
		}
	}
	return w.StandardIncompatibilityStringer.IncompatibilityString(incompatibility, ts, rootPkg)
}

package resolver

import (
	"math"
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestProfileResolution(t *testing.T) {
	resolver := NewDependencyResolver(MockProvider{})

	resolved, err := resolver.ResolveModDependencies(map[string]string{
		"RefinedPower": "3.2.10",
	}, nil, math.MaxInt, nil)

	testza.AssertNoError(t, err)
	testza.AssertNotNil(t, resolved)
	testza.AssertLen(t, resolved.Mods, 4)
}

func TestProfileRequiredOlderVersion(t *testing.T) {
	resolver := NewDependencyResolver(MockProvider{})

	_, err := resolver.ResolveModDependencies(map[string]string{
		"RefinedPower": "3.2.11",
		"RefinedRDLib": "1.1.5",
	}, nil, math.MaxInt, nil)

	testza.AssertEqual(t, "failed to solve dependencies: Because installing Refined Power (RefinedPower) \"3.2.11\" and Refined Power (RefinedPower) \"3.2.11\" depends on RefinedRDLib \"^1.1.6\", installing RefinedRDLib \"^1.1.6\".\nSo, because installing RefinedRDLib \"1.1.5\", version solving failed.", err.Error())
}

func TestResolutionNonExistentMod(t *testing.T) {
	resolver := NewDependencyResolver(MockProvider{})

	_, err := resolver.ResolveModDependencies(map[string]string{
		"ThisModDoesNotExist$$$": ">0.0.0",
	}, nil, math.MaxInt, nil)

	testza.AssertEqual(t, "failed to solve dependencies: failed to make decision: failed to get package versions: failed to fetch mod ThisModDoesNotExist$$$: mod not found", err.Error())
}

func TestInvalidConstraint(t *testing.T) {
	resolver := NewDependencyResolver(MockProvider{})

	_, err := resolver.ResolveModDependencies(map[string]string{
		"ThisModDoesNotExist$$$": "Hello",
	}, nil, math.MaxInt, nil)

	testza.AssertEqual(t, "failed to parse constraint Hello: failed to de-sugar range : invalid comparator string: Hello", err.Error())
}

func TestOldGameVersion(t *testing.T) {
	resolver := NewDependencyResolver(MockProvider{})

	_, err := resolver.ResolveModDependencies(map[string]string{
		"RefinedPower": "*",
	}, nil, 0, nil)

	testza.AssertEqual(t, `failed to solve dependencies: Because Refined Power (RefinedPower) "<3.2.13" depends on Satisfactory Mod Loader (SML) "^3.6.0" and Refined Power (RefinedPower) "3.2.13" depends on Satisfactory Mod Loader (SML) "3.6.1", every version of Refined Power (RefinedPower) depends on Satisfactory Mod Loader (SML) "^3.6.0".
And because Satisfactory Mod Loader (SML) ">=3.6.0" depends on Satisfactory (FactoryGame) ">=264901", every version of Refined Power (RefinedPower) depends on Satisfactory (FactoryGame) ">=264901".
So, because Satisfactory CL0 is installed, version solving failed.`, err.Error())
}

func TestLockfileResolution(t *testing.T) {
	resolver := NewDependencyResolver(MockProvider{})

	lockfile := NewLockfile()
	lockfile.Mods["RefinedPower"] = LockedMod{
		Version: "3.2.11",
	}

	resolved, err := resolver.ResolveModDependencies(map[string]string{
		"RefinedPower": ">=3.2.10",
	}, lockfile, math.MaxInt, nil)

	testza.AssertNoError(t, err)
	testza.AssertNotNil(t, resolved)
	testza.AssertLen(t, resolved.Mods, 4)
	testza.AssertEqual(t, resolved.Mods["RefinedPower"].Version, "3.2.11")
}

func TestMissingTarget(t *testing.T) {
	resolver := NewDependencyResolver(MockProvider{})

	_, err := resolver.ResolveModDependencies(map[string]string{
		"RefinedPower": "*",
	}, nil, math.MaxInt, []TargetName{"NotARealTarget"})

	testza.AssertEqual(t, "invalid target: NotARealTarget", err.Error())
}

func TestResolveForAllTargets(t *testing.T) {
	resolver := NewDependencyResolver(MockProvider{})

	resolved, err := resolver.ResolveModDependencies(map[string]string{
		"ComplexMod": "*",
	}, nil, math.MaxInt, []TargetName{"Windows", "LinuxServer"})

	testza.AssertNoError(t, err)
	testza.AssertNotNil(t, resolved)
	testza.AssertLen(t, resolved.Mods, 2)
	testza.AssertEqual(t, resolved.Mods["ComplexMod"].Version, "2.0.0")
}

func TestNoMatchForAllTargets(t *testing.T) {
	resolver := NewDependencyResolver(MockProvider{})

	_, err := resolver.ResolveModDependencies(map[string]string{
		"ComplexMod": ">=3.0.0",
	}, nil, math.MaxInt, []TargetName{"Windows", "LinuxServer"})

	testza.AssertEqual(t, "failed to solve dependencies: So, because installing ComplexMod \"3.0.0\" and ComplexMod \"3.0.0\" is forbidden, version solving failed.", err.Error())
}

func TestMatchForAllTargetsNotRequiredOnRemote(t *testing.T) {
	resolver := NewDependencyResolver(MockProvider{})

	resolved, err := resolver.ResolveModDependencies(map[string]string{
		"ClientOnlyMod": "*",
	}, nil, math.MaxInt, []TargetName{"Windows", "WindowsServer", "LinuxServer"})

	testza.AssertNoError(t, err)
	testza.AssertNotNil(t, resolved)
	testza.AssertLen(t, resolved.Mods, 1)
	testza.AssertEqual(t, resolved.Mods["ClientOnlyMod"].Version, "1.0.0")

	resolved, err = resolver.ResolveModDependencies(map[string]string{
		"ServerOnlyMod": "<=1.0.0",
	}, nil, math.MaxInt, []TargetName{"Windows", "WindowsServer", "LinuxServer"})

	testza.AssertNoError(t, err)
	testza.AssertNotNil(t, resolved)
	testza.AssertLen(t, resolved.Mods, 1)
	testza.AssertEqual(t, resolved.Mods["ServerOnlyMod"].Version, "1.0.0")

	resolved, err = resolver.ResolveModDependencies(map[string]string{
		"ClientOnlyMod": "*",
		"ServerOnlyMod": "<=1.0.0",
	}, nil, math.MaxInt, []TargetName{"Windows", "WindowsServer", "LinuxServer"})

	testza.AssertNoError(t, err)
	testza.AssertNotNil(t, resolved)
	testza.AssertLen(t, resolved.Mods, 2)
	testza.AssertEqual(t, resolved.Mods["ClientOnlyMod"].Version, "1.0.0")
	testza.AssertEqual(t, resolved.Mods["ServerOnlyMod"].Version, "1.0.0")
}

func TestNoMatchForAllTargetsNotRequiredOnRemote(t *testing.T) {
	resolver := NewDependencyResolver(MockProvider{})

	_, err := resolver.ResolveModDependencies(map[string]string{
		"ServerOnlyMod": ">=2.0.0",
	}, nil, math.MaxInt, []TargetName{"WindowsServer", "LinuxServer"})

	testza.AssertEqual(t, "failed to solve dependencies: So, because installing ServerOnlyMod \"2.0.0\" and ServerOnlyMod \"2.0.0\" is forbidden, version solving failed.", err.Error())
}

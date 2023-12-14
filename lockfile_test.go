package resolver

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestLockFile(t *testing.T) {
	firstLockFile := NewLockfile()

	firstLockFile.Mods["Hello"] = LockedMod{
		Version: "1.0.0",
	}

	firstLockFile.Mods["World"] = LockedMod{
		Version: "2.0.0",
	}

	testza.AssertEqual(t, firstLockFile.Mods["Hello"].Version, "1.0.0")
	testza.AssertEqual(t, firstLockFile.Mods["World"].Version, "2.0.0")

	secondLockFile := firstLockFile.Clone()

	testza.AssertEqual(t, secondLockFile.Mods["Hello"].Version, "1.0.0")
	testza.AssertEqual(t, secondLockFile.Mods["World"].Version, "2.0.0")

	secondLockFile.Mods["Foo"] = LockedMod{
		Version: "3.0.0",
	}

	secondLockFile.Mods["Bar"] = LockedMod{
		Version: "4.0.0",
	}

	testza.AssertEqual(t, secondLockFile.Mods["Foo"].Version, "3.0.0")
	testza.AssertEqual(t, secondLockFile.Mods["Bar"].Version, "4.0.0")

	testza.AssertEqual(t, firstLockFile.Mods["Foo"].Version, "")
	testza.AssertEqual(t, firstLockFile.Mods["Bar"].Version, "")

	firstLockFile = firstLockFile.Remove("Hello")

	testza.AssertEqual(t, firstLockFile.Mods["Hello"].Version, "")
	testza.AssertEqual(t, secondLockFile.Mods["Hello"].Version, "1.0.0")

	secondLockFile = secondLockFile.Remove("Foo")

	testza.AssertEqual(t, secondLockFile.Mods["Foo"].Version, "")
}

package main

import (
	"testing"
)

func TestFetchCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "fetch" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected 'fetch' command to be registered on rootCmd")
	}
}

func TestFetchCommandFlags(t *testing.T) {
	flags := map[string]struct {
		shorthand string
		defValue  string
	}{
		"output": {shorthand: "o", defValue: ""},
		"dir":    {shorthand: "d", defValue: "."},
	}

	for name, want := range flags {
		f := fetchCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("expected flag %q to be registered on fetchCmd", name)
			continue
		}
		if f.Shorthand != want.shorthand {
			t.Errorf("flag %q: shorthand = %q, want %q", name, f.Shorthand, want.shorthand)
		}
		if f.DefValue != want.defValue {
			t.Errorf("flag %q: default = %q, want %q", name, f.DefValue, want.defValue)
		}
	}
}

func TestFetchCommandRequiresArgs(t *testing.T) {
	// The fetch command requires exactly one argument (the URL).
	if fetchCmd.Args == nil {
		t.Fatal("expected fetchCmd.Args to be set (cobra.ExactArgs(1))")
	}
}

func TestFetchCommandHasHelp(t *testing.T) {
	if fetchCmd.Short == "" {
		t.Error("fetch command should have a short description")
	}
	if fetchCmd.Long == "" {
		t.Error("fetch command should have a long description")
	}
}

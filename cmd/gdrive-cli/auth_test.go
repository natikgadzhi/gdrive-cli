package main

import (
	"testing"
)

func TestAuthCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "auth" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected 'auth' command to be registered on rootCmd")
	}
}

func TestAuthSubcommandsRegistered(t *testing.T) {
	subcommands := map[string]bool{
		"login":  false,
		"status": false,
	}

	for _, cmd := range authCmd.Commands() {
		if _, ok := subcommands[cmd.Name()]; ok {
			subcommands[cmd.Name()] = true
		}
	}

	for name, found := range subcommands {
		if !found {
			t.Errorf("expected %q subcommand to be registered on authCmd", name)
		}
	}
}

func TestAuthCommandHasHelp(t *testing.T) {
	if authCmd.Short == "" {
		t.Error("auth command should have a short description")
	}
	if authLoginCmd.Short == "" {
		t.Error("auth login command should have a short description")
	}
	if authStatusCmd.Short == "" {
		t.Error("auth status command should have a short description")
	}
}

package main

import (
	"testing"
)

func TestSearchCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "search" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected 'search' command to be registered on rootCmd")
	}
}

func TestSearchCommandHasHelp(t *testing.T) {
	if searchCmd.Short == "" {
		t.Error("search command should have a short description")
	}
	if searchCmd.Long == "" {
		t.Error("search command should have a long description")
	}
}

func TestSearchCommandRequiresExactlyOneArg(t *testing.T) {
	// ExactArgs(1) means the command needs exactly one argument.
	err := searchCmd.Args(searchCmd, []string{})
	if err == nil {
		t.Error("expected error when no arguments are provided")
	}

	err = searchCmd.Args(searchCmd, []string{"query"})
	if err != nil {
		t.Errorf("expected no error with one argument, got: %v", err)
	}

	err = searchCmd.Args(searchCmd, []string{"arg1", "arg2"})
	if err == nil {
		t.Error("expected error when two arguments are provided")
	}
}

func TestSearchCommandLimitFlag(t *testing.T) {
	flag := searchCmd.Flags().Lookup("limit")
	if flag == nil {
		t.Fatal("expected 'limit' flag to be registered on search command")
	}
	if flag.DefValue != "20" {
		t.Errorf("expected default limit to be 20, got %s", flag.DefValue)
	}
	if flag.Shorthand != "n" {
		t.Errorf("expected limit shorthand to be 'n', got %s", flag.Shorthand)
	}
}

func TestSearchResponseStructure(t *testing.T) {
	// Verify the searchResponse struct can be created with the expected fields.
	resp := searchResponse{
		Query:   "test query",
		Count:   0,
		Results: nil,
	}
	if resp.Query != "test query" {
		t.Errorf("expected query 'test query', got %q", resp.Query)
	}
	if resp.Count != 0 {
		t.Errorf("expected count 0, got %d", resp.Count)
	}
	if resp.Results != nil {
		t.Error("expected nil results")
	}
}

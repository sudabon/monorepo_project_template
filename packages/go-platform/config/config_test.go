package config

import (
	"testing"
)

func TestRequireRejectsMissingAndEmpty(t *testing.T) {
	t.Setenv("CONFIG_TEST_REQUIRE", "value")
	got, err := Require("CONFIG_TEST_REQUIRE")
	if err != nil || got != "value" {
		t.Fatalf("set = %q, %v", got, err)
	}
	t.Setenv("CONFIG_TEST_REQUIRE", "")
	if _, err := Require("CONFIG_TEST_REQUIRE"); err == nil {
		t.Fatal("empty must fail")
	}
	if _, err := Require("CONFIG_TEST_REQUIRE_UNSET"); err == nil {
		t.Fatal("unset must fail")
	}
}

func TestOrFallsBackOnMissingAndEmpty(t *testing.T) {
	t.Setenv("CONFIG_TEST_OR", "set")
	if got := Or("CONFIG_TEST_OR", "fallback"); got != "set" {
		t.Fatalf("set = %q", got)
	}
	t.Setenv("CONFIG_TEST_OR", "")
	if got := Or("CONFIG_TEST_OR", "fallback"); got != "fallback" {
		t.Fatalf("empty = %q", got)
	}
	if got := Or("CONFIG_TEST_OR_UNSET", "fallback"); got != "fallback" {
		t.Fatalf("unset = %q", got)
	}
}

func TestBoolParsesOrFallsBackAndRejectsInvalid(t *testing.T) {
	t.Setenv("CONFIG_TEST_BOOL", "true")
	got, err := Bool("CONFIG_TEST_BOOL", false)
	if err != nil || !got {
		t.Fatalf("true = %v, %v", got, err)
	}
	t.Setenv("CONFIG_TEST_BOOL", "0")
	got, err = Bool("CONFIG_TEST_BOOL", true)
	if err != nil || got {
		t.Fatalf("0 = %v, %v", got, err)
	}
	t.Setenv("CONFIG_TEST_BOOL", "")
	got, err = Bool("CONFIG_TEST_BOOL", true)
	if err != nil || !got {
		t.Fatalf("empty = %v, %v", got, err)
	}
	got, err = Bool("CONFIG_TEST_BOOL_UNSET", true)
	if err != nil || !got {
		t.Fatalf("unset = %v, %v", got, err)
	}
	t.Setenv("CONFIG_TEST_BOOL", "yes")
	got, err = Bool("CONFIG_TEST_BOOL", true)
	if err == nil || got {
		t.Fatalf("invalid = %v, %v", got, err)
	}
}

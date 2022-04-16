package auth

import (
  "testing"
)

// testing https://go.dev/doc/tutorial/add-a-test
func TestRegister(t *testing.T) {
  expected := "registering new user"
  got := register()
  if got != expected {
    t.Errorf("unexpected return value: %q", got)
  }
}

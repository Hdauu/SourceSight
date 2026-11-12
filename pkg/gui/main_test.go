package gui

import (
	"testing"

	"sourcesight/pkg/testenv"
)

// Tests here write history and database files: keep them away from the real ones.
func TestMain(m *testing.M) { testenv.Run(m) }

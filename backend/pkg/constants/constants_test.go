package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsVisiblePart(t *testing.T) {
	assert.False(t, IsVisiblePart(PartQuantitySold), "quantity -1 must be hidden")
	assert.True(t, IsVisiblePart(PartQuantityDefective), "quantity 0 must be visible (defect report)")
	assert.True(t, IsVisiblePart(1), "quantity > 0 must be visible")
	assert.True(t, IsVisiblePart(100), "quantity > 0 must be visible")
}

func TestIsValidRole(t *testing.T) {
	assert.True(t, IsValidRole("admin"))
	assert.True(t, IsValidRole("manager"))
	assert.True(t, IsValidRole("operator"))
	assert.False(t, IsValidRole("superadmin"))
	assert.False(t, IsValidRole(""))
}

package browser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrowserSingleton(t *testing.T) {
	// Initialize the browser with default config
	browser1, playwright1, err := Init(DefaultConfig())
	assert.NoError(t, err, "Failed to initialize browser")
	assert.NotNil(t, browser1, "Browser should not be nil")
	assert.NotNil(t, playwright1, "Playwright instance should not be nil")

	// Initialize again - should return the same instances
	browser2, playwright2, err := Init(DefaultConfig())
	assert.NoError(t, err, "Failed to initialize browser on second call")
	assert.NotNil(t, browser2, "Browser should not be nil on second call")
	assert.NotNil(t, playwright2, "Playwright instance should not be nil on second call")

	// Verify singleton behavior - should be the exact same instances
	assert.Equal(t, browser1, browser2, "Browser instance should be a singleton")
	assert.Equal(t, playwright1, playwright2, "Playwright instance should be a singleton")

	// Clean up after test
	defer CleanupSingleton()
}

func TestNewContext(t *testing.T) {
	// Skip if playwright initialization fails
	browser, _, err := Init(DefaultConfig())
	if err != nil {
		t.Skip("Skipping context test due to browser initialization error:", err)
	}
	defer CleanupSingleton()

	// Create a new context
	ctx, err := NewContext(browser, DefaultConfig())
	assert.NoError(t, err, "Failed to create browser context")
	assert.NotNil(t, ctx, "Browser context should not be nil")

	// Contexts are not singletons, so creating another should give a different instance
	ctx2, err := NewContext(browser, DefaultConfig())
	assert.NoError(t, err, "Failed to create second browser context")
	assert.NotNil(t, ctx2, "Second browser context should not be nil")
	assert.NotEqual(t, ctx, ctx2, "Browser contexts should be different instances")

	// Clean up contexts
	err = ctx.Close()
	assert.NoError(t, err, "Failed to close first context")

	err = ctx2.Close()
	assert.NoError(t, err, "Failed to close second context")
}

func TestNewPage(t *testing.T) {
	// Skip if playwright initialization fails
	browser, _, err := Init(DefaultConfig())
	if err != nil {
		t.Skip("Skipping page test due to browser initialization error:", err)
	}
	defer CleanupSingleton()

	// Create a context for the page
	ctx, err := NewContext(browser, DefaultConfig())
	if err != nil {
		t.Skip("Skipping page test due to context initialization error:", err)
	}
	defer ctx.Close()

	// Create a new page
	page, err := NewPage(ctx)
	assert.NoError(t, err, "Failed to create page")
	assert.NotNil(t, page, "Page should not be nil")

	// Pages are not singletons, so creating another should give a different instance
	page2, err := NewPage(ctx)
	assert.NoError(t, err, "Failed to create second page")
	assert.NotNil(t, page2, "Second page should not be nil")
	assert.NotEqual(t, page, page2, "Pages should be different instances")

	// Clean up pages
	err = page.Close()
	assert.NoError(t, err, "Failed to close first page")

	err = page2.Close()
	assert.NoError(t, err, "Failed to close second page")
}

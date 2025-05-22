package features

import (
	"errors"
	"testing"
)

// MockFeature implements the Feature interface for testing
type MockFeature struct {
	name           string
	description    string
	flags          []Flag
	shouldActivate bool
	validateError  error
	executeError   error
	priority       int
	currentState   map[string]any
	shouldPrompt   bool
}

func (m *MockFeature) Name() string {
	return m.name
}

func (m *MockFeature) Description() string {
	return m.description
}

func (m *MockFeature) Flags() []Flag {
	return m.flags
}

func (m *MockFeature) ShouldActivate(options map[string]any) bool {
	return m.shouldActivate
}

func (m *MockFeature) ValidateOptions(options map[string]any) error {
	return m.validateError
}

func (m *MockFeature) Execute(ctx *ExecutionContext) error {
	return m.executeError
}

func (m *MockFeature) Priority() int {
	return m.priority
}

func (m *MockFeature) DetectCurrentState(ctx *ExecutionContext) (map[string]any, error) {
	if m.currentState == nil {
		return make(map[string]any), nil
	}
	return m.currentState, nil
}

func (m *MockFeature) DisplayCurrentState(ctx *ExecutionContext, state map[string]any) {
	// Do nothing in mock implementation
}

func (m *MockFeature) ShouldPromptUser(ctx *ExecutionContext, state map[string]any) bool {
	return m.shouldPrompt
}

func TestRegistry(t *testing.T) {
	// Create a new registry
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("NewRegistry returned nil")
	}

	if len(registry.features) != 0 {
		t.Errorf("New registry should have 0 features, got %d", len(registry.features))
	}

	// Create mock features
	feature1 := &MockFeature{
		name:           "feature1",
		description:    "Feature 1",
		shouldActivate: true,
	}

	feature2 := &MockFeature{
		name:           "feature2",
		description:    "Feature 2",
		shouldActivate: false,
	}

	// Register features
	registry.Register(feature1)
	registry.Register(feature2)

	// Test GetFeatures
	features := registry.GetFeatures()
	if len(features) != 2 {
		t.Errorf("Expected 2 features, got %d", len(features))
	}

	// Test GetActiveFeatures
	activeFeatures := registry.GetActiveFeatures(map[string]any{})
	if len(activeFeatures) != 1 {
		t.Errorf("Expected 1 active feature, got %d", len(activeFeatures))
	}

	if activeFeatures[0].Name() != "feature1" {
		t.Errorf("Expected active feature to be 'feature1', got %q", activeFeatures[0].Name())
	}
}

func TestExecutionContext(t *testing.T) {
	// Create a new execution context
	ctx := &ExecutionContext{
		Options: map[string]any{
			"key1": "value1",
			"key2": 123,
		},
		DryRun:      true,
		Interactive: false,
		Verbose:     true,
	}

	// Test direct access to Options
	value1, ok := ctx.Options["key1"].(string)
	if !ok {
		t.Error("Options access should work for existing key")
	}
	if value1 != "value1" {
		t.Errorf("Options access returned %q, expected %q", value1, "value1")
	}

	value2, ok := ctx.Options["key2"].(int)
	if !ok {
		t.Error("Options access should work for existing key")
	}
	if value2 != 123 {
		t.Errorf("Options access returned %d, expected %d", value2, 123)
	}

	// Test Options access with wrong type
	_, ok = ctx.Options["key1"].(int)
	if ok {
		t.Error("Options access should return false for wrong type")
	}

	// Test Options access with non-existent key
	_, ok = ctx.Options["non-existent"]
	if ok {
		t.Error("Options access should return false for non-existent key")
	}
}

// Helper functions for testing
func validateFeatures(registry *Registry, options map[string]any) error {
	for _, feature := range registry.GetActiveFeatures(options) {
		if err := feature.ValidateOptions(options); err != nil {
			return err
		}
	}
	return nil
}

func executeFeatures(registry *Registry, ctx *ExecutionContext) error {
	for _, feature := range registry.GetActiveFeatures(ctx.Options) {
		if err := feature.Execute(ctx); err != nil {
			return err
		}
	}
	return nil
}

func TestValidateAndExecuteFeatures(t *testing.T) {
	// Create a registry with mock features
	registry := NewRegistry()

	// Feature that validates and executes successfully
	successFeature := &MockFeature{
		name:           "success",
		description:    "Success Feature",
		shouldActivate: true,
		validateError:  nil,
		executeError:   nil,
	}

	// Feature that fails validation
	validateFailFeature := &MockFeature{
		name:           "validate-fail",
		description:    "Validation Failure Feature",
		shouldActivate: true,
		validateError:  errors.New("validation error"),
		executeError:   nil,
	}

	// Feature that fails execution
	executeFailFeature := &MockFeature{
		name:           "execute-fail",
		description:    "Execution Failure Feature",
		shouldActivate: true,
		validateError:  nil,
		executeError:   errors.New("execution error"),
	}

	// Register features
	registry.Register(successFeature)
	registry.Register(validateFailFeature)
	registry.Register(executeFailFeature)

	// Create execution context
	ctx := &ExecutionContext{
		Options: map[string]any{},
		DryRun:  false,
	}

	// Test validateFeatures
	err := validateFeatures(registry, ctx.Options)
	if err == nil {
		t.Error("validateFeatures should return error when a feature fails validation")
	}

	// Remove the failing feature and try again
	registry = NewRegistry()
	registry.Register(successFeature)
	registry.Register(executeFailFeature)

	err = validateFeatures(registry, ctx.Options)
	if err != nil {
		t.Errorf("validateFeatures should not return error: %v", err)
	}

	// Test executeFeatures
	err = executeFeatures(registry, ctx)
	if err == nil {
		t.Error("executeFeatures should return error when a feature fails execution")
	}

	// Remove the failing feature and try again
	registry = NewRegistry()
	registry.Register(successFeature)

	err = executeFeatures(registry, ctx)
	if err != nil {
		t.Errorf("executeFeatures should not return error: %v", err)
	}
}

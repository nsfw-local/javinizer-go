// Package testutil provides shared test utilities and helpers for javinizer-go tests.
//
// This file contains a standardized table-driven test template that serves as the
// canonical pattern for writing tests across all packages in javinizer-go.
//
// # Table-Driven Test Template
//
// Table-driven tests are the preferred pattern when testing a function with multiple
// input scenarios. They provide consistency, readability, and easy test case addition.
//
// # When to Use Table-Driven Tests
//
//   - Testing a function with 3+ different input scenarios
//   - Validating error handling with multiple failure modes
//   - Testing transformations (input → output pairs)
//   - Comparing multiple implementations or configurations
//
// # When NOT to Use Table-Driven Tests
//
//   - Single test scenario (use simple test function)
//   - Complex setup/teardown per case (use separate test functions)
//   - Tests requiring different assertions per case (separate functions clearer)
//
// # Template Structure
//
// The template below is copy-pasteable actual Go code (not pseudocode).
// Replace placeholders with your specific types and function names.
//
// Key components explained:
//
//  1. Test struct with standard fields:
//     - name: Test case description for t.Run() (required)
//     - input: Input data for the function under test
//     - want: Expected output value
//     - wantErr: Whether this test case expects an error
//
//  2. Test table: Slice of test structs with multiple scenarios
//
//  3. for loop with t.Run(): Executes each test case as a subtest
//     - Enables running individual tests: go test -run TestFunction/specific_case
//     - Enables parallel execution: add t.Parallel() if needed
//     - Provides clear output showing which case failed
//
//  4. Error handling pattern:
//     - Check wantErr first with early return
//     - Prevents false positives from missing error checks
//
//  5. Success assertions:
//     - assert.NoError first to catch unexpected errors
//     - assert.Equal to verify output matches expected
//
// # Complete Copy-Pasteable Template
//
// Replace the following placeholders:
//   - TestFunction: Your test function name (e.g., TestValidateID)
//   - FunctionUnderTest: The function being tested (e.g., ValidateID)
//   - InputType: Type of input parameter (e.g., string, *Movie)
//   - OutputType: Type of return value (e.g., bool, string)
//
// ```go
//
//	func TestFunction(t *testing.T) {
//	    tests := []struct {
//	        name    string      // Test case name for t.Run()
//	        input   InputType   // Input data
//	        want    OutputType  // Expected output
//	        wantErr bool        // Expect error?
//	    }{
//	        {
//	            name:    "success case",
//	            input:   validInput,
//	            want:    expectedOutput,
//	            wantErr: false,
//	        },
//	        {
//	            name:    "error case",
//	            input:   invalidInput,
//	            want:    zeroValue,  // Use zero value when expecting error
//	            wantErr: true,
//	        },
//	    }
//
//	    for _, tt := range tests {
//	        t.Run(tt.name, func(t *testing.T) {
//	            got, err := FunctionUnderTest(tt.input)
//
//	            if tt.wantErr {
//	                assert.Error(t, err)
//	                return  // Early return prevents checking 'got' on error path
//	            }
//
//	            assert.NoError(t, err)
//	            assert.Equal(t, tt.want, got)
//	        })
//	    }
//	}
//
// ```
//
// # Data Pattern Decision Tree
//
// Choose your test data pattern based on complexity:
//
//  1. Inline Data (Simple):
//     - Use when: Input/output are primitives or small structs (≤5 fields)
//     - Example: Strings, numbers, booleans, simple validation
//     - Benefits: Readable, self-contained, easy to understand
//
//  2. Builder Pattern (Complex Entities):
//     - Use when: Input/output are domain models (Movie, Actress)
//     - Use: testutil.NewMovieBuilder().WithTitle("X").Build()
//     - Benefits: Fluent API, sensible defaults, less boilerplate
//     - See: internal/testutil/builders.go
//
//  3. Golden Files (Large Text):
//     - Use when: Output is large text/HTML/JSON/XML
//     - Use: testutil.LoadGoldenFile(t, "file.golden")
//     - Benefits: Keeps tests clean, snapshot testing for complex output
//     - See: internal/testutil/helpers.go
//
// # Examples
//
// See internal/testutil/template_test.go for working examples:
//   - Example 1: Simple validation with inline data
//   - Example 2: Entity transformation with builders
//   - Example 3: Large output with golden files
//
// Canonical reference: internal/matcher/matcher_test.go
//   - 76 test cases demonstrating table-driven pattern
//   - Shows naming conventions and structure
//   - Production-quality example
//
// # Common Mistakes to Avoid
//
//   - ❌ Missing name field (breaks t.Run naming)
//   - ❌ Not using t.Run() (can't run individual subtests)
//   - ❌ Hardcoding file paths (use filepath.Join or golden file helpers)
//   - ❌ Testing too many behaviors in one table (split into separate test functions)
//   - ❌ Using bare if statements instead of testify assertions
//   - ❌ Not checking error when wantErr is true (allows silent failures)
//
// # Naming Conventions
//
//   - Test function: TestFunctionName (capitalized, no underscores)
//   - Subtest names: tt.name field (descriptive, spaces allowed, lowercase preferred)
//   - Test file: package_test.go (matches source file)
//   - Example: TestParseMovieID with subtests "valid format", "invalid format"
package testutil

// TableDrivenTestTemplate is a documentation placeholder.
// This file serves as a reference template and does not export runnable code.
// Copy the template from the godoc comments above when creating new tests.
//
// For working examples, see:
//   - internal/testutil/template_test.go (demonstrates all 3 data patterns)
//   - internal/matcher/matcher_test.go (canonical 76-case example)
//
// Template location for AI agents: internal/testutil/template.go
const TableDrivenTestTemplate = `
See package documentation above for complete copy-pasteable template.

Quick reference structure:

tests := []struct {
    name    string
    input   InputType
    want    OutputType
    wantErr bool
}{
    {name: "case1", input: val1, want: expected1, wantErr: false},
    {name: "error", input: bad, want: zero, wantErr: true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got, err := Function(tt.input)
        if tt.wantErr {
            assert.Error(t, err)
            return
        }
        assert.NoError(t, err)
        assert.Equal(t, tt.want, got)
    })
}
`

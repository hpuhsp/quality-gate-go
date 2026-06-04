package cmd

import "fmt"

// RunPrePush is a placeholder — CI handles real tests.
func RunPrePush() {
	fmt.Println("⚠️  Unit tests are handled by CI, not local pre-push.")
	fmt.Println("   Pre-push hook is active but skips tests by default.")
	fmt.Println("   Run 'quality-gate setup' to enable local test execution.")
}

package gotaskflow

import (
	"fmt"
	"strings"
)

// validate checks a traceRecord against the expected TaskFlow DAG.
// A nil rec (e.g. no tracer configured) is treated as "no trace data" and always returns valid.
// Internal-only: used in tests to verify execution correctness.
func validate(rec traceRecord, tf *TaskFlow) *validationResult {
	if rec == nil {
		return &validationResult{valid: true}
	}
	return newValidator(rec).run(tf)
}

// validationResult contains the result of validating trace events against a TaskFlow.
type validationResult struct {
	valid            bool
	missingTasks     []string          // Tasks defined but not executed
	unexpectedTasks  []string          // Tasks executed but not defined
	dependencyErrors []dependencyError // Dependency mismatches
	skippedBranches  []string          // Condition branches that were skipped (not errors)
}

// dependencyError represents a mismatch in task dependencies.
type dependencyError struct {
	task     string
	expected []string
	actual   []string
}

func (e dependencyError) String() string {
	return fmt.Sprintf("task %q: expected deps %v, actual %v", e.task, e.expected, e.actual)
}

func (r *validationResult) String() string {
	if r.valid {
		return "validation passed"
	}
	var sb strings.Builder
	sb.WriteString("validation failed:\n")
	if len(r.missingTasks) > 0 {
		sb.WriteString(fmt.Sprintf("  missing tasks: %v\n", r.missingTasks))
	}
	if len(r.unexpectedTasks) > 0 {
		sb.WriteString(fmt.Sprintf("  unexpected tasks: %v\n", r.unexpectedTasks))
	}
	for _, e := range r.dependencyErrors {
		sb.WriteString(fmt.Sprintf("  %s\n", e.String()))
	}
	if len(r.skippedBranches) > 0 {
		sb.WriteString(fmt.Sprintf("  skipped branches (OK): %v\n", r.skippedBranches))
	}
	return sb.String()
}

// validator validates a traceRecord against TaskFlow definitions.
type validator struct {
	rec traceRecord
}

func newValidator(rec traceRecord) *validator {
	return &validator{rec: rec}
}

// run compares executed trace events against the expected TaskFlow DAG.
func (v *validator) run(tf *TaskFlow) *validationResult {
	result := &validationResult{valid: true}

	// --- Step 1: collect expected nodes via eGraph.walk ---
	expected := make(map[string]*innerNode)
	tf.graph.walk(func(n *innerNode) { expected[n.name] = n })

	// --- Step 2: build executed map from the immutable record ---
	executed := make(map[string]chromeTraceEvent, len(v.rec))
	for _, ev := range v.rec {
		executed[ev.Name] = ev
	}

	// --- Step 3: missing / skipped check ---
	// A task is a skipped branch if:
	//   (a) it is a direct successor of a condition node that chose a different branch, OR
	//   (b) any of its non-condition predecessors is also skipped (transitive skip).
	skipped := make(map[string]bool)
	for name, node := range expected {
		if _, ran := executed[name]; !ran && node.hasCondPredecessor() {
			skipped[name] = true
		}
	}
	for changed := true; changed; {
		changed = false
		for name, node := range expected {
			if _, ran := executed[name]; ran || skipped[name] {
				continue
			}
			for _, dep := range node.dependents {
				if dep.Typ != nodeCondition && skipped[dep.name] {
					skipped[name] = true
					changed = true
					break
				}
			}
		}
	}
	for name := range expected {
		if _, ran := executed[name]; !ran {
			if skipped[name] {
				result.skippedBranches = append(result.skippedBranches, name)
			} else {
				result.missingTasks = append(result.missingTasks, name)
				result.valid = false
			}
		}
	}

	// --- Step 4: unexpected check ---
	for name := range executed {
		if _, defined := expected[name]; !defined {
			result.unexpectedTasks = append(result.unexpectedTasks, name)
			result.valid = false
		}
	}

	// --- Step 5: dependency check ---
	for name, ev := range executed {
		node, ok := expected[name]
		if !ok {
			continue
		}
		var expDeps []string
		for _, dep := range node.dependents {
			if _, ran := executed[dep.name]; ran {
				expDeps = append(expDeps, dep.name)
			}
		}
		var actDeps []string
		if raw := ev.Args["dependents"]; raw != "" {
			for _, d := range strings.Split(raw, ",") {
				if _, ran := executed[d]; ran {
					actDeps = append(actDeps, d)
				}
			}
		}
		if !stringSliceEqual(expDeps, actDeps) {
			result.dependencyErrors = append(result.dependencyErrors, dependencyError{
				task: name, expected: expDeps, actual: actDeps,
			})
			result.valid = false
		}
	}

	return result
}

// stringSliceEqual checks if two string slices contain the same elements (order-independent).
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	count := make(map[string]int, len(a))
	for _, s := range a {
		count[s]++
	}
	for _, s := range b {
		if count[s]--; count[s] < 0 {
			return false
		}
	}
	return true
}

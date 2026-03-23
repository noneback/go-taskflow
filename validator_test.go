package gotaskflow

import (
	"strings"
	"testing"
)

func TestValidatorSimpleSerial(t *testing.T) {
	executor := NewExecutor(4, WithTracer())
	tf := NewTaskFlow("serial")

	A := tf.NewTask("A", func() {})
	B := tf.NewTask("B", func() {})
	C := tf.NewTask("C", func() {})

	A.Precede(B)
	B.Precede(C)

	executor.Run(tf).Wait()

	result := validate(mustSnapshot(executor), tf)
	if !result.valid {
		t.Errorf("expected valid, got: %s", result.String())
	}
}

func TestValidatorParallel(t *testing.T) {
	executor := NewExecutor(4, WithTracer())
	tf := NewTaskFlow("parallel")

	// Fan-out: A -> B, C in parallel -> D (fan-in)
	A := tf.NewTask("A", func() {})
	B := tf.NewTask("B", func() {})
	C := tf.NewTask("C", func() {})
	D := tf.NewTask("D", func() {})

	A.Precede(B, C)
	B.Precede(D)
	C.Precede(D)

	executor.Run(tf).Wait()

	result := validate(mustSnapshot(executor), tf)
	if !result.valid {
		t.Errorf("expected valid, got: %s", result.String())
	}
}

func TestValidatorSubflow(t *testing.T) {
	executor := NewExecutor(4, WithTracer())
	tf := NewTaskFlow("with_subflow")

	A := tf.NewTask("A", func() {})
	sub := tf.NewSubflow("sub", func(sf *Subflow) {
		S1 := sf.NewTask("S1", func() {})
		S2 := sf.NewTask("S2", func() {})
		S1.Precede(S2)
	})
	B := tf.NewTask("B", func() {})

	A.Precede(sub)
	sub.Precede(B)

	executor.Run(tf).Wait()

	result := validate(mustSnapshot(executor), tf)
	if !result.valid {
		t.Errorf("expected valid, got: %s", result.String())
	}
}

func TestValidatorConditionBranch(t *testing.T) {
	executor := NewExecutor(4, WithTracer())
	tf := NewTaskFlow("condition")

	A := tf.NewTask("A", func() {})
	cond := tf.NewCondition("cond", func() uint { return 0 }) // always choose branch 0
	B := tf.NewTask("B", func() {})                           // branch 0 - will execute
	C := tf.NewTask("C", func() {})                           // branch 1 - will skip
	D := tf.NewTask("D", func() {})

	A.Precede(cond)
	cond.Precede(B, C) // 0 -> B, 1 -> C
	B.Precede(D)

	executor.Run(tf).Wait()

	result := validate(mustSnapshot(executor), tf)
	if !result.valid {
		t.Errorf("expected valid (C should be skipped branch), got: %s", result.String())
	}

	// C should be in skipped branches
	if !containsStr(result.skippedBranches, "C") {
		t.Errorf("expected C in skipped branches, got: %v", result.skippedBranches)
	}
}

func TestValidatorWithoutTracer(t *testing.T) {
	executor := NewExecutor(4) // no tracer
	tf := NewTaskFlow("no_tracer")

	A := tf.NewTask("A", func() {})
	B := tf.NewTask("B", func() {})
	A.Precede(B)

	executor.Run(tf).Wait()

	// nil record (no tracer) should be treated as always valid
	result := validate(mustSnapshot(executor), tf)
	if !result.valid {
		t.Errorf("expected valid (no tracer = nil record), got: %s", result.String())
	}
}

func TestValidatorComplexPipeline(t *testing.T) {
	executor := NewExecutor(8, WithTracer())
	tf := NewTaskFlow("complex")

	// prepare -> (read_config || load_data) -> validate -> check(cond) -> sub_process -> report
	prepare := tf.NewTask("prepare", func() {})
	readConfig := tf.NewTask("read_config", func() {})
	loadData := tf.NewTask("load_data", func() {})
	validateTask := tf.NewTask("validate", func() {})
	check := tf.NewCondition("check", func() uint { return 0 })
	subProcess := tf.NewSubflow("sub_process", func(sf *Subflow) {
		transform := sf.NewTask("transform", func() {})
		enrich := sf.NewTask("enrich", func() {})
		aggregate := sf.NewTask("aggregate", func() {})
		transform.Precede(aggregate)
		enrich.Precede(aggregate)
	})
	fallback := tf.NewTask("fallback", func() {}) // skipped
	report := tf.NewTask("report", func() {})

	prepare.Precede(readConfig, loadData)
	readConfig.Precede(validateTask)
	loadData.Precede(validateTask)
	validateTask.Precede(check)
	check.Precede(subProcess, fallback)
	subProcess.Precede(report)

	executor.Run(tf).Wait()

	result := validate(mustSnapshot(executor), tf)
	if !result.valid {
		t.Errorf("expected valid, got: %s", result.String())
	}

	// fallback should be skipped
	if !containsStr(result.skippedBranches, "fallback") {
		t.Errorf("expected fallback in skipped branches, got: %v", result.skippedBranches)
	}
}

// TestValidatorConditionBranch1 verifies that branch 1 executes and branch 0 is skipped.
func TestValidatorConditionBranch1(t *testing.T) {
	executor := NewExecutor(4, WithTracer())
	tf := NewTaskFlow("cond_branch1")

	cond := tf.NewCondition("cond", func() uint { return 1 }) // always branch 1
	branch0 := tf.NewTask("branch0", func() {})               // skipped
	branch1 := tf.NewTask("branch1", func() {})               // executed
	cond.Precede(branch0, branch1)

	executor.Run(tf).Wait()

	result := validate(mustSnapshot(executor), tf)
	if !result.valid {
		t.Errorf("expected valid, got: %s", result.String())
	}

	if !containsStr(result.skippedBranches, "branch0") {
		t.Errorf("expected branch0 in skipped branches, got: %v", result.skippedBranches)
	}
	if containsStr(result.skippedBranches, "branch1") {
		t.Errorf("branch1 should have executed, not skipped")
	}
}

// TestValidatorNestedSubflow verifies validation with a subflow nested inside another subflow.
func TestValidatorNestedSubflow(t *testing.T) {
	executor := NewExecutor(4, WithTracer())
	tf := NewTaskFlow("nested_subflow")

	outer := tf.NewSubflow("outer", func(sf *Subflow) {
		inner := sf.NewSubflow("inner", func(sf2 *Subflow) {
			x := sf2.NewTask("X", func() {})
			y := sf2.NewTask("Y", func() {})
			x.Precede(y)
		})
		z := sf.NewTask("Z", func() {})
		inner.Precede(z)
	})
	end := tf.NewTask("end", func() {})
	outer.Precede(end)

	executor.Run(tf).Wait()

	result := validate(mustSnapshot(executor), tf)
	if !result.valid {
		t.Errorf("expected valid for nested subflow, got: %s", result.String())
	}
}

// TestValidatorIndependentTasks verifies that fully independent tasks (no edges) are all validated.
func TestValidatorIndependentTasks(t *testing.T) {
	executor := NewExecutor(4, WithTracer())
	tf := NewTaskFlow("independent")

	tf.NewTask("T1", func() {})
	tf.NewTask("T2", func() {})
	tf.NewTask("T3", func() {})

	executor.Run(tf).Wait()

	result := validate(mustSnapshot(executor), tf)
	if !result.valid {
		t.Errorf("expected valid, got: %s", result.String())
	}
	if len(result.missingTasks) > 0 {
		t.Errorf("expected no missing tasks, got: %v", result.missingTasks)
	}
}

// TestValidatorMissingTask verifies that tasks defined but never executed appear in missingTasks.
// We run tf1 (A->B), then validate against tf2 (A->B->C). C is never executed → missing.
func TestValidatorMissingTask(t *testing.T) {
	executor := NewExecutor(4, WithTracer())

	tf1 := NewTaskFlow("run")
	a1 := tf1.NewTask("A", func() {})
	b1 := tf1.NewTask("B", func() {})
	a1.Precede(b1)
	executor.Run(tf1).Wait()

	// Validate against a larger DAG that expects C (never ran)
	tf2 := NewTaskFlow("expected")
	a2 := tf2.NewTask("A", func() {})
	b2 := tf2.NewTask("B", func() {})
	c2 := tf2.NewTask("C", func() {})
	a2.Precede(b2)
	b2.Precede(c2)

	result := validate(mustSnapshot(executor), tf2)
	if result.valid {
		t.Error("expected invalid: C was defined but not executed")
	}
	if !containsStr(result.missingTasks, "C") {
		t.Errorf("expected C in missing tasks, got: %v", result.missingTasks)
	}
}

// TestValidatorUnexpectedTask verifies that tasks executed but not defined appear in unexpectedTasks.
// We run tf1 (A->B->C), then validate against tf2 (A->B only). C is unexpected.
func TestValidatorUnexpectedTask(t *testing.T) {
	executor := NewExecutor(4, WithTracer())

	tf1 := NewTaskFlow("run")
	a1 := tf1.NewTask("A", func() {})
	b1 := tf1.NewTask("B", func() {})
	c1 := tf1.NewTask("C", func() {})
	a1.Precede(b1)
	b1.Precede(c1)
	executor.Run(tf1).Wait()

	// Validate against a smaller DAG that doesn't know about C
	tf2 := NewTaskFlow("expected")
	a2 := tf2.NewTask("A", func() {})
	b2 := tf2.NewTask("B", func() {})
	a2.Precede(b2)

	result := validate(mustSnapshot(executor), tf2)
	if result.valid {
		t.Error("expected invalid: C was executed but not defined")
	}
	if !containsStr(result.unexpectedTasks, "C") {
		t.Errorf("expected C in unexpected tasks, got: %v", result.unexpectedTasks)
	}
}

// TestValidatorResultString verifies the String() output format.
func TestValidatorResultString(t *testing.T) {
	r := &validationResult{
		valid:            false,
		missingTasks:     []string{"X"},
		unexpectedTasks:  []string{"Y"},
		dependencyErrors: []dependencyError{{task: "Z", expected: []string{"A"}, actual: []string{"B"}}},
		skippedBranches:  []string{"W"},
	}
	s := r.String()
	for _, want := range []string{"X", "Y", "Z", "W", "validation failed"} {
		if !containsSubstr(s, want) {
			t.Errorf("expected %q in String() output, got:\n%s", want, s)
		}
	}

	// Valid result should return short string
	valid := &validationResult{valid: true}
	if valid.String() != "validation passed" {
		t.Errorf("unexpected valid string: %q", valid.String())
	}
}

// ---- helpers ----

// mustSnapshot extracts a traceRecord from an executor.
// Returns nil if the executor was not created with WithTracer().
func mustSnapshot(e Executor) traceRecord {
	impl := e.(*innerExecutorImpl)
	if impl.tracer == nil {
		return nil
	}
	return impl.tracer.snapshot()
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func containsSubstr(s, sub string) bool {
	return strings.Contains(s, sub)
}

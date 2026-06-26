package main

import (
	"fmt"
	"log"
	"os"

	gotaskflow "github.com/noneback/go-taskflow"
)

func main() {
	executor := gotaskflow.NewExecutor(4, gotaskflow.WithProfiler())
	tf := gotaskflow.NewTaskFlow("ci-pipeline")

	// Stage 1: Build — parallel compile of modules
	build := tf.NewSubflow("build", func(sf *gotaskflow.Subflow) {
		frontend := sf.NewTask("compile_frontend", func() {
			fmt.Println("  compiling frontend...")
		})
		backend := sf.NewTask("compile_backend", func() {
			fmt.Println("  compiling backend...")
		})
		common := sf.NewTask("compile_common", func() {
			fmt.Println("  compiling shared libs...")
		})
		link := sf.NewTask("link", func() {
			fmt.Println("  linking binaries...")
		})
		frontend.Precede(link)
		backend.Precede(link)
		common.Precede(link)
	})

	// Stage 2: Test — parallel test suites
	test := tf.NewSubflow("test", func(sf *gotaskflow.Subflow) {
		unit := sf.NewTask("unit_test", func() {
			fmt.Println("  running unit tests...")
		})
		integration := sf.NewTask("integration_test", func() {
			fmt.Println("  running integration tests...")
		})
		e2e := sf.NewTask("e2e_test", func() {
			fmt.Println("  running e2e tests...")
		})
		report := sf.NewTask("test_report", func() {
			fmt.Println("  generating test report...")
		})
		unit.Precede(report)
		integration.Precede(report)
		e2e.Precede(report)
	})

	// Stage 3: Deploy
	deploy := tf.NewTask("deploy", func() {
		fmt.Println("deploying to production")
	})

	build.Precede(test)
	test.Precede(deploy)

	executor.Run(tf).Wait()

	if err := tf.Dump(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

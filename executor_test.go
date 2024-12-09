package gotaskflow_test

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	gotaskflow "github.com/noneback/go-taskflow"
)

func TestExecutor(t *testing.T) {
	executor := gotaskflow.NewExecutor(uint(runtime.NumCPU()))
	tf := gotaskflow.NewTaskFlow("G")
	A, B, C :=
		tf.NewTask("A", func() {
			fmt.Println("A")
		}),
		tf.NewTask("B", func() {
			fmt.Println("B")
		}),
		tf.NewTask("C", func() {
			fmt.Println("C")
		})

	A1, B1, _ :=
		tf.NewTask("A1", func() {
			fmt.Println("A1")
		}),
		tf.NewTask("B1", func() {
			fmt.Println("B1")
		}),
		tf.NewTask("C1", func() {
			fmt.Println("C1")
		})
	A.Precede(B)
	C.Precede(B)
	A1.Precede(B)
	C.Succeed(A1)
	C.Succeed(B1)

	executor.Run(tf).Wait()
	executor.Profile(os.Stdout)
}

func TestPanicInSubflow(t *testing.T) {
	executor := gotaskflow.NewExecutor(100000)
	tf := gotaskflow.NewTaskFlow("G")
	copy_gcc_source_file := tf.NewTask("copy_gcc_source_file", func() {
		time.Sleep(1 * time.Second)
		fmt.Println("copy_gcc_source_file")
	})
	tar_gcc_source_file := tf.NewTask("tar_gcc_source_file", func() {
		time.Sleep(1 * time.Second)
		fmt.Println("tar_gcc_source_file")
	})
	download_prerequisites := tf.NewSubflow("download_prerequisites", func(sf *gotaskflow.Subflow) {
		sf.NewTask("", func() {
			time.Sleep(1 * time.Second)
			fmt.Println("download_prerequisites")
			panic(1)
		})
	})
	yum_install_dependency_package := tf.NewTask("yum_install_dependency_package", func() {
		time.Sleep(1 * time.Second)
		fmt.Println("yum_install_dependency_package")
	})
	mkdir_and_prepare_build := tf.NewTask("mkdir_and_prepare_build", func() {
		time.Sleep(1 * time.Second)
		fmt.Println("mkdir_and_prepare_build")
	})
	make_build := tf.NewTask("make_build", func() {
		time.Sleep(1 * time.Second)
		fmt.Println("make_build")
	})
	make_install := tf.NewTask("make_install", func() {
		time.Sleep(1 * time.Second)
		fmt.Println("make_install")
	})
	relink := tf.NewTask("relink", func() {
		time.Sleep(1 * time.Second)
		fmt.Println("relink")
	})
	copy_gcc_source_file.Precede(tar_gcc_source_file)
	yum_install_dependency_package.Precede(download_prerequisites)
	tar_gcc_source_file.Precede(download_prerequisites)
	download_prerequisites.Precede(mkdir_and_prepare_build)
	mkdir_and_prepare_build.Precede(make_build)
	make_build.Precede(make_install)
	make_install.Precede(relink)
	executor.Run(tf).Wait()
}

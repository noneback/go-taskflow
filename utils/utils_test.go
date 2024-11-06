package utils

import (
	"fmt"
	"testing"
)

// UnsafeToString 测试
func TestUnsafeToString(t *testing.T) {
	original := "Hello, World!"
	b := []byte(original)

	// 将字节切片转换为字符串
	s := UnsafeToString(b)

	if s != original {
		t.Errorf("Expected %q but got %q", original, s)
	}
}

// UnsafeToBytes 测试
func TestUnsafeToBytes(t *testing.T) {
	original := "Hello, World!"
	// 将字符串转换为字节切片
	b := UnsafeToBytes(original)

	if string(b) != original {
		t.Errorf("Expected %q but got %q", original, string(b))
	}
}

// RC 结构体测试
func TestRC(t *testing.T) {
	rc := NewRC()

	// 测试初始值
	if rc.Value() != 0 {
		t.Errorf("Expected count to be 0, got %d", rc.Value())
	}

	// 测试增加计数
	rc.Increase()
	if rc.Value() != 1 {
		t.Errorf("Expected count to be 1, got %d", rc.Value())
	}

	// 测试减少计数
	rc.Decrease()
	if rc.Value() != 0 {
		t.Errorf("Expected count to be 0, got %d", rc.Value())
	}

	// 测试不能减少到负值
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when decreasing below zero, but did not")
		}
	}()
	rc.Decrease() // 这里应该触发 panic
}

// 测试 Set 和负值
func TestSet(t *testing.T) {
	rc := NewRC()
	rc.Set(5)

	if rc.Value() != 5 {
		t.Errorf("Expected count to be 5, got %d", rc.Value())
	}

	// 测试负值情况
	rc.Set(-1)

	if rc.Value() != -1 {
		t.Errorf("Expected count to be -1, got %d", rc.Value())
	}
}

func TestPanic(t *testing.T) {
	f := func() {
		defer func() {
			// 使用 recover 捕获 panic
			if r := recover(); r != nil {
				fmt.Println("Recovered in causePanic:", r)
			}
			fmt.Println("1")
		}()

		fmt.Println("result")
		// panic("Atest")
	}
	f()
}

// result
// Recovered in causePanic: Atest
// 1

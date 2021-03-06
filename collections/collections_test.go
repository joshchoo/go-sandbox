package collections

import (
	"testing"
)

func TestEquals(t *testing.T) {
	t.Run("one slice is nil", func(t *testing.T) {
		var a []int
		b := []int{}
		res := Equals(a, b)
		exp := false
		if res != exp {
			t.Errorf("Expected %v, got %v", exp, res)
		}
	})

	t.Run("both slices are nil", func(t *testing.T) {
		var a []int
		var b []int
		res := Equals(a, b)
		exp := true
		if res != exp {
			t.Errorf("Expected %v, got %v", exp, res)
		}
	})

	t.Run("both slices are non-nil", func(t *testing.T) {
		a := []int{}
		b := []int{}
		res := Equals(a, b)
		exp := true
		if res != exp {
			t.Errorf("Expected %v, got %v", exp, res)
		}
	})

	t.Run("int slices are equal", func(t *testing.T) {
		a := []int{1, 2, 3}
		b := []int{1, 2, 3}
		res := Equals(a, b)
		exp := true
		if res != exp {
			t.Errorf("Expected %v, got %v", exp, res)
		}
	})

	t.Run("int slices are not equal", func(t *testing.T) {
		a := []int{1, 2, 3}
		b := []int{3, 4, 5}
		res := Equals(a, b)
		exp := false
		if res != exp {
			t.Errorf("Expected %v, got %v", exp, res)
		}
	})

	t.Run("string slices are equal", func(t *testing.T) {
		a := []string{"a", "b", "c"}
		b := []string{"a", "b", "c"}
		res := Equals(a, b)
		exp := true
		if res != exp {
			t.Errorf("Expected %v, got %v", exp, res)
		}
	})

	t.Run("string slices are not equal", func(t *testing.T) {
		a := []string{"a", "b", "c"}
		b := []string{"x", "y", "z"}
		res := Equals(a, b)
		exp := false
		if res != exp {
			t.Errorf("Expected %v, got %v", exp, res)
		}
	})
}

func TestMap(t *testing.T) {
	t.Run("int slice", func(t *testing.T) {
		s := []int{1, 2, 3, 4}
		res := Map(s, func(e int) int {
			return e * 2
		})

		exp := []int{2, 4, 6, 8}
		if !Equals(res, exp) {
			t.Errorf("Expected %v, got %v", exp, res)
		}
	})

	t.Run("string slice", func(t *testing.T) {
		s := []string{"args", "bools", "characters", "deltas"}
		res := Map(s, func(e string) int {
			return len(e)
		})

		exp := []int{4, 5, 10, 6}
		if !Equals(res, exp) {
			t.Errorf("Expected %v, got %v", exp, res)
		}
	})
}

func TestFilter(t *testing.T) {
	t.Run("int slice", func(t *testing.T) {
		s := []int{1, 2, 3, 4}
		isEven := func(e int) bool {
			return e%2 == 0
		}
		res := Filter(s, isEven)

		exp := []int{2, 4}
		if !Equals(res, exp) {
			t.Errorf("Expected %v, got %v", exp, res)
		}
	})

	t.Run("string slice", func(t *testing.T) {
		s := []string{"a", "defg", "bcd", "hi", "lmnop"}
		isShort := func(e string) bool {
			return len(e) <= 3
		}
		res := Filter(s, isShort)

		exp := []string{"a", "bcd", "hi"}
		if !Equals(res, exp) {
			t.Errorf("Expected %v, got %v", exp, res)
		}
	})
}

func TestMax(t *testing.T) {
	t.Run("int slice", func(t *testing.T) {
		s := []int{8, 2, -1, 11, 3}

		res, err := Max(s)

		if err != nil {
			t.Errorf("Expected no error, got %q", err)
		}
		exp := 11
		if res != exp {
			t.Errorf("Expected %v, got %v", exp, res)
		}
	})

	t.Run("float slice", func(t *testing.T) {
		s := []float64{8.1, 2.3, -1.2, 11.34, 3.45}

		res, err := Max(s)

		if err != nil {
			t.Errorf("Expected no error, got %q", err)
		}
		exp := 11.34
		if res != exp {
			t.Errorf("Expected %v, got %v", exp, res)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		var s []int

		_, err := Max(s)

		if err != ErrEmptySlice {
			t.Errorf("Expected %q, got nil", ErrEmptySlice)
		}
	})
}

func TestMin(t *testing.T) {
	t.Run("int slice", func(t *testing.T) {
		s := []int{8, 2, -1, 11, 3}

		res, err := Min(s)

		if err != nil {
			t.Errorf("Expected no error, got %q", err)
		}
		exp := -1
		if res != exp {
			t.Errorf("Expected %v, got %v", exp, res)
		}
	})

	t.Run("float slice", func(t *testing.T) {
		s := []float64{8.1, 2.3, -1.2, 11.34, 3.45}

		res, err := Min(s)

		if err != nil {
			t.Errorf("Expected no error, got %q", err)
		}
		exp := -1.2
		if res != exp {
			t.Errorf("Expected %v, got %v", exp, res)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		var s []int

		_, err := Min(s)

		if err != ErrEmptySlice {
			t.Errorf("Expected %q, got nil", ErrEmptySlice)
		}
	})
}

package datastructures

import "testing"

func TestBinarySearch(t *testing.T) {
	tests := []struct {
		name       string
		collection []int64
		target     int64
		want       int
	}{
		{
			name:       "target can be found",
			collection: []int64{1, 2, 3, 4, 5, 6, 7, 8},
			target:     5,
			want:       4,
		},
		{
			name:       "target cannot be found",
			collection: []int64{1, 2, 3, 4, 5, 6, 7, 8},
			target:     12,
			want:       -9,
		},
		{
			name:       "target cannot be found #2",
			collection: []int64{1, 2, 3, 5, 6, 7, 8},
			target:     4,
			want:       -4,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := binarySearchInt64(test.collection, test.target)
			if got != test.want {
				t.Errorf("want: %d, got: %d", test.want, got)
			}
		})
	}
}

func binarySearchInt64(collection []int64, target int64) (foundIdx int) {
	left := 0
	right := len(collection) - 1
	for left <= right {
		mid := (left + right) / 2
		midVal := collection[mid]
		if midVal < target {
			left = mid + 1
		} else if midVal > target {
			right = mid - 1
		} else {
			return mid
		}
	}
	// Returns -(index where the number would be inserted)-1
	return -left - 1
}

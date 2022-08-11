package utils

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPHeap(t *testing.T) {
	rpheap := &Heap{}
	numbers := []int{10, 4, 3, 2, 5, 1}
	for _, number := range numbers {
		rpheap.Insert(int64(number))
	}

	sort.Ints(numbers)

	for _, number := range numbers {
		m := rpheap.DeleteMin()
		assert.Equal(t, int64(number), m, "number:%v m:%v", number, m)
	}

	assert.Equal(t, 0, rpheap.Size(), "rpheap not empty")

	runTestMeld([]int{2, 8, 5, 7}, []int{4, 9, 6}, t)
	runTestMeld([]int{4, 9, 6}, []int{2, 8, 5, 7}, t)
	runTestMeld([]int{2}, []int{4, 9, 6}, t)
	runTestMeld([]int{2, 8, 5, 7}, []int{4}, t)
	runTestMeld([]int{2, 8, 5, 7}, []int{}, t)
	runTestMeld([]int{}, []int{4, 9, 6}, t)

}

func runTestMeld(arr1, arr2 []int, t *testing.T) {
	ans := append(arr1, arr2...)
	sort.Ints(ans)

	rpheap1 := &Heap{}
	rpheap2 := &Heap{}
	for _, number := range arr1 {
		rpheap1.Insert(int64(number))
	}
	assert.Equal(t, rpheap1.Size(), len(arr1), "rpheap1 size not match")
	for _, number := range arr2 {
		rpheap2.Insert(int64(number))
	}
	assert.Equal(t, rpheap2.Size(), len(arr2), "rpheap2 size not match")

	rpheap1.Meld(rpheap2)

	assert.Equal(t, 0, rpheap2.Size(), "rpheap2 not empty")
	assert.Equal(t, rpheap1.Size(), len(ans), "rpheap1 size not match")
	for _, number := range ans {
		m := rpheap1.DeleteMin()
		assert.Equal(t, int64(number), m, "number:%v m:%v", number, m)
	}

	assert.Equal(t, 0, rpheap1.Size(), "rpheap1 not empty")
}

package set_test

import (
	"bftkvstore/set"
	"fmt"
	"testing"
)

func TestFromSlice(t *testing.T) {
	slice := []int{3, 1, 2, 3, 2}
	setA := set.FromSlice[int](slice)

	if fmt.Sprint(setA) != fmt.Sprint([]int{1, 2, 3}) {
		fmt.Println(setA, "should be", fmt.Sprint([]int{1, 2, 3}))
		t.FailNow()
	}
}

func TestUnion(t *testing.T) {
	sliceA := []int{1, 2, 3}
	sliceB := []int{3, 4}
	setA := set.FromSlice[int](sliceA)
	setB := set.FromSlice[int](sliceB)

	expectedResult := []int{1, 2, 3, 4}
	unionResult := set.Union(setA, setB)
	if fmt.Sprint(unionResult) != fmt.Sprint(expectedResult) {
		t.Error("Union of", setA, "with", setB, "should be", expectedResult, "but is", unionResult)
	}

	set.Union(setA, setB)
	if fmt.Sprint(setA) != fmt.Sprint(sliceA) {
		t.Error("After an union of", setA, "with", setB, "setA should be", sliceA, "but is", setA)
	}
	if fmt.Sprint(setB) != fmt.Sprint(sliceB) {
		t.Error("After an union of", setA, "with", setB, "setB should be", sliceB, "but is", setB)
	}
}

func TestDiff(t *testing.T) {
	sliceA := []int{1, 2, 3}
	sliceB := []int{3, 4}
	setA := set.FromSlice[int](sliceA)
	setB := set.FromSlice[int](sliceB)

	expectedResult := []int{1, 2}
	diffResult := set.Diff(setA, setB)
	if fmt.Sprint(diffResult) != fmt.Sprint(expectedResult) {
		t.Error("Diff of", setA, "with", setB, "should be", expectedResult, "but is", diffResult)
	}

	set.Diff(setA, setB)
	if fmt.Sprint(setA) != fmt.Sprint(sliceA) {
		t.Error("After a diff of", setA, "with", setB, "setA should be", sliceA, "but is", setA)
	}
	if fmt.Sprint(setB) != fmt.Sprint(sliceB) {
		t.Error("After a diff of", setA, "with", setB, "setB should be", sliceB, "but is", setB)
	}
}

func TestIntersect(t *testing.T) {
	sliceA := []int{1, 2, 3}
	sliceB := []int{3, 4}
	setA := set.FromSlice[int](sliceA)
	setB := set.FromSlice[int](sliceB)

	expectedResult := []int{3}
	intersectResult := set.Intersect(setA, setB)
	if fmt.Sprint(intersectResult) != fmt.Sprint(expectedResult) {
		t.Error("Intesect of", setA, "with", setB, "should be", expectedResult, "but is", intersectResult)
	}

	set.Intersect(setA, setB)
	if fmt.Sprint(setA) != fmt.Sprint(sliceA) {
		t.Error("After an intersect of", setA, "with", setB, "setA should be", sliceA, "but is", setA)
	}
	if fmt.Sprint(setB) != fmt.Sprint(sliceB) {
		t.Error("After an intersect of", setA, "with", setB, "setB should be", sliceB, "but is", setB)
	}
}


package test

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

// Selection Sort Algorithm Section

// This is the function that actually performs the concurrent sorting
func concurrentSelectionSort(arr []int, wg *sync.WaitGroup, ch chan []int) {
	defer wg.Done()

	n := len(arr)

	for i := 0; i < n-1; i++ {
		minIndex := i

		for j := i + 1; j < n; j++ {
			if arr[j] < arr[minIndex] {
				minIndex = j
			}
		}

		// Swap the found minimum element with the element at index i
		arr[i], arr[minIndex] = arr[minIndex], arr[i]
	}

	// Notify that the segment is sorted
	ch <- arr
}

// This function orchestrates the sorting process and is where the number of segments is passed
func concurrentSort(arr []int, numSegments int) []int {
	var wg sync.WaitGroup
	ch := make(chan []int, numSegments)

	// Calculate the segment size based on the size of the array
	segmentSize := len(arr) / numSegments

	//fmt.Println("Original Array:", arr)

	// Run concurrent Selection Sort for each segment
	for i := 0; i < numSegments; i++ {
		wg.Add(1)
		startIndex := i * segmentSize
		endIndex := (i + 1) * segmentSize
		go concurrentSelectionSort(arr[startIndex:endIndex], &wg, ch)
	}

	// Wait for all segments to be sorted
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Merge the sorted segments
	sortedArray := mergeSortedSegments(ch, numSegments)

	return sortedArray
}

// This function is what contructs the final result by calling the "mergeSortedSlices" function
// for each segment
func mergeSortedSegments(ch chan []int, numSegments int) []int {
	//fmt.Println("Number of segments", numSegments)
	// Initialize the mergedArray with an empty slice
	mergedArray := []int{}

	// Collect the sorted segments from the channel
	for i := 0; i < numSegments; i++ {
		sortedSegment := <-ch
		mergedArray = mergeSortedSlices(mergedArray, sortedSegment)
	}

	return mergedArray
}

// Function to generate random array based on specified size
func generateRandomArray(size int) ([]int, []int) {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	arr := make([]int, size)
	for i := 0; i < size; i++ {
		arr[i] = r.Intn(1000)
	}

	// Create two copies of the array for testing the algorithms
	copy1 := make([]int, size)
	copy2 := make([]int, size)
	copy(copy1, arr)
	copy(copy2, arr)

	return copy1, copy2
}

//Merge Sort Algorithm Section

// This function is what combines two sorted slices into one sorted slice
// It is called during the execution of both sorting algorithms and serves the same purpose
func mergeSortedSlices(left, right []int) []int {

	result := make([]int, len(left)+len(right))

	i, j, k := 0, 0, 0

	//Merge the elements from both slices into result
	for i < len(left) && j < len(right) {
		if left[i] <= right[j] {
			result[k] = left[i]
			i++
		} else {
			result[k] = right[j]
			j++
		}
		k++
	}

	//copies remaining elements from slice left
	for i < len(left) {
		result[k] = left[i]
		i++
		k++
	}

	//copies remaining elements from right slice
	for j < len(right) {
		result[k] = right[j]
		j++
		k++
	}

	return result
}

func MergeSort(arr []int, ch chan []int) {
	//checks if slice is already sorted
	if len(arr) <= 1 {
		ch <- arr
		return
	}
	mid := len(arr) / 2

	leftCh := make(chan []int)
	rightCh := make(chan []int)

	//This is where the recusive aspect of merge sort is implemented
	go MergeSort(arr[:mid], leftCh)
	go MergeSort(arr[mid:], rightCh)

	//receive results from both halves
	left := <-leftCh
	right := <-rightCh

	//close the channels
	close(leftCh)
	close(rightCh)

	//merges the sorted halves
	ch <- mergeSortedSlices(left, right)
}

// This function is used to initialise and execute the merge sort algorithm
func runParallelMergeSort(array []int) {
	//runtime.GOMAXPROCS(runtime.NumCPU())

	var wg sync.WaitGroup
	ch := make(chan []int)

	//startTime := time.Now()

	go MergeSort(array, ch)

	//This go routine is needed to wait for the results to prevent deadlocks
	wg.Add(1)
	go func() {
		defer wg.Done()
		sortedArray := <-ch
		_ = sortedArray
		//fmt.Println("Sorted array", sortedArray[1])
	}()
	wg.Wait()

	//_ = sortedArray
	//elapsedTime := time.Since(startTime)
	//fmt.Printf("Parallel Merge Sort took %d\n", elapsedTime.Microseconds())
}

var ArraySelectionSort, ArrayMergeSort []int

func init() {
	ArraySelectionSort, ArrayMergeSort = generateRandomArray(1000)
}

func BenchmarkTestA(b *testing.B) {
	numSegments := 4
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = concurrentSort(ArraySelectionSort, numSegments)
	}
}

func BenchmarkTestB(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		runParallelMergeSort(ArrayMergeSort)
	}
}

package models

import (
	"sync"
	"testing"
	"time"
)

func TestNewQueue(t *testing.T) {
	q := NewQueue()
	if q == nil {
		t.Fatal("NewQueue() returned nil")
	}
	if !q.IsEmpty() {
		t.Error("New queue should be empty")
	}
	if q.Len() != 0 {
		t.Errorf("New queue length = %d, want 0", q.Len())
	}
}

func TestQueue_Add(t *testing.T) {
	q := NewQueue()

	tracks := []Track{
		{ID: "1", Title: "Song 1"},
		{ID: "2", Title: "Song 2"},
	}

	q.Add(tracks)

	if q.Len() != 2 {
		t.Errorf("Queue length = %d, want 2", q.Len())
	}

	list := q.List()
	if len(list) != 2 {
		t.Errorf("List length = %d, want 2", len(list))
	}
	if list[0].Title != "Song 1" {
		t.Errorf("First track title = %v, want Song 1", list[0].Title)
	}
	if list[1].Title != "Song 2" {
		t.Errorf("Second track title = %v, want Song 2", list[1].Title)
	}
}

func TestQueue_AddMultiple(t *testing.T) {
	q := NewQueue()

	q.Add([]Track{{ID: "1", Title: "Song 1"}})
	q.Add([]Track{{ID: "2", Title: "Song 2"}})
	q.Add([]Track{{ID: "3", Title: "Song 3"}})

	if q.Len() != 3 {
		t.Errorf("Queue length = %d, want 3", q.Len())
	}
}

func TestQueue_Next(t *testing.T) {
	q := NewQueue()

	// Next from empty queue
	if q.Next() != nil {
		t.Error("Next() from empty queue should return nil")
	}

	// Add and next
	tracks := []Track{
		{ID: "1", Title: "Song 1"},
		{ID: "2", Title: "Song 2"},
	}
	q.Add(tracks)

	track := q.Next()
	if track == nil {
		t.Fatal("Next() returned nil")
	}
	if track.Title != "Song 1" {
		t.Errorf("Next() = %v, want Song 1", track.Title)
	}

	// Verify queue reduced
	if q.Len() != 1 {
		t.Errorf("Queue length = %d, want 1", q.Len())
	}

	// Next remaining
	track = q.Next()
	if track == nil {
		t.Fatal("Next() returned nil")
	}
	if track.Title != "Song 2" {
		t.Errorf("Next() = %v, want Song 2", track.Title)
	}

	// Queue should be empty now
	if q.Len() != 0 {
		t.Errorf("Queue length = %d, want 0", q.Len())
	}

	// Next from empty
	if q.Next() != nil {
		t.Error("Next() from empty queue should return nil")
	}
}

func TestQueue_Peek(t *testing.T) {
	q := NewQueue()

	// Peek from empty queue
	if q.Peek() != nil {
		t.Error("Peek() from empty queue should return nil")
	}

	// Add and peek
	tracks := []Track{
		{ID: "1", Title: "Song 1"},
		{ID: "2", Title: "Song 2"},
	}
	q.Add(tracks)

	track := q.Peek()
	if track == nil {
		t.Fatal("Peek() returned nil")
	}
	if track.Title != "Song 1" {
		t.Errorf("Peek() = %v, want Song 1", track.Title)
	}

	// Verify queue unchanged
	if q.Len() != 2 {
		t.Errorf("Queue length = %d, want 2", q.Len())
	}

	// Peek again
	track = q.Peek()
	if track.Title != "Song 1" {
		t.Errorf("Peek() = %v, want Song 1", track.Title)
	}
}

func TestQueue_Clear(t *testing.T) {
	q := NewQueue()

	q.Add([]Track{{ID: "1", Title: "Song 1"}, {ID: "2", Title: "Song 2"}})

	q.Clear()

	if !q.IsEmpty() {
		t.Error("Queue should be empty after Clear()")
	}
	if q.Len() != 0 {
		t.Errorf("Queue length = %d, want 0", q.Len())
	}
}

func TestQueue_List(t *testing.T) {
	q := NewQueue()

	// List from empty
	list := q.List()
	if len(list) != 0 {
		t.Errorf("List from empty = %d, want 0", len(list))
	}

	// Add and list
	tracks := []Track{
		{ID: "1", Title: "Song 1"},
		{ID: "2", Title: "Song 2"},
		{ID: "3", Title: "Song 3"},
	}
	q.Add(tracks)

	list = q.List()
	if len(list) != 3 {
		t.Errorf("List length = %d, want 3", len(list))
	}

	// Verify it's a copy (modifying list doesn't affect queue)
	list[0].Title = "Modified"
	if q.List()[0].Title != "Song 1" {
		t.Error("List() should return a copy, not original")
	}
}

func TestQueue_IsEmpty(t *testing.T) {
	q := NewQueue()

	if !q.IsEmpty() {
		t.Error("New queue should be empty")
	}

	q.Add([]Track{{ID: "1"}})
	if q.IsEmpty() {
		t.Error("Queue with tracks should not be empty")
	}

	q.Next()
	if !q.IsEmpty() {
		t.Error("Queue after removing all tracks should be empty")
	}
}

func TestQueue_Len(t *testing.T) {
	q := NewQueue()

	if q.Len() != 0 {
		t.Errorf("Initial length = %d, want 0", q.Len())
	}

	q.Add([]Track{{ID: "1"}})
	if q.Len() != 1 {
		t.Errorf("After add = %d, want 1", q.Len())
	}

	q.Add([]Track{{ID: "2"}, {ID: "3"}})
	if q.Len() != 3 {
		t.Errorf("After add 2 more = %d, want 3", q.Len())
	}

	q.Clear()
	if q.Len() != 0 {
		t.Errorf("After clear = %d, want 0", q.Len())
	}
}

func TestQueue_Concurrent(t *testing.T) {
	q := NewQueue()
	const numGoroutines = 100
	const opsPerGoroutine = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Add goroutines
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				q.Add([]Track{{ID: "track", Title: "Song"}})
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	// Next goroutines
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				q.Next()
				time.Sleep(time.Microsecond)
			}
		}()
	}

	// Wait for completion with timeout
	done := make(chan bool)
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Concurrent test timed out")
	}

	// Queue should be in consistent state
	if q.Len() < 0 {
		t.Error("Queue length should not be negative")
	}
}

func TestQueue_NextDoesNotPanicOnEmpty(t *testing.T) {
	q := NewQueue()

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Next() panicked on empty queue: %v", r)
		}
	}()

	for i := 0; i < 100; i++ {
		q.Next()
	}
}

func TestQueue_PeekDoesNotPanicOnEmpty(t *testing.T) {
	q := NewQueue()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Peek() panicked on empty queue: %v", r)
		}
	}()

	for i := 0; i < 100; i++ {
		q.Peek()
	}
}

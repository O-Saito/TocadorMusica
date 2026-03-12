package domain

import (
	"sync"
	"testing"
	"time"
)

func TestQueue_Enqueue_Dequeue_FIFO(t *testing.T) {
	q := NewQueue(10)

	track1 := NewTrackFromYouTube("http://1.com", "Track 1", "", "http://audio1.com")
	track2 := NewTrackFromYouTube("http://2.com", "Track 2", "", "http://audio2.com")
	track3 := NewTrackFromYouTube("http://3.com", "Track 3", "", "http://audio3.com")

	q.Enqueue(track1)
	q.Enqueue(track2)
	q.Enqueue(track3)

	if q.Size() != 3 {
		t.Errorf("expected size 3, got %d", q.Size())
	}

	d1, err := q.Dequeue()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d1.Title() != "Track 1" {
		t.Errorf("expected Track 1, got %s", d1.Title())
	}

	d2, err := q.Dequeue()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d2.Title() != "Track 2" {
		t.Errorf("expected Track 2, got %s", d2.Title())
	}

	d3, err := q.Dequeue()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d3.Title() != "Track 3" {
		t.Errorf("expected Track 3, got %s", d3.Title())
	}

	if !q.IsEmpty() {
		t.Error("expected queue to be empty")
	}
}

func TestQueue_Enqueue_QueueFull(t *testing.T) {
	q := NewQueue(3)

	q.Enqueue(NewTrackFromYouTube("http://1.com", "Track 1", "", "http://audio1.com"))
	q.Enqueue(NewTrackFromYouTube("http://2.com", "Track 2", "", "http://audio2.com"))
	q.Enqueue(NewTrackFromYouTube("http://3.com", "Track 3", "", "http://audio3.com"))

	err := q.Enqueue(NewTrackFromYouTube("http://4.com", "Track 4", "", "http://audio4.com"))
	if err == nil {
		t.Error("expected error when queue is full")
	}
	if err != ErrQueueFull {
		t.Errorf("expected ErrQueueFull, got %v", err)
	}
}

func TestQueue_Peek(t *testing.T) {
	q := NewQueue(10)

	track := NewTrackFromYouTube("http://1.com", "Track 1", "", "http://audio1.com")
	q.Enqueue(track)

	peeked, err := q.Peek()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if peeked.Title() != "Track 1" {
		t.Errorf("expected Track 1, got %s", peeked.Title())
	}

	if q.Size() != 1 {
		t.Error("size should not change after peek")
	}
}

func TestQueue_Peek_Empty(t *testing.T) {
	q := NewQueue(10)

	_, err := q.Peek()
	if err == nil {
		t.Error("expected error when peeking empty queue")
	}
	if err != ErrQueueEmpty {
		t.Errorf("expected ErrQueueEmpty, got %v", err)
	}
}

func TestQueue_Dequeue_Empty(t *testing.T) {
	q := NewQueue(10)

	_, err := q.Dequeue()
	if err == nil {
		t.Error("expected error when dequeuing empty queue")
	}
	if err != ErrQueueEmpty {
		t.Errorf("expected ErrQueueEmpty, got %v", err)
	}
}

func TestQueue_Clear(t *testing.T) {
	q := NewQueue(10)

	q.Enqueue(NewTrackFromYouTube("http://1.com", "Track 1", "", "http://audio1.com"))
	q.Enqueue(NewTrackFromYouTube("http://2.com", "Track 2", "", "http://audio2.com"))

	q.Clear()

	if !q.IsEmpty() {
		t.Error("expected queue to be empty after clear")
	}
	if q.Size() != 0 {
		t.Errorf("expected size 0, got %d", q.Size())
	}
}

func TestQueue_Remove(t *testing.T) {
	q := NewQueue(10)

	q.Enqueue(NewTrackFromYouTube("http://1.com", "Track 1", "", "http://audio1.com"))
	q.Enqueue(NewTrackFromYouTube("http://2.com", "Track 2", "", "http://audio2.com"))
	q.Enqueue(NewTrackFromYouTube("http://3.com", "Track 3", "", "http://audio3.com"))

	err := q.Remove(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if q.Size() != 2 {
		t.Errorf("expected size 2, got %d", q.Size())
	}

	first, _ := q.Dequeue()
	if first.Title() != "Track 1" {
		t.Errorf("expected Track 1 first, got %s", first.Title())
	}

	second, _ := q.Dequeue()
	if second.Title() != "Track 3" {
		t.Errorf("expected Track 3 second, got %s", second.Title())
	}
}

func TestQueue_Remove_InvalidIndex(t *testing.T) {
	q := NewQueue(10)

	q.Enqueue(NewTrackFromYouTube("http://1.com", "Track 1", "", "http://audio1.com"))

	err := q.Remove(5)
	if err == nil {
		t.Error("expected error for invalid index")
	}
	if err != ErrInvalidIndex {
		t.Errorf("expected ErrInvalidIndex, got %v", err)
	}

	err = q.Remove(-1)
	if err == nil {
		t.Error("expected error for negative index")
	}
}

func TestQueue_Remove_First(t *testing.T) {
	q := NewQueue(10)

	q.Enqueue(NewTrackFromYouTube("http://1.com", "Track 1", "", "http://audio1.com"))
	q.Enqueue(NewTrackFromYouTube("http://2.com", "Track 2", "", "http://audio2.com"))

	err := q.Remove(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	first, _ := q.Dequeue()
	if first.Title() != "Track 2" {
		t.Errorf("expected Track 2, got %s", first.Title())
	}
}

func TestQueue_Remove_Last(t *testing.T) {
	q := NewQueue(10)

	q.Enqueue(NewTrackFromYouTube("http://1.com", "Track 1", "", "http://audio1.com"))
	q.Enqueue(NewTrackFromYouTube("http://2.com", "Track 2", "", "http://audio2.com"))

	err := q.Remove(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	first, _ := q.Dequeue()
	if first.Title() != "Track 1" {
		t.Errorf("expected Track 1, got %s", first.Title())
	}

	if !q.IsEmpty() {
		t.Error("queue should be empty")
	}
}

func TestQueue_Size(t *testing.T) {
	q := NewQueue(10)

	if q.Size() != 0 {
		t.Errorf("expected size 0, got %d", q.Size())
	}

	q.Enqueue(NewTrackFromYouTube("http://1.com", "Track 1", "", "http://audio1.com"))
	if q.Size() != 1 {
		t.Errorf("expected size 1, got %d", q.Size())
	}

	q.Enqueue(NewTrackFromYouTube("http://2.com", "Track 2", "", "http://audio2.com"))
	if q.Size() != 2 {
		t.Errorf("expected size 2, got %d", q.Size())
	}

	q.Dequeue()
	if q.Size() != 1 {
		t.Errorf("expected size 1, got %d", q.Size())
	}
}

func TestQueue_IsEmpty(t *testing.T) {
	q := NewQueue(10)

	if !q.IsEmpty() {
		t.Error("expected empty queue")
	}

	q.Enqueue(NewTrackFromYouTube("http://1.com", "Track 1", "", "http://audio1.com"))

	if q.IsEmpty() {
		t.Error("expected non-empty queue")
	}
}

func TestQueue_Concurrent(t *testing.T) {
	q := NewQueue(500)

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			q.Enqueue(NewTrackFromYouTube("http://test.com", "Track", "", "http://audio.com"))
		}(i)
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q.Dequeue()
		}()
	}

	wg.Wait()

	if q.Size() < 0 || q.Size() > 500 {
		t.Errorf("unexpected size: %d", q.Size())
	}
}

func TestQueue_Concurrent_Clear_While_Reading(t *testing.T) {
	q := NewQueue(500)

	for i := 0; i < 100; i++ {
		q.Enqueue(NewTrackFromYouTube("http://test.com", "Track", "", "http://audio.com"))
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				q.Dequeue()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				q.Enqueue(NewTrackFromYouTube("http://test.com", "Track", "", "http://audio.com"))
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				q.Clear()
			}
		}
	}()

	time.Sleep(100 * time.Millisecond)
	close(stop)

	wg.Wait()

	if q.Size() > 500 {
		t.Errorf("size exceeded max: %d", q.Size())
	}
}

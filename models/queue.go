package models

import "sync"

type Queue struct {
	tracks []Track
	mutex  sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		tracks: make([]Track, 0),
	}
}

func (q *Queue) Add(tracks []Track) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.tracks = append(q.tracks, tracks...)
}

func (q *Queue) Next() *Track {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.tracks) == 0 {
		return nil
	}

	track := q.tracks[0]
	q.tracks = q.tracks[1:]
	return &track
}

func (q *Queue) Peek() *Track {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.tracks) == 0 {
		return nil
	}

	return &q.tracks[0]
}

func (q *Queue) Clear() {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.tracks = make([]Track, 0)
}

func (q *Queue) List() []Track {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	result := make([]Track, len(q.tracks))
	copy(result, q.tracks)
	return result
}

func (q *Queue) IsEmpty() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.tracks) == 0
}

func (q *Queue) Len() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.tracks)
}

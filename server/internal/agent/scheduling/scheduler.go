package scheduling

import (
	"container/heap"
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Scheduler struct {
	store       *AssignmentStore
	workerQueue chan<- RunRequest
	log         *slog.Logger
	wake        chan struct{}
	metrics     Metrics
}

type Metrics interface {
	IncScheduledRun()
	IncSkippedRun(reason string)
}

func NewScheduler(store *AssignmentStore, workerQueue chan<- RunRequest, log *slog.Logger, metrics Metrics) *Scheduler {
	return &Scheduler{
		store:       store,
		workerQueue: workerQueue,
		log:         log,
		wake:        make(chan struct{}, 1),
		metrics:     metrics,
	}
}

// Wake wakes the scheduler to check for new tasks to schedule.
func (s *Scheduler) Wake() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

// Run starts the scheduler, which schedules tasks from the assignment store.
func (s *Scheduler) Run(ctx context.Context) error {
	items := make(scheduleHeap, 0)
	heap.Init(&items)
	// Push active tasks into the heap.
	s.pushActiveTasks(&items)

	// Start the timer to check for new tasks.
	timer := time.NewTimer(time.Hour)
	defer timer.Stop()

	for {
		// if the heap is empty, wait for a wakeup or context cancellation.
		if len(items) == 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-s.wake:
				s.pushActiveTasks(&items)
				continue
			}
		}

		next := items[0]
		now := time.Now().UTC()
		// Pop the next task from the heap and dispatch it if it is due.
		if next.due.After(now) {
			resetTimer(timer, next.due.Sub(now))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-s.wake:
				s.rebuildHeap(&items)
				continue
			case <-timer.C:
			}
		}

		// Pop the next task from the heap and dispatch it if it is due.
		item := popScheduleItem(&items)
		now = time.Now().UTC()
		if !s.store.IsFresh(now) {
			resetTimer(timer, time.Second)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-s.wake:
				s.rebuildHeap(&items)
			case <-timer.C:
			}
			continue
		}

		task, ok := s.store.CurrentForSchedule(item.assignmentID, item.generation, item.due)
		if !ok {
			continue
		}
		s.dispatchOrSkip(task, item.due, now)

		nextDue := ComputeNextFutureDue(item.due, now, task.Check.IntervalTime())
		if advanced, ok := s.store.AdvanceNextDue(task.AssignmentID, task.Generation, nextDue); ok {
			heap.Push(&items, scheduleItem{
				assignmentID: advanced.AssignmentID,
				generation:   advanced.Generation,
				due:          advanced.NextDue,
			})
		}
	}
}

func (s *Scheduler) pushActiveTasks(items *scheduleHeap) {
	for _, task := range s.store.ActiveTasks() {
		heap.Push(items, scheduleItem{
			assignmentID: task.AssignmentID,
			generation:   task.Generation,
			due:          task.NextDue,
		})
	}
}

func (s *Scheduler) rebuildHeap(items *scheduleHeap) {
	*items = (*items)[:0]
	s.pushActiveTasks(items)
}

func (s *Scheduler) dispatchOrSkip(task TaskState, scheduledAt, now time.Time) {
	if s.metrics != nil {
		s.metrics.IncScheduledRun()
	}
	if IsTooLate(scheduledAt, now, task.Check.IntervalTime()) {
		if s.metrics != nil {
			s.metrics.IncSkippedRun("late")
		}
		s.log.Debug("skipped late occurrence", "assignment_id", task.AssignmentID, "check_id", task.Check.ID, "scheduled_at", scheduledAt, "interval", task.Check.IntervalTime())
		return
	}

	request := task.RunRequest(scheduledAt, now)
	select {
	case s.workerQueue <- request:
	default:
		if s.metrics != nil {
			s.metrics.IncSkippedRun("worker_queue_full")
		}
		s.log.Warn("skipped occurrence because worker queue is full", "assignment_id", task.AssignmentID, "check_id", task.Check.ID, "check_type", task.Check.Type, "scheduled_at", scheduledAt)
	}
}

type scheduleItem struct {
	assignmentID string
	generation   uint64
	due          time.Time
}

type scheduleHeap []scheduleItem

func (h scheduleHeap) Len() int {
	return len(h)
}

func (h scheduleHeap) Less(i, j int) bool {
	return h[i].due.Before(h[j].due)
}

func (h scheduleHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *scheduleHeap) Push(x any) {
	item, ok := x.(scheduleItem)
	if !ok {
		panic(fmt.Sprintf("schedule heap received %T", x))
	}
	*h = append(*h, item)
}

func (h *scheduleHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]

	return item
}

func popScheduleItem(items *scheduleHeap) scheduleItem {
	popped := heap.Pop(items)
	item, ok := popped.(scheduleItem)
	if !ok {
		panic(fmt.Sprintf("schedule heap contained %T", popped))
	}

	return item
}

func resetTimer(timer *time.Timer, duration time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(duration)
}

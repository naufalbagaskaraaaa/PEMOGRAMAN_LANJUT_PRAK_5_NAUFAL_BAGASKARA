package repository

import (
	"context"
	"sort"
	"strings"
	"sync"

	"latihan-fiber/app/model"
)

// StudentMemoryRepository is a Modul 2 implementation for tests and demos.
// The production composition root continues to use PostgreSQL.
type StudentMemoryRepository struct {
	mu       sync.RWMutex
	nextID   int
	students []model.Student
}

func NewStudentMemoryRepository(initial []model.Student) *StudentMemoryRepository {
	students := append([]model.Student(nil), initial...)
	nextID := 1
	for _, student := range students {
		if student.ID >= nextID {
			nextID = student.ID + 1
		}
	}
	return &StudentMemoryRepository{nextID: nextID, students: students}
}

func (r *StudentMemoryRepository) FindAll(_ context.Context, q model.ListQuery) ([]model.Student, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := append([]model.Student(nil), r.students...)
	filtered := items[:0]
	for _, student := range items {
		if q.Search != "" && !strings.Contains(strings.ToLower(student.Name), strings.ToLower(q.Search)) && !strings.Contains(strings.ToLower(student.NIM), strings.ToLower(q.Search)) {
			continue
		}
		if q.IsActive != nil && student.IsActive != *q.IsActive {
			continue
		}
		filtered = append(filtered, student)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		less := func() bool {
			switch q.Sort {
			case "name":
				return filtered[i].Name < filtered[j].Name
			case "nim":
				return filtered[i].NIM < filtered[j].NIM
			case "grade":
				return filtered[i].Grade < filtered[j].Grade
			default:
				return filtered[i].ID < filtered[j].ID
			}
		}()
		if q.Order == "desc" {
			return !less
		}
		return less
	})
	total := len(filtered)
	start := q.Offset()
	if start > total {
		start = total
	}
	end := start + q.Limit
	if end > total {
		end = total
	}
	return filtered[start:end], total, nil
}

func (r *StudentMemoryRepository) FindByID(_ context.Context, id int) (model.Student, error) {
	return r.find(id, 0, false)
}
func (r *StudentMemoryRepository) FindByIDOwned(_ context.Context, id, ownerID int) (model.Student, error) {
	return r.find(id, ownerID, true)
}
func (r *StudentMemoryRepository) find(id, ownerID int, owned bool) (model.Student, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, student := range r.students {
		if student.ID == id && (!owned || student.OwnerID != nil && *student.OwnerID == ownerID) {
			return student, nil
		}
	}
	return model.Student{}, ErrNotFound
}

func (r *StudentMemoryRepository) Create(_ context.Context, student model.Student) (model.Student, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, item := range r.students {
		if item.NIM == student.NIM {
			return model.Student{}, ErrDuplicate
		}
	}
	student.ID = r.nextID
	r.nextID++
	r.students = append(r.students, student)
	return student, nil
}
func (r *StudentMemoryRepository) Update(_ context.Context, student model.Student) (model.Student, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, item := range r.students {
		if item.ID == student.ID {
			for _, other := range r.students {
				if other.ID != student.ID && other.NIM == student.NIM {
					return model.Student{}, ErrDuplicate
				}
			}
			student.OwnerID = item.OwnerID
			r.students[i] = student
			return student, nil
		}
	}
	return model.Student{}, ErrNotFound
}
func (r *StudentMemoryRepository) Delete(_ context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, student := range r.students {
		if student.ID == id {
			r.students = append(r.students[:i], r.students[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

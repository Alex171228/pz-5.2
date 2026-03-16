package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
	Done        bool   `json:"done"`
	CreatedAt   string `json:"created_at,omitempty"`
}

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
	Done        *bool   `json:"done,omitempty"`
}

type TaskRepository interface {
	Create(task *Task) error
	GetAll() ([]*Task, error)
	GetByID(id string) (*Task, error)
	Update(task *Task) error
	Delete(id string) error
	SearchByTitle(title string) ([]*Task, error)
}

type TaskService struct {
	repo TaskRepository
}

func NewTaskService(repo TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) Create(req CreateTaskRequest) (*Task, error) {
	task := &Task{
		ID:          "t_" + uuid.New().String()[:8],
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		Done:        false,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	if err := s.repo.Create(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) GetAll() ([]*Task, error) {
	return s.repo.GetAll()
}

func (s *TaskService) GetByID(id string) (*Task, error) {
	return s.repo.GetByID(id)
}

func (s *TaskService) Update(id string, req UpdateTaskRequest) (*Task, error) {
	task, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.DueDate != nil {
		task.DueDate = *req.DueDate
	}
	if req.Done != nil {
		task.Done = *req.Done
	}

	if err := s.repo.Update(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *TaskService) SearchByTitle(title string) ([]*Task, error) {
	return s.repo.SearchByTitle(title)
}

package todo

import "time"

type Task struct {
	Description    string     
	Name           string    
	TimeOfCreate   time.Time  
	TimeOfComplete *time.Time 
	Complete       bool      
}

func NewTask(desc string, name string) *Task {
	now :=  time.Now()
	return &Task{
		Description:    desc,
		Name:           name,
		TimeOfCreate:  now,
		TimeOfComplete: nil,
		Complete:       false,
	}
}

func (t *Task) Done() {
	now := time.Now()
	t.Complete = true
	t.TimeOfComplete = &now
}

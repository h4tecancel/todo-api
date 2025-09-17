package handlers

import (
	"encoding/json"
	"errors"
	"time"
)

type TaskResponse struct {
	ID        int64
	Name      string
	Execution string
}


type CompleteTaskDTO struct {
	ID       int64 `json:"id"`
	Complete bool  `json:"complete"`
}

type TaskDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (t TaskDTO) ValidateForCreate() error {
	if t.Name == "" {
		return errors.New("name is empty")
	}

	if t.Description == "" {
		return errors.New("description is empty")
	}

	return nil
}

type ErrorDTO struct {
	Message string
	Time    time.Time
}

func (e ErrorDTO) ToString() string {
	b, err := json.MarshalIndent(e, "", "    ")
	if err != nil {
		panic(err)
	}

	return string(b)
}

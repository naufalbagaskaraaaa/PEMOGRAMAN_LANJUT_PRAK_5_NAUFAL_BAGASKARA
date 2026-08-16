package main

import "fmt"

type Student struct {
	ID       int
	Name     string
	Grade    float64
	IsActive bool
}

// GetInfo menggunakan value receiver karena method hanya membaca data Student.
func (s Student) GetInfo() string {
	return fmt.Sprintf(
		"ID: %d | Name: %s | Grade: %.2f | Active: %t",
		s.ID,
		s.Name,
		s.Grade,
		s.IsActive,
	)
}

// UpdateGrade menggunakan pointer receiver karena method mengubah nilai Grade pada Student.
func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

// Activate menggunakan pointer receiver karena method mengubah nilai IsActive.
func (s *Student) Activate() {
	s.IsActive = true
}

// Deactivate menggunakan pointer receiver karena method mengubah nilai IsActive.
func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {
	student := Student{
		ID:       1,
		Name:     "Naufal",
		Grade:    80.5,
		IsActive: false,
	}

	fmt.Println("DATA AWAL")
	fmt.Println(student.GetInfo())

	fmt.Println("ACTIVATE")
	student.Activate()
	fmt.Println(student.GetInfo())

	fmt.Println("UPDATE GRADE")
	student.UpdateGrade(90.0)
	fmt.Println(student.GetInfo())

	fmt.Println("DEACTIVATE")
	student.Deactivate()
	fmt.Println(student.GetInfo())
}
package struct2sql

import (
	"fmt"
	"testing"
)

type Student struct {
	StudentID   uint     `json:"student_id"`
	StudentName string   `json:"student_name"`
	Age         int      `json:"age"`
	Class       Class    `json:"class" s2sql:"relation_field:ClassID"`
	HobbyList   []Hobby  `json:"hobby_list" s2sql:"relation_field:HobbyID"`
	CourseList  []Course `json:"course_list" s2sql:"relation_field:CourseID"`
}

type Class struct {
	ClassID   uint   `json:"class_id"`
	ClassName string `json:"class_name"`
	Level     int    `json:"level"`
}

type Hobby struct {
	HobbyID   uint   `json:"hobby_id"`
	HobbyName string `json:"hobby_name"`
}

type Course struct {
	CourseID    uint      `json:"course_id"`
	StudentList []Student `json:"student_list"  s2sql:"relation_field:StudentID"`
}

func TestStruct2Sql_scan(t *testing.T) {
	s := NewStruct2Sql()
	var v, err = s.BuildInstall(Student{
		StudentID:   1,
		StudentName: "bin",
		Age:         2,
		Class: Class{
			ClassID:   1,
			ClassName: "向日葵",
			Level:     1,
		},
		HobbyList: []Hobby{
			{
				HobbyID:   1,
				HobbyName: "唱",
			},
			{
				HobbyID:   2,
				HobbyName: "跳",
			},
			{
				HobbyID:   3,
				HobbyName: "Rap",
			},
			{
				HobbyID:   4,
				HobbyName: "篮球",
			},
		},
		CourseList: []Course{
			{
				CourseID: 10,
			},
			{
				CourseID: 20,
			},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(v)
}

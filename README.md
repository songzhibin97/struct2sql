# struct to sql

使用struct的形式生成sql语句,支持1对1,1对多,多对多

指定tag进行操作, `s2sql:"-"`忽略, `s2sql:relation_field:xxx` 指定对应关系字段,用于ognl取出, `s2sql:alias:xxx` 指定生成sql列别名, struct实现 `func TableName() string` 修正table name

```go
package main

import "github.com/songzhibin97/struct2sql"

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

func main() {
	s := struct2sql.NewStruct2Sql()
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
	// [{INSERT INTO Student (Age,Class,StudentID,StudentName) VALUES (?,?,?,?) [2 1 1 bin]} {INSERT INTO Student (Age,Class,StudentID,StudentName) VALUES (?,?,?,?) [2 1 1 bin]} {INSERT INTO Student (Age,Class,StudentID,StudentName) VALUES (?,?,?,?) [2 1 1 bin]} {INSERT INTO Student (Age,Class,StudentID,StudentName) VALUES (?,?,?,?) [2 1 1 bin]} {INSERT INTO Student (Age,Class,StudentID,StudentName) VALUES (?,?,?,?) [2 1 1 bin]} {INSERT INTO Student (Age,Class,StudentID,StudentName) VALUES (?,?,?,?) [2 1 1 bin]} {INSERT INTO Student (Age,Class,StudentID,StudentName) VALUES (?,?,?,?) [2 1 1 bin]} {INSERT INTO Student (Age,Class,StudentID,StudentName) VALUES (?,?,?,?) [2 1 1 bin]} {INSERT INTO Student_Course (CourseID,StudentID) VALUES (?,?) [10 1]} {INSERT INTO Student_Course (CourseID,StudentID) VALUES (?,?) [20 1]} {INSERT INTO Student_Course (CourseID,StudentID) VALUES (?,?) [10 1]} {INSERT INTO Student_Course (CourseID,StudentID) VALUES (?,?) [20 1]} {INSERT INTO Student_Course (CourseID,StudentID) VALUES (?,?) [10 1]} {INSERT INTO Student_Course (CourseID,StudentID) VALUES (?,?) [20 1]} {INSERT INTO Student_Course (CourseID,StudentID) VALUES (?,?) [10 1]} {INSERT INTO Student_Course (CourseID,StudentID) VALUES (?,?) [20 1]}]
}

```
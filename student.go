package main

import "strconv"

func (student *Student) GetIdString() string {
	return strconv.FormatUint(uint64(student.Id), 10)
}

func (student *Student) GetNamePrefix() string {
	if student.Gender == Student_FEMALE {
		return "Пані"
	}

	return "Пане"
}

func (student *Student) GetTemplateData() StudentMessageData {
	return StudentMessageData{
		NamePrefix: student.GetNamePrefix(),
		Name:       student.FirstName,
	}
}

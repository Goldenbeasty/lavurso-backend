package data

import "database/sql"

type Models struct {
	Users       UserModel
	Classes     ClassModel
	Subjects    SubjectModel
	Journals    JournalModel
	Lessons     LessonModel
	Assignments AssignmentModel
}

func NewModel(db *sql.DB) Models {
	return Models{
		Users:       UserModel{DB: db},
		Classes:     ClassModel{DB: db},
		Subjects:    SubjectModel{DB: db},
		Journals:    JournalModel{DB: db},
		Lessons:     LessonModel{DB: db},
		Assignments: AssignmentModel{DB: db},
	}
}

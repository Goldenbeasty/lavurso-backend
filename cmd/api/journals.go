package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/annusingmar/lavurso-backend/internal/validator"
	"github.com/go-chi/chi/v5"
)

func (app *application) listAllJournals(w http.ResponseWriter, r *http.Request) {
	var archived bool

	archivedParam := r.URL.Query().Get("archived")

	if archivedParam == "true" {
		archived = true
	} else {
		archived = false
	}

	journals, err := app.models.Journals.AllJournals(archived)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	err = app.outputJSON(w, http.StatusOK, envelope{"journals": journals})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getJournal(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	journalID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if journalID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchJournal.Error())
		return
	}

	journal, err := app.models.Journals.GetJournalByID(journalID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	switch sessionUser.Role {
	case data.RoleAdministrator:
	case data.RoleTeacher:
		if journal.Teacher.ID != sessionUser.ID {
			app.notAllowed(w, r)
			return
		}
	case data.RoleParent:
		ok, err := app.models.Journals.DoesParentHaveChildInJournal(sessionUser.ID, journal.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	case data.RoleStudent:
		ok, err := app.models.Journals.IsUserInJournal(sessionUser.ID, journal.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	default:
		app.notAllowed(w, r)
		return
	}

	journal.CurrentCourse, err = app.models.Journals.GetCurrentCourseForJournal(journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"journal": journal})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) createJournal(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	var input struct {
		Name      string `json:"name"`
		SubjectID int    `json:"subject_id"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	journal := &data.Journal{
		Name:     &input.Name,
		Teacher:  &data.User{ID: sessionUser.ID},
		Subject:  &data.Subject{ID: &input.SubjectID},
		Archived: helpers.ToPtr(false),
	}

	v.Check(*journal.Name != "", "name", "must be provided")
	v.Check(*journal.Subject.ID > 0, "subject_id", "must be provided and valid")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	_, err = app.models.Subjects.GetSubjectByID(*journal.Subject.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchSubject):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.models.Journals.InsertJournal(journal)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusCreated, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
}

func (app *application) updateJournal(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	journalID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if journalID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchJournal.Error())
		return
	}

	journal, err := app.models.Journals.GetJournalByID(journalID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if journal.Teacher.ID != sessionUser.ID && sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	if *journal.Archived {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrJournalArchived.Error())
		return
	}

	var input struct {
		Name      *string `json:"name"`
		TeacherID *int    `json:"teacher_id"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if input.Name != nil {
		journal.Name = input.Name
	}
	if input.TeacherID != nil {
		if *input.TeacherID != sessionUser.ID && sessionUser.Role != data.RoleAdministrator {
			app.notAllowed(w, r)
			return
		}
		journal.Teacher.ID = *input.TeacherID
	}

	v := validator.NewValidator()

	v.Check(*journal.Name != "", "name", "must be provided")
	v.Check(journal.Teacher.ID > 0, "teacher_id", "must be provided and valid")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	teacher, err := app.models.Users.GetUserByID(journal.Teacher.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if teacher.Role != data.RoleAdministrator && teacher.Role != data.RoleTeacher {
		app.writeErrorResponse(w, r, http.StatusBadRequest, "user not a teacher")
		return
	}

	err = app.models.Journals.UpdateJournal(journal)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.models.Journals.SetJournalLastUpdated(journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) deleteJournal(w http.ResponseWriter, r *http.Request) {
	journalID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if journalID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchJournal.Error())
		return
	}

	journal, err := app.models.Journals.GetJournalByID(journalID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if !*journal.Archived {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrJournalNotArchived.Error())
		return
	}

	err = app.models.Journals.DeleteJournal(journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getJournalsForTeacher(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	teacherID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if teacherID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
		return
	}

	if teacherID != sessionUser.ID && sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	var archived bool

	archivedParam := r.URL.Query().Get("archived")

	if archivedParam == "true" {
		archived = true
	} else {
		archived = false
	}

	teacher, err := app.models.Users.GetUserByID(teacherID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if teacher.Role != data.RoleAdministrator && teacher.Role != data.RoleTeacher {
		app.writeErrorResponse(w, r, http.StatusBadRequest, "user not an admin")
		return
	}

	journals, err := app.models.Journals.GetJournalsForTeacher(teacher.ID, archived)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"journals": journals})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) addStudentsToJournal(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	journalID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if journalID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchJournal.Error())
		return
	}

	journal, err := app.models.Journals.GetJournalByID(journalID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if journal.Teacher.ID != sessionUser.ID && sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	if *journal.Archived {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrJournalArchived.Error())
		return
	}

	var input struct {
		StudentIDs []int `json:"student_ids"`
		ClassIDs   []int `json:"class_ids"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	allStudentIDs, err := app.models.Users.GetAllStudentIDs()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	badIDs := helpers.VerifyExistsInSlice(input.StudentIDs, allStudentIDs)
	if badIDs != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("%s: %v", data.ErrNoSuchStudents.Error(), badIDs))
		return
	}

	allClassIDs, err := app.models.Classes.GetAllClassIDs()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	badIDs = helpers.VerifyExistsInSlice(input.ClassIDs, allClassIDs)
	if badIDs != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("%s: %v", data.ErrNoSuchClass.Error(), badIDs))
		return
	}

	for _, id := range input.StudentIDs {
		err = app.models.Journals.InsertUserIntoJournal(id, journal.ID)
		if err != nil && !errors.Is(err, data.ErrUserAlreadyInJournal) {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	for _, id := range input.ClassIDs {
		users, err := app.models.Classes.GetUsersForClassID(id)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}

		for _, u := range users {
			err = app.models.Journals.InsertUserIntoJournal(u.ID, journal.ID)
			if err != nil && !errors.Is(err, data.ErrUserAlreadyInJournal) {
				app.writeInternalServerError(w, r, err)
				return
			}
		}
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) removeStudentFromJournal(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	journalID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if journalID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchJournal.Error())
		return
	}

	journal, err := app.models.Journals.GetJournalByID(journalID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if journal.Teacher.ID != sessionUser.ID && sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	if *journal.Archived {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrJournalArchived.Error())
		return
	}

	var input struct {
		StudentID int `json:"student_id"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()
	v.Check(input.StudentID > 0, "student_id", "must be provided and valid")
	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	user, err := app.models.Users.GetUserByID(input.StudentID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if user.Role != data.RoleStudent {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
		return
	}

	err = app.models.Journals.DeleteUserFromJournal(user.ID, journal.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrUserNotInJournal):
			app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getStudentsForJournal(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	journalID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if journalID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchJournal.Error())
		return
	}

	journal, err := app.models.Journals.GetJournalByID(journalID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if journal.Teacher.ID != sessionUser.ID && sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	students, err := app.models.Journals.GetUsersByJournalID(journal.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"students": students})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getJournalsForStudent(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if userID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
		return
	}

	switch sessionUser.Role {
	case data.RoleAdministrator:
	case data.RoleParent:
		ok, err := app.models.Users.IsParentOfChild(sessionUser.ID, userID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	case data.RoleStudent:
		if sessionUser.ID != userID {
			app.notAllowed(w, r)
			return
		}
	default:
		app.notAllowed(w, r)
		return
	}

	user, err := app.models.Users.GetUserByID(userID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if user.Role != data.RoleStudent {
		app.writeErrorResponse(w, r, http.StatusBadRequest, data.ErrNotAStudent.Error())
		return
	}

	journals, err := app.models.Journals.GetJournalsByUserID(user.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"journals": journals})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}

}

func (app *application) archiveJournal(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	journalID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if journalID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchJournal.Error())
		return
	}

	journal, err := app.models.Journals.GetJournalByID(journalID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if journal.Teacher.ID != sessionUser.ID && sessionUser.Role != data.RoleAdministrator {
		app.notAllowed(w, r)
		return
	}

	err = app.models.Journals.SetJournalArchived(journal.ID, true)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) unarchiveJournal(w http.ResponseWriter, r *http.Request) {
	journalID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if journalID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchJournal.Error())
		return
	}

	journal, err := app.models.Journals.GetJournalByID(journalID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchJournal):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.models.Journals.SetJournalArchived(journal.ID, false)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

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

func (app *application) getAllYears(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	var err error
	var years []*data.Year

	if *sessionUser.Role == data.RoleAdministrator && r.URL.Query().Get("stats") == "true" {
		years, err = app.models.Years.ListAllYearsWithStats()
	} else {
		years, err = app.models.Years.ListAllYears()

	}

	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"years": years})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getYearsForStudent(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if userID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchUser.Error())
		return
	}

	student, err := app.models.Users.GetStudentByID(userID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchUser):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	if *sessionUser.ID != *student.ID && *sessionUser.Role != data.RoleAdministrator {
		ok, err := app.models.Users.IsUserTeacherOrParentOfStudent(*student.ID, *sessionUser.ID)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
		if !ok {
			app.notAllowed(w, r)
			return
		}
	}

	years, err := app.models.Years.GetYearsForStudent(*student.ID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"years": years})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) newYear(w http.ResponseWriter, r *http.Request) {
	sessionUser := app.getUserFromContext(r)

	var input struct {
		DisplayName string `json:"display_name"`
		Courses     int    `json:"courses"`
		NewClasses  []struct {
			Name        string `json:"name"`
			DisplayName string `json:"display_name"`
		} `json:"new_classes"`
		OldClasses []struct {
			ClassID     int    `json:"class_id"`
			DisplayName string `json:"display_name"`
		} `json:"old_classes"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.NewValidator()

	v.Check(input.DisplayName != "", "display_name", "cannot be empty")
	v.Check(input.Courses > 0, "courses", "must be valid")

	var classIDs []int

	for _, oc := range input.OldClasses {
		classIDs = append(classIDs, oc.ClassID)
		v.Check(oc.DisplayName != "", "display_name", fmt.Sprintf("class id %d name cannot be empty", oc.ClassID))
	}

	for _, nc := range input.NewClasses {
		v.Check(nc.DisplayName != "", "display_name", "new class display name cannot be empty")
		v.Check(nc.Name != "", "name", "new class name cannot be empty")
	}

	allClassIDs, err := app.models.Classes.GetAllClassIDs()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}
	badIDs := helpers.VerifyExistsInSlice(classIDs, allClassIDs)
	v.Check(badIDs != nil, "class_id", fmt.Sprintf("invalid class id(s): %v", badIDs))

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	year := &data.Year{
		DisplayName: input.DisplayName,
		Courses:     &input.Courses,
	}

	newYearID, err := app.models.Years.InsertYear(year)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	var classYears []*data.ClassYear

	for _, nc := range input.NewClasses {
		class := &data.Class{
			Name:    &nc.Name,
			Teacher: &data.User{ID: sessionUser.ID},
		}

		insertClassID, err := app.models.Classes.InsertClass(class)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}

		classYears = append(classYears, &data.ClassYear{YearID: *newYearID, ClassID: *insertClassID, DisplayName: nc.DisplayName})
	}

	for _, oc := range input.OldClasses {
		classYears = append(classYears, &data.ClassYear{
			DisplayName: oc.DisplayName,
			ClassID:     oc.ClassID,
			YearID:      *newYearID,
		})
	}

	for _, cy := range classYears {
		err := app.models.Years.InsertYearForClass(cy)
		if err != nil {
			app.writeInternalServerError(w, r, err)
			return
		}
	}

	err = app.models.Years.RemoveCurrentYear()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.models.Years.SetYearAsCurrent(*newYearID)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusCreated, envelope{"message": "success"})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}
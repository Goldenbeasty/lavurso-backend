package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/annusingmar/lavurso-backend/internal/data"
	"github.com/annusingmar/lavurso-backend/internal/validator"
	"github.com/julienschmidt/httprouter"
)

func (app *application) listAllGrades(w http.ResponseWriter, r *http.Request) {
	grades, err := app.models.Grades.AllGrades()
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"grades": grades})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) getGrade(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	gradeID, err := strconv.Atoi(params.ByName("id"))
	if gradeID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchGrade.Error())
		return
	}

	grade, err := app.models.Grades.GetGradeByID(gradeID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchGrade):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"grade": grade})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) createGrade(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Identifier string `json:"identifier"`
		Value      int    `json:"value"`
	}

	err := app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	grade := &data.Grade{
		Identifier: input.Identifier,
		Value:      input.Value,
	}

	v := validator.NewValidator()

	v.Check(grade.Identifier != "", "identifier", "must be provided")
	v.Check(grade.Value > 0, "value", "must be provided and valid")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	err = app.models.Grades.InsertGrade(grade)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusCreated, envelope{"grade": grade})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

func (app *application) updateGrade(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	gradeID, err := strconv.Atoi(params.ByName("id"))
	if gradeID < 0 || err != nil {
		app.writeErrorResponse(w, r, http.StatusNotFound, data.ErrNoSuchGrade.Error())
		return
	}

	grade, err := app.models.Grades.GetGradeByID(gradeID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoSuchGrade):
			app.writeErrorResponse(w, r, http.StatusNotFound, err.Error())
		default:
			app.writeInternalServerError(w, r, err)
		}
		return
	}

	var input struct {
		Identifier *string `json:"identifier"`
		Value      *int    `json:"value"`
	}

	err = app.inputJSON(w, r, &input)
	if err != nil {
		app.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if input.Identifier != nil {
		grade.Identifier = *input.Identifier
	}
	if input.Value != nil {
		grade.Value = *input.Value
	}

	v := validator.NewValidator()

	v.Check(grade.Identifier != "", "identifier", "must be provided")
	v.Check(grade.Value > 0, "value", "must be provided and valid")

	if !v.Valid() {
		app.writeErrorResponse(w, r, http.StatusBadRequest, v.Errors)
		return
	}

	err = app.models.Grades.UpdateGrade(grade)
	if err != nil {
		app.writeInternalServerError(w, r, err)
		return
	}

	err = app.outputJSON(w, http.StatusOK, envelope{"grade": grade})
	if err != nil {
		app.writeInternalServerError(w, r, err)
	}
}

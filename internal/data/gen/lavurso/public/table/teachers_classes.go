//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/postgres"
)

var TeachersClasses = newTeachersClassesTable("public", "teachers_classes", "")

type teachersClassesTable struct {
	postgres.Table

	//Columns
	TeacherID postgres.ColumnInteger
	ClassID   postgres.ColumnInteger

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type TeachersClassesTable struct {
	teachersClassesTable

	EXCLUDED teachersClassesTable
}

// AS creates new TeachersClassesTable with assigned alias
func (a TeachersClassesTable) AS(alias string) *TeachersClassesTable {
	return newTeachersClassesTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new TeachersClassesTable with assigned schema name
func (a TeachersClassesTable) FromSchema(schemaName string) *TeachersClassesTable {
	return newTeachersClassesTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new TeachersClassesTable with assigned table prefix
func (a TeachersClassesTable) WithPrefix(prefix string) *TeachersClassesTable {
	return newTeachersClassesTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new TeachersClassesTable with assigned table suffix
func (a TeachersClassesTable) WithSuffix(suffix string) *TeachersClassesTable {
	return newTeachersClassesTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newTeachersClassesTable(schemaName, tableName, alias string) *TeachersClassesTable {
	return &TeachersClassesTable{
		teachersClassesTable: newTeachersClassesTableImpl(schemaName, tableName, alias),
		EXCLUDED:             newTeachersClassesTableImpl("", "excluded", ""),
	}
}

func newTeachersClassesTableImpl(schemaName, tableName, alias string) teachersClassesTable {
	var (
		TeacherIDColumn = postgres.IntegerColumn("teacher_id")
		ClassIDColumn   = postgres.IntegerColumn("class_id")
		allColumns      = postgres.ColumnList{TeacherIDColumn, ClassIDColumn}
		mutableColumns  = postgres.ColumnList{}
	)

	return teachersClassesTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		TeacherID: TeacherIDColumn,
		ClassID:   ClassIDColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
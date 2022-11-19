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

var ThreadsRecipients = newThreadsRecipientsTable("public", "threads_recipients", "")

type threadsRecipientsTable struct {
	postgres.Table

	//Columns
	ThreadID postgres.ColumnInteger
	UserID   postgres.ColumnInteger
	GroupID  postgres.ColumnInteger

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type ThreadsRecipientsTable struct {
	threadsRecipientsTable

	EXCLUDED threadsRecipientsTable
}

// AS creates new ThreadsRecipientsTable with assigned alias
func (a ThreadsRecipientsTable) AS(alias string) *ThreadsRecipientsTable {
	return newThreadsRecipientsTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new ThreadsRecipientsTable with assigned schema name
func (a ThreadsRecipientsTable) FromSchema(schemaName string) *ThreadsRecipientsTable {
	return newThreadsRecipientsTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new ThreadsRecipientsTable with assigned table prefix
func (a ThreadsRecipientsTable) WithPrefix(prefix string) *ThreadsRecipientsTable {
	return newThreadsRecipientsTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new ThreadsRecipientsTable with assigned table suffix
func (a ThreadsRecipientsTable) WithSuffix(suffix string) *ThreadsRecipientsTable {
	return newThreadsRecipientsTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newThreadsRecipientsTable(schemaName, tableName, alias string) *ThreadsRecipientsTable {
	return &ThreadsRecipientsTable{
		threadsRecipientsTable: newThreadsRecipientsTableImpl(schemaName, tableName, alias),
		EXCLUDED:               newThreadsRecipientsTableImpl("", "excluded", ""),
	}
}

func newThreadsRecipientsTableImpl(schemaName, tableName, alias string) threadsRecipientsTable {
	var (
		ThreadIDColumn = postgres.IntegerColumn("thread_id")
		UserIDColumn   = postgres.IntegerColumn("user_id")
		GroupIDColumn  = postgres.IntegerColumn("group_id")
		allColumns     = postgres.ColumnList{ThreadIDColumn, UserIDColumn, GroupIDColumn}
		mutableColumns = postgres.ColumnList{ThreadIDColumn, UserIDColumn, GroupIDColumn}
	)

	return threadsRecipientsTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ThreadID: ThreadIDColumn,
		UserID:   UserIDColumn,
		GroupID:  GroupIDColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}

package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/lovego/bsql"
	"github.com/lovego/kala/job"
)

const TABLE_NAME = "jobs"

var (
	tableSql = bsql.Table{
		Name:   TABLE_NAME,
		Desc:   "任务管理",
		Struct: &job.Job{},
		Constraints: []string{
			"UNIQUE(id)", "UNIQUE(owner, name)",
		},
	}.Sql()
	insertFields     = bsql.FieldsFromStruct(job.Job{}, nil)
	insertColumns    = bsql.Fields2ColumnsStr(insertFields)
	conflictFields   = bsql.FieldsFromStruct(job.Job{}, []string{"id", "name", "Owner", "JobType"})
	conflictColumns  = bsql.Fields2ColumnsStr(conflictFields)
	conflictExcluded = bsql.FieldsToColumnsStr(conflictFields, "excluded.", nil)
)

type DB struct {
	conn *sql.DB
}

// New instantiates a new DB.
func New(dsn string) *DB {
	connection, err := sql.Open("postgres", dsn)
	if err != nil {
		job.Logger.Fatal(err)
	}
	// passive attempt to create table
	connection.Exec(tableSql)
	return &DB{
		conn: connection,
	}
}

// GetAll returns all persisted Jobs.
func (d DB) GetAll() ([]*job.Job, error) {
	query := fmt.Sprintf(`select coalesce(json_agg(row_to_json(j)), '[]'::json) from %[1]s AS j`, TABLE_NAME)

	var r sql.NullString
	err := d.conn.QueryRow(query).Scan(&r)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	err = nil
	jobs := []*job.Job{}
	if r.Valid {
		err = json.Unmarshal([]byte(r.String), &jobs)
	}

	jobsInitiated := []*job.Job{}
	for _, j := range jobs {
		if err = j.InitDelayDuration(false); err != nil {
			return nil, err
		}
		jobsInitiated = append(jobsInitiated, j)
	}

	return jobsInitiated, err
}

// Get returns a persisted Job.
func (d DB) Get(id string) (*job.Job, error) {
	template := `select to_jsonb(j) from (select row_to_json(j) from %[1]s AS j where id = $1) as j;`
	query := fmt.Sprintf(template, TABLE_NAME)
	var r sql.NullString
	err := d.conn.QueryRow(query, id).Scan(&r)
	if err != nil {
		return nil, err
	}
	result := &job.Job{}
	if r.Valid {
		err = json.Unmarshal([]byte(r.String), &result)
	}
	return result, err
}

// Delete deletes a persisted Job.
func (d DB) Delete(id string) error {
	query := fmt.Sprintf(`delete from %v where id = $1;`, TABLE_NAME)
	_, err := d.conn.Exec(query, id)
	return err
}

// Save persists a Job.
func (d DB) Save(j *job.Job) error {
	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES %s ON CONFLICT (id) DO UPDATE SET (%s) = (%s)`,
		TABLE_NAME, insertColumns,
		bsql.StructValues(j, insertFields),
		conflictColumns, conflictExcluded,
	)
	_, err := d.conn.Exec(query)
	return err
}

// Close closes the connection to Postgres.
func (d DB) Close() error {
	return d.conn.Close()
}

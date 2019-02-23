/*
package justSQL is just a helper function. You should build statements and create functions to use them.
*/

package justSQL

import (
	"database/sql"
	"errors"
	"fmt"
	log "github.com/autopogo/justLogging"

	pq "github.com/lib/pq"
)

// DBInst is a configuration structure for SQL. The auth will be zero'd once its opened.
type DBConfig struct {
	user     string
	password string
	name     string
	host		string
	DB       *sql.DB
}

// The only unique error for this package so far
var (
	ErrStmtConflict = errors.New("justSQL: Tried to create two statements of the same name")
)

// ErrorConv converts a standard error message to its SQL State code, so we can interpret the error if we need to. Eg, "unique conflict" would be 23505.
func ErrorConv(passedError error) string {
	return string(passedError.(*pq.Error).Code)
}



// Open opens the database connection, and makes the maps of precompiled statements
func (d *DBConfig) Open(user, pass, name string) error {
	var err error
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s host=%s",
		user,
		password,
		name,
		host)
	if d.DB, err = sql.Open("postgres", dbinfo); err != nil {
		log.Errorf("justSql, .Open(): Error opening database: %v", err)
		return err
	} else if _, err = d.DB.Exec("SELECT 1"); err != nil {
		log.Errorf("justSql: .Open(): Error opening database after test: %v", err)
		return err
	}

	return nil
}

// Close closes the database
func (d *DBConfig) Close() error {
	return d.DB.Close()
}

// stmt writes to a map while also compiling a sql statement.
func (d *DBConfig) Stmt(queryString string) (*sql.Stmt, error) {
	queryStatement, err := d.DB.Prepare(queryString);
	if err != nil {
		log.Errorf("justSql, .PushStmt(): Some kind of database error: %v", err)
		return nil, err
	}
	return queryStatement, err
}


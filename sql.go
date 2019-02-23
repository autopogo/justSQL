/*
package justSQL presents a configuration structure you can use to open and close a database connection. It also allows you to register common SQL queries into a map (by name), and other modules can then look up SQL queries they need and shared them. You do not lookup queries through the map on every call, only when your module is initialized, as they return a pointer to the actual value. This adds some pressure to the garbage collector, and I wish there was a way to take over manual control of memory. It's helper, doesn't fully wrap
*/
package justSQL

import (
	"database/sql"
	"errors"
	"fmt"
	log "github.com/autopogo/justLogging"
	pq "github.com/lib/pq"
)

// TODO jSQL multirow process

// DBInst is a configuration structure for SQL. The auth will be zero'd once its opened.
type DBConfig struct {
	User     string
	Password string
	Name     string
	DB       *sql.DB
}

// The starting size of the statements map
const (
	numStatements int = 50 // numStatements is the default map size
)

// The only unique error for this package so far
var (
	ErrStmtConflict = errors.New("justSQL: Tried to create two statements of the same name")
)

// ErrorConv converts a standard error message to its SQL State code, so we can interpret the error if we need to. Eg, "unique conflict" would be 23505.
func ErrorConv(passedError error) string {
	return string(passedError.(*pq.Error).Code)
}

// Open opens the database connection, and makes the maps of precompiled statements
func (d *DBConfig) Open() error {
	/* I used to zero auth data but now I'm worried about reconnecting
		defer func(){
		 d.User = ""
		 d.Password = ""
		 d.Name = ""
	 }() */

	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		d.User,
		d.Password,
		d.Name)
	var err error
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

// pushStmt writes to a map while also compiling a sql statement.
func (d *DBConfig) PushStmt(nameString string, queryString string) (*sql.Stmt, error) {
	if queryStatement, err := d.DB.Prepare(queryString); err != nil {
		log.Errorf("justSql, .PushStmt(): Some kind of database error: %v", err)
		return nil, err
	}
}


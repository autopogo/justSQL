package justSQL
/* 
package justSQL presents a configuration structure you can use to open and close a database connection. It also allows you to register common SQL queries into a map (by name), and other modules can then look up SQL queries they need and shared them. You do not lookup queries through the map on every call, only when your module is initialized, as they return a pointer to the actual value. This adds some pressure to the garbage collector, and I wish there was a way to take over manual control of memory.
*/

import (
  "database/sql"
  pq "github.com/lib/pq"
  "fmt"
  "errors"
  log "github.com/autopogo/justLogging"
)

// TODO jSQL multirow process

// DBInst is a configuration structure for SQL. The auth will be zero'd once its opened.
type DBConfig struct {
 User       string
 Password   string
 Name       string
 DB         *sql.DB
 stmtsMap   map[string]*sql.Stmt
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
func ErrorConv(passedError error) (string) {
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
    log.Errorf("justSql, .Open(): Error opening database: %v", err);
    return err
  } else if _, err = d.DB.Exec("SELECT 1"); err != nil {
		log.Errorf("justSql: .Open(): Error opening database after test: %v", err);
    return err
  }

  d.stmtsMap = make(map[string]*sql.Stmt, numStatements)
  return nil
}

// Close closes the database
func (d *DBConfig) Close() error {
  return d.DB.Close()
}

// pushStmt writes to a map while also compiling a sql statement.
func (d *DBConfig) PushStmt(nameString string, queryString string) (*sql.Stmt, error) {
	if gotStmt, ok := d.stmtsMap[nameString]; ok {
    log.Enterf("justSql, .PushStmt(): Tried to add statement that already exists.")
    return gotStmt, ErrStmtConflict
  }
	if queryStatement, err := d.DB.Prepare(queryString); err != nil {
		log.Errorf("justSql, .PushStmt(): Some kind of database error: %v", err)
    return nil, err
  } else {
		d.stmtsMap[nameString] = queryStatement
    return d.stmtsMap[nameString], nil
	}
  // We're pressuring go's gc because it has to do escape analysis on this. Unsafe pointers would solve this maybe? In C, we do manual memory management, so we could just say "Hey let this live until the program dies, don't worry about it."
}

// getStmt is just a shim for the statement map
func (d *DBConfig) GetStmt(nameString string) (*sql.Stmt) {
	return d.stmtsMap[nameString]
}

/*

There is already a golang sql utility wrapper for structure/map queries. 

// these should be structures, and there should be receivers for each type of query. i guess they take the db as an argument
func (d *DBinst) Insert(table string, cols []string, values []string) // should these be struct pointers with all these values
func (d *DBinst) Update(table string, cols []string,  values []string, where_col []string, where_val []string)
func (d *DBinst) Delete(table string, where_col []string, where_val []string)
func (d *DBinst) Select(table string, cols []string, where_col []string, where_val []string)
func (d *DBinst) InnerJoin(ltable string,
	rtable string,
	cols []string,
	lcols []string,
	rcols []string,
	where_col []string,
	where_val []string)
func (d *DBinst) LeftJoin(ltable string,
	rtable string,
	cols []string,
	lcols []string,
	rcols []string,
	where_col []string,
	where_val []string)
func (d *DBinst) RightJoin(ltable string,
	rtable string,
	cols []string,
	lcols []string,
	rcols []string,
	where_col []string,
	where_val []string)
func (d *DBinst) FullJoin(ltable string,
	rtable string,
	cols []string,
	lcols []string,
	rcols []string,
	where_col []string,
	where_val []string)
//func (d *DBinst) query( // an interator: next, and scan
//func (d *DBinst) exec(

// write a receiver for init, make it like open
// write a receive for close- close all the statements
*/

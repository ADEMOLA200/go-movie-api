package utils

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"

	_ "github.com/lib/pq"
)

func ConnectToSQLDB(host, user, password, dbname string, port int) *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	// defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	log.Println("db successfully connected!")
	return db
}

func WriteError(statusCode int, message string, err interface{}) []byte {
	response := APIResponse{
		Status:  statusCode,
		Message: message,
		Data:    err,
	}
	data, err := json.Marshal(response)
	if err == nil {
		return data
	} else {
		log.Printf("Err: %s", err)
	}
	return nil
}

func WriteInfo(format string, args ...interface{}) []byte {
	response := map[string]string{
		"info": fmt.Sprintf(format, args...),
	}
	if data, err := json.Marshal(response); err == nil {
		return data
	} else {
		log.Printf("Err: %s", err)
	}
	return nil
}

// check if a table exist in the pg db
func CheckTableExist(ctx context.Context, db *sql.DB, tableName string) (bool, error) {
	query := `
		SELECT EXISTS (
   SELECT FROM pg_tables
   WHERE  schemaname = 'public'
   AND    tablename  = $1
   );
	`
	row := db.QueryRowContext(ctx, query, tableName)
	var response bool
	_ = row.Scan(&response)
	return response, nil
}

// 500 - internal server error
func Dispatch500Error(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(WriteError(http.StatusInternalServerError, "", fmt.Sprintf("%v", err)))
}

// 501 - not implemented
func Dispatch501Error(w http.ResponseWriter, msg string, err error) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write(WriteError(http.StatusNotImplemented, msg, err))
}

// 405 - method not allowed
func Dispatch405Error(w http.ResponseWriter, msg string, err error) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write(WriteError(http.StatusMethodNotAllowed, msg, err))
}

// 400 - bad request
func Dispatch400Error(w http.ResponseWriter, msg string, err any) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write(WriteError(http.StatusBadRequest, msg, err))
}

// 404 - not found
func Dispatch404Error(w http.ResponseWriter, msg string, err any) {
	w.WriteHeader(http.StatusNotFound)
	w.Write(WriteError(http.StatusNotFound, msg, err))
}

func CmToFeetInches(cm float64) string {
	feet := int(cm / 30.48)
	inches := (cm / 30.48) - float64(feet)
	inches = math.Round(inches*12*100) / 100 // round to 2 decimal places

	return fmt.Sprintf("%dft %.2fin", feet, inches)
}

package utils

import (
	"database/sql"
	"fmt"
)

//utils for temporary usage, to be deleted during development

// PrintQueryResults prints all rows and columns from the query results in raw text form.
func PrintQueryResults(rows *sql.Rows) error {
	// Get the column names
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %v", err)
	}

	// Create a slice to hold the values for each row
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Iterate over each row
	for rows.Next() {
		// Scan the row values into the value pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}

		// Print each column and its value
		for i, col := range columns {
			val := values[i]
			fmt.Printf("%s: %v\n", col, val)
		}
		fmt.Println("----") // Separator between rows
	}

	// Check for any errors encountered during iteration
	if err := rows.Err(); err != nil {
		return fmt.Errorf("row iteration error: %v", err)
	}

	return nil
}

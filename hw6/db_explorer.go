package main 

import ( 
	"database/sql" 
	"encoding/json" 
	"fmt" 
	"net/http" 
	"strconv" 
	"strings"
) 

type Column struct { 
	Name string 
	Type string // "int" | "string" 
	Nullable bool 
	IsPrimary bool 
} 

type Table struct { 
	Name string 
	Columns []Column 
	ColumnMap map[string]Column 
	PrimaryKey Column 
} 

type DbExplorer struct { 
	db *sql.DB 
	tables map[string]*Table 
} 

func NewDbExplorer(db *sql.DB) (http.Handler, error) { 
	explorer := &DbExplorer{ 
		db: db, 
		tables: map[string]*Table{}, 
	} 
	
	if err := explorer.loadMetadata(); err != nil { 
		return nil, err 
	} 
	return explorer, nil 
} 

func (e *DbExplorer) loadMetadata() error { 
	rows, err := e.db.Query("SHOW TABLES") 
	if err != nil { 
		return err 
	} 

	var tableNames []string 
	for rows.Next() { 
		var name string 
		if err := rows.Scan(&name); err != nil { 
			rows.Close() 
			return err 
		} 
		
		tableNames = append(tableNames, name) 
	} 
	rows.Close() 
	
	for _, tableName := range tableNames { 
		table := &Table{ 
			Name: tableName, 
			ColumnMap: map[string]Column{}, 
		} 
		
		colRows, err := e.db.Query("SHOW FULL COLUMNS FROM `" + tableName + "`") 
		if err != nil { 
			return err 
		} 
		
		for colRows.Next() { 
			var field, typ, collation, null, key, def, extra, privileges, comment sql.NullString 
			if err := colRows.Scan(&field, &typ, &collation, &null, &key, &def, &extra, &privileges, &comment); err != nil { 
				colRows.Close() 
				return err 
			} 
			
			colType := "string" 
			if strings.HasPrefix(typ.String, "int") { 
				colType = "int" 
			} 
			
			col := Column{ 
				Name: field.String, 
				Type: colType, 
				Nullable: null.String == "YES", 
				IsPrimary: key.String == "PRI", 
			} 
			
			if col.IsPrimary { 
				table.PrimaryKey = col 
			} 
			
			table.Columns = append(table.Columns, col) 
			table.ColumnMap[col.Name] = col 
		} 
		colRows.Close() 
		
		e.tables[tableName] = table 
	} 
	return nil 
} 

func (e *DbExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) { 
	path := strings.Trim(r.URL.Path, "/") 
	if path == "" { 
		e.handleTables(w) 
		return 
	} 
	
	parts := strings.Split(path, "/") 
	tableName := parts[0] 
	
	table, ok := e.tables[tableName] 
	if !ok { 
		writeError(w, http.StatusNotFound, "unknown table") 
		return 
	} 
	
	if len(parts) == 1 || parts[1] == "" { 
		switch r.Method { 
		case http.MethodGet: 
			e.handleSelectAll(w, r, table)
		case http.MethodPut: 
			e.handleInsert(w, r, table) 
		default: 
			writeError(w, http.StatusNotFound, "unknown method") 
		} 
		return 
	} 
	
	id, err := strconv.Atoi(parts[1]) 
	if err != nil { 
		writeError(w, http.StatusNotFound, "record not found") 
		return 
	} 
	
	switch r.Method { 
	case http.MethodGet: 
		e.handleSelectOne(w, table, id) 
	case http.MethodPost: 
		e.handleUpdate(w, r, table, id) 
	case http.MethodDelete: 
		e.handleDelete(w, table, id) 
	default: 
		writeError(w, http.StatusNotFound, "unknown method") 
	} 
} 

func (e *DbExplorer) handleTables(w http.ResponseWriter) { 
	names := make([]string, 0, len(e.tables)) 
	for name := range e.tables { 
		names = append(names, name) 
	} 
	
	writeJSON(w, http.StatusOK, map[string]interface{}{ 
		"response": map[string]interface{}{ 
			"tables": names, 
		}, 
	}) 
} 

func (e *DbExplorer) handleSelectAll(w http.ResponseWriter, r *http.Request, t *Table) { 
	limit := 5 
	offset := 0 
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil { 
		limit = l 
	} 
	if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil { 
		offset = o 
	} 
	
	query := fmt.Sprintf("SELECT * FROM `%s` LIMIT ? OFFSET ?", t.Name) 
	rows, err := e.db.Query(query, limit, offset) 
	
	if err != nil { 
		writeError(w, 500, err.Error()) 
		return 
	} 
	defer rows.Close() 
	
	var records []map[string]interface{} 
	
	for rows.Next() { 
		record, err := scanRow(rows, t) 
		if err != nil { 
			writeError(w, 500, err.Error()) 
			return 
		} 
		records = append(records, record) 
	} 
	
	writeJSON(w, 200, map[string]interface{}{ 
		"response": map[string]interface{}{ 
			"records": records, 
		}, 
	}) 
} 

func (e *DbExplorer) handleSelectOne(w http.ResponseWriter, t *Table, id int) { 
	query := fmt.Sprintf("SELECT * FROM `%s` WHERE `%s` = ?", t.Name, t.PrimaryKey.Name) 
	rowrow := e.db.QueryRow(query, id) 
	
	record, err := scanSingleRow(row, t) 
	if err == sql.ErrNoRows { 
		writeError(w, 404, "record not found") 
		return 
	} 
	if err != nil { 
		writeError(w, 500, err.Error()) 
		return 
	} 
	
	writeJSON(w, 200, map[string]interface{}{ 
		"response": map[string]interface{}{ 
			"record": record, 
		}, 
	}) 
} 

func (e *DbExplorer) handleInsert(w http.ResponseWriter, r *http.Request, t *Table) { 
	var input map[string]interface{} 
	json.NewDecoder(r.Body).Decode(&input) 
	
	var fields []string 
	var placeholders []string 
	var values []interface{} 
	
	for _, col := range t.Columns { 
		if col.IsPrimary { 
			continue 
		} 
		
		val, exists := input[col.Name] 
		if !exists { 
			if col.Nullable { 
				fields = append(fields, "`"+col.Name+"`") 
				values = append(values, nil) 
				placeholders = append(placeholders, "?") 
			} else { 
				fields = append(fields, "`"+col.Name+"`") 
				values = append(values, "") 
				placeholdersplaceholders = append(placeholders, "?") 
			} 
			continue 
		} 
		
		validVal, ok := validateType(col, val) 
		if !ok { 
			writeError(w, 400, fmt.Sprintf("field %s have invalid type", col.Name)) 
			return 
		} 
		
		fields = append(fields, "`"+col.Name+"`") 
		values = append(values, validVal) 
		placeholders = append(placeholders, "?") 
	} 
	
	query := fmt.Sprintf( 
		"INSERT INTO `%s` (%s) VALUES (%s)", 
		t.Name, 
		strings.Join(fields, ","), 
		strings.Join(placeholders, ","), 
	) 
	
	res, err := e.db.Exec(query, values...) 
	if err != nil { 
		writeError(w, 500, err.Error()) 
		return 
	} 
	
	id, _ := res.LastInsertId() 
	
	writeJSON(w, 200, map[string]interface{}{ 
		"response": map[string]interface{}{ 
			t.PrimaryKey.Name: id, 
		}, 
	}) 
} 

func (e *DbExplorer) handleUpdate(w http.ResponseWriter, r *http.Request, t *Table, id int) { 
	var input map[string]interface{} 
	json.NewDecoder(r.Body).Decode(&input) 

	var setParts []string 
	var values []interface{} 
	
	for name, val := range input { 
		col, ok := t.ColumnMap[name] 
		if !ok || col.IsPrimary { 
			writeError(w, 400, fmt.Sprintf("field %s have invalid type", name)) 
			return 
		} 
		
		validVal, ok := validateType(col, val) 
		if !ok { 
			writeError(w, 400, fmt.Sprintf("field %s have invalid type", name)) 
			return 
		} 
		
		setParts = append(setParts, "`"+name+"` = ?") 
		values = append(values, validVal) 
	} 
	
	if len(setParts) == 0 { 
		writeJSON(w, 200, map[string]interface{}{ 
			"response": map[string]interface{}{ 
				"updated": 0, 
			}, 
		}) 
		return 
	} 
	
	values = append(values, id) 
	query := fmt.Sprintf( 
		"UPDATE `%s` SET %s WHERE `%s` = ?", 
		t.Name, 
		strings.Join(setParts, ","), 
		t.PrimaryKey.Name, ) 
		
	res, err := e.db.Exec(query, values...) 
	if err != nil { 
		writeError(w, 500, err.Error()) 
		return 
	} 
	
	affected, _ := res.RowsAffected() 
	writeJSON(w, 200, map[string]interface{}{ 
		"response": map[string]interface{}{ 
			"updated": affected, 
		}, 
	}) 
} 

func (e *DbExplorer) handleDelete(w http.ResponseWriter, t *Table, id int) { 
	query := fmt.Sprintf(
		"DELETE FROM `%s` WHERE `%s` = ?", 
		t.Name, 
		t.PrimaryKey.Name
	) 
	res, err := e.db.Exec(query, id) 
	
	if err != nil { 
		writeError(w, 500, err.Error()) 
		return 
	} 
	
	affected, _ := res.RowsAffected() 
	writeJSON(w, 200, map[string]interface{}{ 
		"response": map[string]interface{}{ 
			"deleted": affected, 
		}, 
	}) 
} 

func validateType(col Column, val interface{}) (interface{}, bool) { 
	if val == nil { 
		if col.Nullable { 
			return nil, true 
		} 
		return nil, false 
	} 
	
	switch col.Type { 
	case "int": 
		v, ok := val.(float64) 
		if !ok { 
			return nil, false 
		} 
		return int(v), true 
	case "string": 
		v, ok := val.(string) 
		if !ok { 
			return nil, false 
		} 
		return v, true 
	} 
	return nil, false 
	
} 

func scanRow(rows *sql.Rows, t *Table) (map[string]interface{}, error) { 
	values := make([]interface{}, len(t.Columns)) 
	ptrs := make([]interface{}, len(t.Columns)) 
	
	for i := range values { 
		ptrs[i] = &values[i] 
	} 
	
	if err := rows.Scan(ptrs...); err != nil { 
		return nil, err 
	} 
	
	result := map[string]interface{}{} 
	
	for i, col := range t.Columns { 
		val := values[i] 
		if val == nil { 
			result[col.Name] = nil 
			continue 
		} 
		if col.Type == "int" { 
			result[col.Name] = val.(int64) 
		} else { 
			result[col.Name] = string(val.([]byte)) 
		} 
	} 
	return result, nil 
} 

func scanSingleRow(row *sql.Row, t *Table) (map[string]interface{}, error) { 
	values := make([]interface{}, len(t.Columns)) 
	ptrs := make([]interface{}, len(t.Columns)) 
	
	for i := range values { 
		ptrs[i] = &values[i] 
	} 
	
	if err := row.Scan(ptrs...); err != nil { 
		return nil, err 
	} 
	
	result := map[string]interface{}{} 
	
	for i, col := range t.Columns { 
		val := values[i] 
		if val == nil { 
			result[col.Name] = nil 
			continue 
		} 
		
		if col.Type == "int" { 
			result[col.Name] = val.(int64) 
		} else { 
			result[col.Name] = string(val.([]byte)) 
		} 
	} 
	
	return result, nil 
} 
			
func writeError(w http.ResponseWriter, status int, msg string) { 
	writeJSON(w, status, map[string]interface{}{ 
		"error": msg, 
	}) 
} 


func writeJSON(w http.ResponseWriter, status int, data interface{}) { 
	w.Header().Set("Content-Type", "application/json") 
	w.WriteHeader(status) 
	json.NewEncoder(w).Encode(data) 
}
package main

type schema struct {
	Name      string `xorm:"'SCHEMA_NAME'"`
	CharSet   string `xorm:"'DEFAULT_CHARACTER_SET_NAME'"`
	Collation string `xorm:"'DEFAULT_COLLATION_NAME'"`
}

func (schema) TableName() string {
	return "SCHEMATA"
}

type table struct {
	Schema    string `xorm:"'TABLE_SCHEMA'"`
	Name      string `xorm:"'TABLE_NAME'"`
	Collation string `xorm:"'TABLE_COLLATION'"`
	Comment   string `xorm:"'TABLE_COMMENT'"`
}

func (table) TableName() string {
	return "TABLES"
}

type column struct {
	Schema    string `xorm:"'TABLE_SCHEMA'"`
	Table     string `xorm:"'TABLE_NAME'"`
	Name      string `xorm:"'COLUMN_NAME'"`
	Default   string `xorm:"'COLUMN_DEFAULT'"`
	Nullable  string `xorm:"'IS_NULLABLE'"`
	DataType  string `xorm:"'COLUMN_TYPE'"`
	Key       string `xorm:"'COLUMN_KEY'"`
	CharSet   string `xorm:"'CHARACTER_SET_NAME'"`
	Collation string `xorm:"'COLLATION_NAME'"`
	Comment   string `xorm:"'COLUMN_COMMENT'"`
}

func (column) TableName() string {
	return "COLUMNS"
}

// data instance

// DB 数据库实例元信息
type DB struct {
	Name      string            `json:"name,omitempty"`
	CharSet   string            `json:"charset,omitempty"`
	Collation string            `json:"collation,omitempty"`
	Tables    []Table           `json:"tables,omitempty"`
	Extra     map[string]string `json:"extra,omitempty"`
}

// Table 表元信息
type Table struct {
	DB        string            `json:"-"`
	Name      string            `json:"name,omitempty"`
	Collation string            `json:"collation,omitempty"`
	Comment   string            `json:"comment,omitempty"`
	Columns   []Column          `json:"columns,omitempty"`
	Extra     map[string]string `json:"extra,omitempty"`
}

// Column 列元信息
type Column struct {
	DB      string `json:"-"`
	Table   string `json:"-"`
	Name    string `json:"name,omitempty"`
	Default string `json:"default,omitempty"`
	// Nullable  string            `json:"nullable,omitempty"`
	DataType string `json:"data_type,omitempty"`
	// Key       string            `json:"key,omitempty"`
	// CharSet   string            `json:"charset,omitempty"`
	// Collation string            `json:"collation,omitempty"`
	Comment string            `json:"comment,omitempty"`
	Extra   map[string]string `json:"extra,omitempty"`
}

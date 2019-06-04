package do

import (
	"fmt"

	"gitee.com/flwwsg/utils-go/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

// Config 配置信息
type Config struct {
	Host          string
	Port          int
	Socket        string
	Username      string
	Password      string
	DBName        string
	Output        string
	Debug         bool
	IgnorePattern []string // 需要忽略的表名
	Module        map[string][]string
}

type Repo struct {
	engine *xorm.Engine
}

func NewRepo(c *Config) Repo {
	sqlURL := GenDataSource(c, "information_schema", "charset=utf8&parseTime=true&loc=Local")
	engine, err := xorm.NewEngine("mysql", sqlURL)
	errors.PanicOnErr(err)
	engine.ShowSQL(c.Debug)
	return Repo{engine: engine}
}

// 生成 mysql 连接
func GenDataSource(c *Config, dbName string, params string) string {
	if c.Socket == "" {
		// use tcp
		if c.Password == "" {
			return fmt.Sprintf("%s@tcp(%s:%d)/%s?%s", c.Username, c.Host, c.Port, dbName, params)
		}
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", c.Username, c.Password, c.Host, c.Port, dbName, params)
	}
	if c.Password == "" {
		// use unix
		return fmt.Sprintf("%s@unix(%s)/%s?%s", c.Username, c.Socket, dbName, params)
	}
	return fmt.Sprintf("%s:%s@unix(%s)/%s?%s", c.Username, c.Password, c.Socket, dbName, params)
}

func (repo *Repo) GetDB(cond *DB) (item DB, err error) {
	var sCond schema
	if cond == nil {
		panic("db name required")
	}
	sCond.Name = cond.Name
	sCond.CharSet = cond.CharSet
	sCond.Collation = cond.Collation
	schemas, err := repo.getSchemas(&sCond)
	if err != nil || len(schemas) != 1 {
		// 数据库重名或者不存在
		return DB{}, err
	}
	targetSchema := schemas[0]
	var tables []Table
	tables, err = repo.GetTables(&Table{DB: targetSchema.Name})
	if err != nil {
		return DB{}, err
	}
	return DB{Name: targetSchema.Name, CharSet: targetSchema.CharSet, Collation: targetSchema.Collation, Tables: tables}, nil
}

// 获取数据库下的所有表
func (repo *Repo) GetTables(cond *Table) (items []Table, err error) {
	var tCond table
	if cond != nil {
		tCond.Name = cond.Name
		tCond.Schema = cond.DB
		tCond.Collation = cond.Collation
		tCond.Comment = cond.Comment
	}
	tables, err := repo.getTables(&tCond)
	if err != nil {
		return nil, err
	}
	for i := range tables {
		cols, err := repo.GetColumns(&Column{
			DB:    tables[i].Schema,
			Table: tables[i].Name,
		})
		if err != nil {
			return nil, err
		}
		items = append(items, Table{
			DB:        tables[i].Schema,
			Name:      tables[i].Name,
			Collation: tables[i].Collation,
			Comment:   tables[i].Comment,
			Columns:   cols,
		})
	}
	return items, nil
}

// 获取所有列信息
func (repo *Repo) GetColumns(cond *Column) (items []Column, err error) {
	var cCond column
	if cond != nil {
		cCond.Schema = cond.DB
		cCond.Table = cond.Table
		cCond.Default = cond.Default
		cCond.DataType = cond.DataType
		cCond.Comment = cond.Comment
	}
	cols, err := repo.getColumns(&cCond)
	if err != nil {
		return nil, err
	}
	for i := range cols {
		col := cols[i]
		items = append(items, Column{
			DB:       col.Schema,
			Table:    col.Table,
			Name:     col.Name,
			Default:  col.Default,
			DataType: col.DataType,
			Comment:  col.Comment,
		})
	}
	return items, nil
}

func (repo *Repo) getColumns(cond *column) (items []column, err error) {
	if err = repo.engine.Find(&items, cond); err != nil {
		return nil, err
	}
	return items, nil
}

// 获取数据库信息
func (repo *Repo) getSchemas(cond *schema) (items []schema, err error) {
	err = repo.engine.Find(&items, cond)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (repo *Repo) getTables(cond *table) (items []table, err error) {
	if err = repo.engine.Find(&items, cond); err != nil {
		return nil, err
	}
	return items, nil
}

package struct2sql

import (
	"sync"

	"github.com/Masterminds/squirrel"
)

// 传入结构体,进行去重判断 如果第一次出现 插入一条create语句
// 随后在insert语句 根据tag传入数据,这里需要判断关系 如果这个结构里面有其他的struct 转为 one to one
// 如果这个结构里面有 list struct 转化为 one to many (暂时) 如果list 的struct 里面也有当前的 []struct 定义为many to many
// one to many 插入一条当前数据 将列表数据展开
// many to many 情况下 创建一张中间表,将两个表的id插入到中间表中(创建第三张表也放到set里 create语句)

type model string

func (m model) String() string {
	return string(m)
}

var (
	_default   model = "default"
	_one2one   model = "one2one"
	_one2many  model = "one2many"
	_many2many model = "many2many"
)

type pair struct {
	model    model
	key      string // 用来通过by name寻找
	exported bool   // 是否是大写暴漏的字段
	alias    string // 用来导出的别名
	tags     map[string]string
	extra    string // 根据model除了_default以外会用到的额外字段,主要标记键
}

type meta struct {
	// model name (one2one; one2many; many2many; default)
	field []pair
}

type Struct2Sql struct {
	tag               string // read tag name: default is `s2sql`
	extraExport       bool   // enhanced mode: forced reading of lowercase unexposed fields
	splitIdent        string // split Ident: default is `;`
	placeholderFormat squirrel.PlaceholderFormat

	lock sync.Mutex
	set  map[string]*meta // map[struct name] filed name

	createSql   []string
	insertSql   []string
	relationSql []string // many to many
}

func NewStruct2Sql() *Struct2Sql {
	return &Struct2Sql{
		tag:               "s2sql",
		extraExport:       false,
		splitIdent:        ";",
		placeholderFormat: squirrel.Question,
		set:               make(map[string]*meta),
	}
}

// TableName Map to table name
type TableName interface {
	TableName() string
}

const (
	RelationField = "relation_field"
	Alias         = "alias"
)

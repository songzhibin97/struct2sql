package struct2sql

import (
	"errors"
	"reflect"
	"strings"

	"github.com/songzhibin97/gkit/tools/deepcopy"

	"github.com/Masterminds/squirrel"

	"github.com/songzhibin97/go-baseutils/base/bternaryexpr"
	"github.com/songzhibin97/go-ognl"
)

var (
	errMustStruct = errors.New("struct2sql: must struct")
)

var tableNameType = reflect.TypeOf((*TableName)(nil)).Elem()

type InstallPair struct {
	Sql string        `json:"sql"`
	Arg []interface{} `json:"arg"`
}

type insertBuilder struct {
	squirrel.InsertBuilder
	mp map[string]interface{}
}

func (s *Struct2Sql) BuildInstall(val interface{}) ([]InstallPair, error) {
	meta, err := s.scan(val)
	if err != nil {
		return nil, err
	}
	build := squirrel.Insert(s.getTableName(reflect.TypeOf(val)))
	setMap := make(map[string]interface{})
	queue := make([]pair, 0)
	for _, field := range meta.field {
		switch field.model {
		case _default:
			setMap[field.alias] = ognl.Get(val, field.key).Value()
		case _one2one:
			setMap[field.alias] = ognl.Get(val, field.key+"."+field.extra).Value()
		case _one2many, _many2many:
			// 需要展开
			// 笛卡尔积?
			queue = append(queue, field)
		}
	}
	builds := []insertBuilder{{
		InsertBuilder: build.SetMap(setMap).PlaceholderFormat(s.placeholderFormat),
		mp:            setMap,
	}}
	var many2manyBuilds []squirrel.InsertBuilder
	for _, p := range queue {
		ln := len(builds)

		for i := 0; i < ln; i++ {
			extraKey := p.extra
			if p.model == _many2many {
				sp := strings.SplitN(p.extra, ";", 3)
				id := ognl.Get(val, sp[1]).Value()
				for _, extra := range ognl.Get(val, p.key+".#"+sp[0]).Values() {
					many2manyBuilds = append(many2manyBuilds, squirrel.Insert(sp[2]).SetMap(map[string]interface{}{sp[0]: extra, sp[1]: id}))
				}
				extraKey = sp[0]
			}

			for _, extra := range ognl.Get(val, p.key+".#"+extraKey).Values() {
				cpSetMap := make(map[string]interface{})
				err = deepcopy.DeepCopy(&cpSetMap, &builds[i].mp)
				if err != nil {
					return nil, err
				}
				cpSetMap[p.alias] = extra
				builds = append(builds, insertBuilder{
					InsertBuilder: build.SetMap(setMap).PlaceholderFormat(s.placeholderFormat),
					mp:            cpSetMap,
				})
			}
		}
		builds = builds[ln:]
	}
	res := make([]InstallPair, 0, len(builds)+len(many2manyBuilds))
	for _, builder := range builds {
		sql, arg, err := builder.ToSql()
		if err != nil {
			return nil, err
		}
		res = append(res, InstallPair{
			Sql: sql,
			Arg: arg,
		})
	}
	for _, manyBuild := range many2manyBuilds {
		sql, arg, err := manyBuild.ToSql()
		if err != nil {
			return nil, err
		}
		res = append(res, InstallPair{
			Sql: sql,
			Arg: arg,
		})
	}

	return res, nil
}

// scan 扫描struct
func (s *Struct2Sql) scan(val interface{}) (*meta, error) {
	t, v := reflect.TypeOf(val), reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		return s.scan(v.Elem().Interface())
	}
	if v.Kind() != reflect.Struct {
		return nil, errMustStruct
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	singleName := t.PkgPath() + "." + t.Name()
	if s.set[singleName] != nil {
		return s.set[singleName], nil
	}
	meta := &meta{
		field: nil,
	}
	s.set[singleName] = meta
	for i := 0; i < v.NumField(); i++ {
		vField := v.Field(i)
		tField := t.Field(i)
		name := tField.Name
		p := pair{model: _default, key: name, tags: s.parseTag(tField)}
		if !vField.IsValid() {
			continue
		}
		if !s.extraExport && !vField.CanInterface() {
			continue
		}
		p.exported = vField.CanInterface()
		p.alias = bternaryexpr.TernaryExpr(p.tags[Alias] != "", p.tags[Alias], p.key)

		// 判断字段类型
		nvField := vField.Type()
		if nvField.Kind() == reflect.Ptr {
			nvField = nvField.Elem()
		}
		switch nvField.Kind() {
		case reflect.Struct:
			// 如果对应的是struct 升级为 one2one模式
			p.model = _one2one
			p.extra = bternaryexpr.TernaryExpr(p.tags[RelationField] == "", nvField.Name()+"ID", p.tags[RelationField])

		case reflect.Slice, reflect.Array:
			toField := nvField.Elem()
			if toField.Kind() == reflect.Ptr {
				toField = toField.Elem()
			}
			if toField.Kind() != reflect.Struct {
				continue
			}
			// 如果是对应slice/array struct 升级为 one2many模式
			p.model = _one2many
			p.extra = bternaryexpr.TernaryExpr(p.tags[RelationField] == "", toField.Name()+"ID", p.tags[RelationField])
		loop:
			for j := 0; j < toField.NumField(); j++ {
				tovField := toField.Field(j)
				nTovField := tovField.Type
				if nTovField.Kind() == reflect.Ptr {
					nTovField = nTovField.Elem()
				}
				switch nTovField.Kind() {
				case reflect.Slice, reflect.Array:
					nTovField := nTovField.Elem()
					if nTovField.Kind() == reflect.Ptr {
						nTovField = nTovField.Elem()
					}
					if nTovField.Kind() != reflect.Struct || nTovField != v.Type() {
						continue
					}
					p.model = _many2many
					tags := s.parseTag(toField.Field(j))
					// p.extra = now struct filed name ; opposite end filed name ; tripartite table name
					p.extra += ";" + bternaryexpr.TernaryExpr(tags[RelationField] == "", nTovField.Name()+"ID", tags[RelationField])
					p.extra += ";"
					nTovFieldName, toFieldName := s.getTableName(nTovField), s.getTableName(toField)
					if nTovFieldName > toFieldName {
						p.extra += nTovFieldName + "_" + toFieldName
					} else {
						p.extra += toFieldName + "_" + nTovFieldName
					}
					break loop
				}
			}
		}
		meta.field = append(meta.field, p)
	}
	return meta, nil
}

func (s *Struct2Sql) getTableName(p reflect.Type) string {
	if !p.Implements(tableNameType) {
		return p.Name()
	}
	method, ok := p.MethodByName("TableName")
	if !ok {
		return p.Name()
	}
	value := method.Func.Call([]reflect.Value{reflect.New(p).Elem()})
	if len(value) != 1 {
		return p.Name()
	}
	return value[0].String()
}

func (s *Struct2Sql) parseTag(field reflect.StructField) map[string]string {
	tag := field.Tag.Get(s.tag)
	res := make(map[string]string)
	for _, ident := range strings.Split(tag, s.splitIdent) {
		ident = strings.TrimSpace(ident)
		if ident == "" {
			continue
		}
		list := strings.SplitN(ident, ":", 2)
		res[list[0]] = bternaryexpr.TernaryExpr(len(list) == 2, list[1], "")
	}
	return res
}

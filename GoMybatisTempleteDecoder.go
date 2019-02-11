package GoMybatis

import (
	"bytes"
	"github.com/zhuxiujia/GoMybatis/utils"
	"strings"
)

var equalOperator = []string{"/", "+", "-", "*", "**", "|", "^", "&", "%", "<", ">", ">=", "<=", " in ", " not in ", " or ", "||", " and ", "&&", "==", "!="}

type GoMybatisTempleteDecoder struct {
	tree map[string]*MapperXml
}

func (it *GoMybatisTempleteDecoder) DecodeTree(tree map[string]*MapperXml) error {
	it.tree = tree
	for _, v := range tree {
		it.Decode(v)
	}
	return nil
}

func (it *GoMybatisTempleteDecoder) Decode(mapper *MapperXml) error {

	if mapper.Tag == "selectTemplete" {
		var tables = mapper.Propertys["tables"]
		var columns = mapper.Propertys["columns"]
		var wheres = mapper.Propertys["wheres"]
		var sql bytes.Buffer
		sql.WriteString("select ")
		if columns == "" {
			columns = "*"
		}
		sql.WriteString(columns)
		sql.WriteString(" from ")
		sql.WriteString(tables)
		if len(wheres) > 0 {
			sql.WriteString(" where ")
			mapper.ElementItems = append(mapper.ElementItems, ElementItem{
				ElementType: Element_String,
				DataString:  sql.String(),
			})
			//TODO decode wheres
			sql.Reset()
			it.DecodeWheres(wheres, mapper)
		}

		if mapper.Id == "" {
			mapper.Id = mapper.Tag
		}
		mapper.Tag = Element_Select
	} else if mapper.Tag == "insertTemplete" {
		var tables = mapper.Propertys["tables"]
		var resultMap = mapper.Propertys["resultMap"]
		var inserts = mapper.Propertys["inserts"]

		if resultMap == "" {
			resultMap = "BaseResultMap"
		}
		if inserts == "" {
			inserts = "*?*"
		}

		var resultMapData = it.tree[resultMap]
		if resultMapData == nil {
			panic(utils.NewError("GoMybatisTempleteDecoder", "resultMap not define! id = ", resultMap))
		}
		var sql bytes.Buffer
		sql.WriteString("insert into ")
		sql.WriteString(tables)

		mapper.ElementItems = append(mapper.ElementItems, ElementItem{
			ElementType: Element_String,
			DataString:  sql.String(),
		})

		//add insert column
		var trimColumn = ElementItem{
			ElementType:  Element_Trim,
			Propertys:    map[string]string{"prefix": "(", "suffix": ")", "suffixOverrides": ","},
			ElementItems: []ElementItem{},
		}
		if inserts == "*?*" {
			for _, v := range resultMapData.ElementItems {
				trimColumn.ElementItems = append(trimColumn.ElementItems, ElementItem{
					ElementType: Element_If,
					Propertys:   map[string]string{"test": it.convertEqualAction(v.Propertys["property"])},
					ElementItems: []ElementItem{
						{
							ElementType: Element_String,
							DataString:  v.Propertys["column"] + ",",
						},
					},
				})
			}
		} else if inserts == "*" {
			for _, v := range resultMapData.ElementItems {
				trimColumn.ElementItems = append(trimColumn.ElementItems, ElementItem{
					ElementType: Element_String,
					DataString:  v.Propertys["column"] + ",",
				})
			}
		} else {
			panic(utils.NewError("GoMybatisTempleteDecoder", `inserts only support "*" or "*?*"`))
		}

		mapper.ElementItems = append(mapper.ElementItems, trimColumn)

		//add insert arg
		var trimArg = ElementItem{
			ElementType:  Element_Trim,
			Propertys:    map[string]string{"prefix": "values (", "suffix": ")", "suffixOverrides": ","},
			ElementItems: []ElementItem{},
		}
		if inserts == "*?*" {
			for _, v := range resultMapData.ElementItems {
				trimArg.ElementItems = append(trimArg.ElementItems, ElementItem{
					ElementType: Element_If,
					Propertys:   map[string]string{"test": it.convertEqualAction(v.Propertys["property"])},
					DataString:  "#{" + v.Propertys["property"] + "},",
				})
			}
		} else if inserts == "*" {
			for _, v := range resultMapData.ElementItems {
				trimArg.ElementItems = append(trimArg.ElementItems, ElementItem{
					ElementType: Element_String,
					DataString:  "#{" + v.Propertys["property"] + "},",
				})
			}
		}
		mapper.ElementItems = append(mapper.ElementItems, trimArg)
	}

	return nil
}

//解码逗号分隔的where
func (it *GoMybatisTempleteDecoder) DecodeWheres(arg string, mapper *MapperXml) {
	var wheres = strings.Split(arg, ",")
	for index, v := range wheres {
		var expressions = strings.Split(v, "?")
		if len(expressions) > 1 {
			//TODO have ?
			var newWheres bytes.Buffer
			for k, v := range expressions {
				if k > 0 {
					if index > 0 {
						newWheres.WriteString(" and ")
					}
					newWheres.WriteString(v)
				}
			}
			var item = ElementItem{
				ElementType: Element_If,
				Propertys:   map[string]string{"test": it.convertEqualAction(expressions[0])},
				DataString:  newWheres.String(),
			}
			mapper.ElementItems = append(mapper.ElementItems, item)
		}
	}
}

func (it *GoMybatisTempleteDecoder) convertEqualAction(arg string) string {
	for _, v := range equalOperator {
		if strings.Contains(arg, v) {
			return arg
		}
	}
	return arg + ` != null`
}

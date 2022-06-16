package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/songcocl/fast-go-lib/model"
	"reflect"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
)

type Backend struct {
	*BackendOption
}
type BackendOption struct {
	NoNeedLogin            []string               `json:"noNeedLogin"`                           //无需登录的方法,同时也就不需要鉴权了
	NoNeedRight            []string               `json:"noNeedRight"`                           //无需鉴权的方法,但需要登录
	Layout                 string                 `json:"layout" default:"default"`              //布局模板
	Auth                   map[string]interface{} `json:"auth"`                                  //权限控制类
	Model                  string                 `json:"model"`                                 //模型对象
	SearchFields           string                 `json:"searchFields" default:"id"`             //快速搜索时执行查找的字段
	RelationSearch         bool                   `json:"relationSearch"`                        //是否是关联查询
	DataLimit              bool                   `json:"dataLimit"`                             //是否开启数据限制
	DataLimitField         string                 `json:"dataLimitField" default:"admin_id"`     //数据限制字段
	DataLimitFieldAutoFill string                 `json:"dataLimitFieldAutoFill" default:"true"` //数据限制开启时自动填充限制字段值
	ModelValidate          bool                   `json:"modelValidate"`                         //是否开启Validate验证
	ModelSceneValidate     bool                   `json:"modelSceneValidate"`                    //是否开启模型场景验证
	MultiFields            string                 `json:"multiFields" default:"status"`          //Multi方法可批量修改的字段
	SelectpageFields       string                 `json:"selectpageFields" default:"*"`          //Selectpage可显示的字段
	ExcludeFields          []string               `json:"excludeFields"`                         //前台提交过来,需要排除的字段数据
	ImportHeadType         string                 `json:"importHeadType" default:"comment"`      //导入文件首行类型
}

func GetDfBackendOption() (o *BackendOption) {
	return &BackendOption{
		Layout:                 "default",
		SearchFields:           "id",
		DataLimitField:         "admin_id",
		DataLimitFieldAutoFill: "true",
		MultiFields:            "status",
		SelectpageFields:       "*",
		ImportHeadType:         "comment",
	}
}
func NewBackend(o *BackendOption) *Backend {
	dfOpt := GetDfBackendOption()
	if o.Layout == "" {
		o.Layout = dfOpt.Layout
	}
	if o.SearchFields == "" {
		o.SearchFields = dfOpt.SearchFields
	}
	if o.DataLimitField == "" {
		o.DataLimitField = dfOpt.DataLimitField
	}
	if o.DataLimitFieldAutoFill == "" {
		o.DataLimitFieldAutoFill = dfOpt.DataLimitFieldAutoFill
	}
	if o.MultiFields == "" {
		o.MultiFields = dfOpt.MultiFields
	}
	if o.SelectpageFields == "" {
		o.SelectpageFields = dfOpt.SelectpageFields
	}
	if o.ImportHeadType == "" {
		o.ImportHeadType = dfOpt.ImportHeadType
	}
	b := &Backend{
		BackendOption: o,
	}

	return b
}
func NewBackendByMap(opt map[string]interface{}) *Backend {
	o := GetDfBackendOption()

	var (
		field reflect.StructField
		ok    bool
	)

	for k, v := range opt {
		if field, ok = (reflect.ValueOf(o)).Elem().Type().FieldByName(k); !ok {
			continue
		}
		if field.Type == reflect.TypeOf(v) {
			reflect.ValueOf(o).Elem().FieldByName(k).Set(reflect.ValueOf(v))
		}
	}
	b := &Backend{
		BackendOption: o,
	}
	return b
}

/**
 * 生成查询所需要的条件,排序方式
 * searchfields   快速查询的字段
 * relationSearch 是否关联查询
 */
func (b *Backend) BuildParams(req *model.ApiPageReq, searchfields string, relationSearch bool) (map[string]interface{}, string, int, int) {
	if searchfields == "" {
		searchfields = b.SearchFields
	}
	search := req.Search
	sort := req.Sort
	order := req.Order
	offset := req.Offset
	limit := req.Limit
	var filter map[string]interface{}
	json.Unmarshal([]byte(req.Filter), &filter)
	var op map[string]interface{}
	json.Unmarshal([]byte(req.Op), &op)
	var where = make(map[string]interface{})
	var tableName string
	if relationSearch {
		if b.Model != "" {
			tableName = b.Model + "."
		}
	}

	sortArr := strings.Split(sort, ",")
	for key, item := range sortArr {
		if strings.Index(item, ".") == -1 {
			item = tableName + strings.Trim(item, " ")
		}
		sortArr[key] = item + " " + order
	}
	sort = strings.Join(sortArr, ",")

	adminIds := b.GetDataLimitAdminIds()
	if adminIds != nil {
		where[tableName+b.DataLimitField+" in (?)"] = adminIds
	}
	if search != "" {
		fmt.Println(searchfields)
		searcharr := strings.Split(searchfields, ",")
		for _, val := range searcharr {
			if strings.Index(val, ".") == -1 {
				val = tableName + val
			}
		}
		where[strings.Join(searcharr, "|")+" like ?"] = "%" + search + "%"
	}

	for key, val := range filter {
		var sym string = "="
		if op[key] != nil {
			sym = op[key].(string)
		}
		key = gstr.CaseSnakeFirstUpper(key)
		sym = strings.ToLower(sym)
		switch sym {
		case "=", "<>":
			where[key+" "+sym+" ?"] = val.(string)
			break
		case "like", "not like", "like %...%", "not like %...%":
			where[key+" "+strings.ReplaceAll(sym, "%...%", "")+" ?"] = "%" + val.(string) + "%"
			break
		case ">", ">=", "<", "<=":
			where[key+" "+sym+" "] = val.(int)
			break
		case "findin", "findinset", "find_in_set":
			where["find_in_set"] = "('" + val.(string) + "', "
			if relationSearch {
				where["find_in_set"] = where["find_in_set"].(string) + key
			} else {
				where["find_in_set"] = where["find_in_set"].(string) + "`"
			}
			where[where["find_in_set"].(string)+strings.ReplaceAll(key, ".", "`.`")+")"] = nil
			break
		case "in", "in(...)", "not in", "not in(...)":
			if op, ok := val.([]interface{}); ok {
				where[key+" "+sym+"(?)"] = op
			} else {
				where[key+" "+sym+"(?)"] = strings.Split(val.(string), ",")
			}
			break
		case "between", "not between":
			arr := strings.Split(val.(string), ",")
			arr = arr[:2]
			if arr[0] == "" {
				if sym == "between" {
					sym = "<="
				} else {
					sym = ">"
				}
				where[key+" "+sym+" "] = arr[1]
			} else if arr[1] == "" {
				if sym == "between" {
					sym = ">="
				} else {
					sym = "<"
				}
				where[key+" "+sym+" "] = arr[0]
			} else {
				where[key+" "+sym+" ? and ?"] = arr
			}
			break
		case "range", "not range":
			val = strings.ReplaceAll(val.(string), " - ", ",")
			arr := strings.Split(val.(string), ",")
			arr = arr[:2]
			if arr[0] == "" {
				if sym == "range" {
					sym = "<="
				} else {
					sym = ">"
				}
				where[key+" "+sym+" "] = arr[1]
			} else if arr[1] == "" {
				if sym == "range" {
					sym = ">="
				} else {
					sym = "<"
				}
				where[key+" "+sym+" "] = arr[0]
			} else {
				where[key+" "+strings.ReplaceAll(sym, "range", "between")+" ? and ?"] = arr
			}
			break
		case "null", "is null", "not null", "is not null":
			where[key+" "+strings.ReplaceAll(sym, "is", "")] = nil
			break
		default:
		}
	}

	return where, sort, offset, limit
}

/**
 * Selectpage的实现方法
* 当前方法只是一个比较通用的搜索匹配,请按需重载此方法来编写自己的搜索逻辑,$where按自己的需求写即可
* 这里示例了所有的参数，所以比较复杂，实现上自己实现只需简单的几行即可
*/
func (b *Backend) Selectpage(r *ghttp.Request) {
	/*query := r.GetFormMap()
	//搜索关键词,客户端输入以空格分开,这里接收为数组
	word := query["q_word"].(map[string]interface{})
	//当前页
	page := r.GetQuery("PageNumber")
	//分页大小
	pagesize := r.GetQuery("PageSize").Int()
	//搜索条件
	andor := query["andOr"]
	//排序方式
	orderby := query["orderBy"].(map[string]interface{})
	//显示的字段
	field := r.GetQuery("ShowField", "name")
	//主键
	primarykey := r.GetQuery("KeyField").String()
	//主键值
	primaryvalue := r.GetQuery("KeyValue").String()
	//搜索字段
	searchfield := query["searchField"]
	//自定义搜索条件
	custom = query["custom"]
	//是否返回树形结构
	istree := r.GetQuery("IsTree", 0).Bool()
	ishtml := r.GetQuery("IsHtml", 0).Bool()
	if istree {
		word = nil
		pagesize = 99999
	}
	var order map[string]interface{}
	for _, val := range orderby {
		//order[val[0]] = val[1]
	}

	var where map[string]interface{}
	//如果有primaryvalue,说明当前是初始化传值
	if primaryvalue != "" {
		where[primarykey] = primaryvalue
		pagesize = 99999
	}

	adminIds := b.GetDataLimitAdminIds()
	if adminIds != nil {
		where[b.DataLimitField+" in (?)"] = adminIds
	}

	var list []map[string]interface{}
	total, _ := g.Model(b.Model).Where(where).Count()

	datalist, _ := g.Model(b.Model).Fields(b.SelectpageFields).FieldsEx("password,salt").Where(where).Order(order).All()

	for index, val := range datalist {

	}

	return map[string]interface{}{
		"total": total,
		"rows":  list,
	}*/
}

/**
 * 获取数据限制的管理员ID
 * 禁用数据限制时返回的是null
 */
func (b *Backend) GetDataLimitAdminIds() []string {
	if !b.DataLimit {
		return nil
	}

	return nil
}

/**
 * 排除前台提交过来的字段
 */
func (b *Backend) PreExcludeFields(params map[string]interface{}) map[string]interface{} {
	for _, val := range b.ExcludeFields {
		if params[val] != nil {
			delete(params, val)
		}
	}
	return params
}

/**
 * 查看
 */
func (b *Backend) Index(r *model.ApiPageReq, pointer interface{}) (total int, err error) {
	where, sort, offset, limit := b.BuildParams(r, "", false)

	total, _ = g.Model(b.Model).Where(where).Count()

	err = g.Model(b.Model).Where(where).Order(sort).Limit(offset, limit).Scan(pointer)

	return
}

/**
* 添加
 */
func (b *Backend) Add(ctx context.Context) (int64, error) {
	r := g.RequestFromCtx(ctx)
	query := r.GetFormMap()
	params := query["row"].(map[string]interface{})
	if len(params) > 0 {
		params = b.PreExcludeFields(params)

		if b.DataLimit && b.DataLimitFieldAutoFill == "true" {
			params[b.DataLimitField] = b.Auth["id"]

		}

		tx, err := g.DB().Begin(ctx)
		if err != nil {
			return 0, err
		}
		// 方法退出时检验返回值，
		// 如果结果成功则执行tx.Commit()提交,
		// 否则执行tx.Rollback()回滚操作。
		defer func() {
			if err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}()

		result, inErr := g.Model(b.Model).Data(params).Insert()

		if inErr != nil {
			return 0, inErr
		}

		id, idErr := result.LastInsertId()

		if idErr != nil {
			return 0, idErr
		}

		return id, nil
	}
	return 0, nil
}

/**
* 编辑
 */
func (b *Backend) Edit(ctx context.Context, id uint64) (int64, error) {
	r := g.RequestFromCtx(ctx)
	query := r.GetFormMap()
	params := query["row"].(map[string]interface{})
	if len(params) > 0 {
		params = b.PreExcludeFields(params)

		tx, err := g.DB().Begin(ctx)
		if err != nil {
			return 0, err
		}
		// 方法退出时检验返回值，
		// 如果结果成功则执行tx.Commit()提交,
		// 否则执行tx.Rollback()回滚操作。
		defer func() {
			if err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}()

		result, upErr := g.Model(b.Model).Data(params).Where("id", id).Update()

		if upErr != nil {
			return 0, upErr
		}

		row, idErr := result.RowsAffected()

		if idErr != nil {
			return 0, idErr
		}

		return row, nil
	}
	return 0, nil
}

/**
 * 删除
 */
func (b *Backend) Del(pk string, ids []interface{}) (int64, error) {
	var where = make(map[string]interface{})
	adminIds := b.GetDataLimitAdminIds()
	if adminIds != nil {
		where[b.DataLimitField+" in (?)"] = adminIds
	}

	where[pk+" in (?)"] = ids
	result, err := g.Model(b.Model).Where(where).Delete()

	if err != nil {
		return 0, err
	}

	row, idErr := result.RowsAffected()

	if idErr != nil {
		return 0, idErr
	}

	return row, nil
}

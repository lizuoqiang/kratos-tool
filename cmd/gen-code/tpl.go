// Package main
// @Author: lzq
// @Email: lizuoqiang@huanjutang.com
// @Date: 2024-01-20 13:03:00
// @Description:
package main

import "strings"

const BizTpl = `package biz

import (
	"context"
	"errors"
	"time"
)

type {{model_name}} struct {
{{biz_field}}
}

type {{model_name}}Search struct {
    Page int
    Size int
    Orders []map[string]string

    {{model_name}}
}

type {{model_name}}DaoInterface interface {
	Create(ctx context.Context, data *{{model_name}}) (*{{model_name}}, error)
	DeleteById(ctx context.Context, id int) error
	GetById(ctx context.Context, id int, fields []string, with []string) (*{{model_name}}, error)
	GetByIds(ctx context.Context, ids []int, fields []string, with []string) (map[int]*{{model_name}}, error)
	GetByKeyAndValue(ctx context.Context, key string, value interface{}) (*{{model_name}}, error)
	UpdateById(ctx context.Context, id int, data map[string]interface{}) error
	List(ctx context.Context, data *{{model_name}}Search) ([]*{{model_name}}, int, error)
}

type {{model_name}}Business struct {
	logger     *log.Helper
	conf       *conf.Bootstrap
	{{model_name}}Dao {{model_name}}DaoInterface
}

func New{{model_name}}Business(logger log.Logger, conf *conf.Bootstrap, {{model_name}}Dao {{model_name}}DaoInterface) *{{model_name}}Business {
	return &{{model_name}}Business{
		logger:     log.NewHelper(logger),
		conf:       conf,
		{{model_name}}Dao: {{model_name}}Dao,
	}
}

func (c *{{model_name}}Business) {{model_name}}List(ctx context.Context, data *{{model_name}}Search) ([]*{{model_name}}, int, error)  {
	list, total, err := c.{{model_name}}Dao.List(ctx, data)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (c *{{model_name}}Business) {{model_name}}Create(ctx context.Context, data *{{model_name}}) (*{{model_name}}, error) {
	create, err := c.{{model_name}}Dao.Create(ctx, data)
	if err != nil {
		return nil, err
	}
	return create, nil
}

func (c *{{model_name}}Business) {{model_name}}Detail(ctx context.Context, id int) (*{{model_name}}, error) {
	detail, err := c.{{model_name}}Dao.GetById(ctx, id, []string{}, []string{})
	if err != nil {
		return nil, err
	}
	return detail, nil
}

func (c *{{model_name}}Business) {{model_name}}Delete(ctx context.Context, id int) error {
	err := c.{{model_name}}Dao.DeleteById(ctx, id)
	return err
}

func (c *{{model_name}}Business) {{model_name}}Update(ctx context.Context, id int, data *{{model_name}}) error {
	upData := make(map[string]interface{})
	// upData["name"] = data.Name
	err := c.{{model_name}}Dao.UpdateById(ctx, id, upData)
	return err
}`

const DaoTpl = `package data

import (
	"fmt"
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type {{model_name}}Dao struct {
	data *Data
	log  *log.Helper
}

func New{{model_name}}Dao(d *Data, logger log.Logger) biz.{{model_name}}DaoInterface {
	return &{{model_name}}Dao{
		data: d,
		log:  log.NewHelper(logger),
	}
}

func convert{{model_name}}(mod model.{{model_name}}) *biz.{{model_name}} {
	return &biz.{{model_name}}{
{{convert_fields}},
	}
}

func (dao *{{model_name}}Dao) Create(ctx context.Context, data *biz.{{model_name}}) (*biz.{{model_name}}, error) {
	mod := &model.{{model_name}}{}
	{{convert_exclude_id_fields}}

	err := dao.data.DB(ctx).Create(mod).Error
	if err != nil {
		return nil, err
	}

	data.Id = mod.Id
	return data, nil
}

func (dao *{{model_name}}Dao) DeleteById(ctx context.Context, id int) error {
	return dao.data.DB(ctx).Delete(&model.{{model_name}}{}, id).Error
}

func (dao *{{model_name}}Dao) GetById(ctx context.Context, id int, fields []string, with []string) (*biz.{{model_name}}, error) {
	mod := &model.{{model_name}}{}

	query := dao.data.DB(ctx).Model(mod)
	if len(with) > 0 {
	    for _, v := range with {
            query = query.Preload(v)
        }
	}
	if len(fields) > 0 {
		query = query.Select(fields)
	}

	err := query.First(mod, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return convert{{model_name}}(*mod), nil
}

func (dao *{{model_name}}Dao) GetByIds(ctx context.Context, ids []int, fields []string, with []string) (map[int]*biz.{{model_name}}, error) {
	var models []model.{{model_name}}
	query := dao.data.DB(ctx)

    if len(with) > 0 {
        for _, v := range with {
            query = query.Preload(v)
        }
    }
    if len(fields) > 0 {
        query = query.Select(fields)
    }

    err := query.Find(&models, ids).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	res := make(map[int]*biz.{{model_name}}, 0)
	for _, v := range models {
		res[v.Id] = convert{{model_name}}(v)
	}
	return res, nil
}

func (dao *{{model_name}}Dao) GetByKeyAndValue(ctx context.Context, key string, value interface{}) (*biz.{{model_name}}, error) {
	mod := &model.{{model_name}}{}
	err := dao.data.DB(ctx).Where(fmt.Sprintf("%s =?", key), value).First(mod).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return convert{{model_name}}(*mod), nil
}

func (dao *{{model_name}}Dao) UpdateById(ctx context.Context, id int, data map[string]interface{}) error {
	err := dao.data.DB(ctx).Model(&model.{{model_name}}{Id: id}).Updates(data).Error
	return err
}

func (dao *{{model_name}}Dao) List(ctx context.Context, data *biz.{{model_name}}Search) ([]*biz.{{model_name}}, int, error) {
	var models []model.{{model_name}}
	var total int64

	query := dao.data.DB(ctx).Model(&model.{{model_name}}{})
{{list_where}}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = query.Scopes(OrderScope(data.Orders)).Limit(data.Size).Offset((data.Page - 1) * data.Size).Find(&models).Error
	if err != nil {
		return nil, int(total), err
	}

	result := make([]*biz.{{model_name}}, 0)
	for _, item := range models {
		result = append(result, convert{{model_name}}(item))
	}
	return result, int(total), nil
}`

const ModelTpl = `package model

import (
	"gorm.io/plugin/soft_delete"
	"time"
)

type {{model_name}} struct {
{{gorm_field}}
}

func ({{model_name}}) TableName() string {
	return "{{table_name}}"
}`

const ProtobufTpl = `service Demo {
    rpc {{module_name}}Create ({{module_name}}CreateRequest) returns ({{module_name}}CreateReply) {
        option (google.api.http) = {
            post: "/{{route_prefix}}/create",
            body: "*"
        };
    }

    rpc {{module_name}}Delete ({{module_name}}DeleteRequest) returns ({{module_name}}DeleteReply) {
        option (google.api.http) = {
            post: "/{{route_prefix}}/delete",
            body: "*"
        };
    }

    rpc {{module_name}}Update ({{module_name}}UpdateRequest) returns ({{module_name}}UpdateReply) {
        option (google.api.http) = {
            post: "/{{route_prefix}}/update",
            body: "*"
        };
    }

    rpc {{module_name}}Detail ({{module_name}}DetailRequest) returns ({{module_name}}DetailReply) {
        option (google.api.http) = {
            get: "/{{route_prefix}}/detail",
        };
    }

    rpc {{module_name}}List ({{module_name}}ListRequest) returns ({{module_name}}ListReply) {
        option (google.api.http) = {
            get: "/{{route_prefix}}/list",
        };
    }
}

message {{module_name}}CreateRequest {
{{fields_exclude_id}}
}

message {{module_name}}CreateReply {
{{fields}}
}

message {{module_name}}DeleteRequest {
  int32 id = 1;
}

message {{module_name}}DeleteReply {}

message {{module_name}}UpdateRequest {
{{fields}}
}

message {{module_name}}UpdateReply {
{{fields}}
}

message {{module_name}}DetailRequest {
  int32 id = 1;
}

message {{module_name}}DetailReply {
{{fields}}
}

message {{module_name}}ListRequest {
  int32 page = 1 [(validate.rules).int32 = {gt:0}];
  int32 size = 2 [(validate.rules).int32 = {gt:0,lt:100}];
}

message {{module_name}}ListReply {
  int32 total = 1;

  message {{module_name}} {
{{fields}}
  }
  repeated {{module_name}} list = 2;
}`

const ServiceTpl = `package service

import (
	"context"
	"time"
)

func (s *DemoService) {{model_name}}Create(ctx context.Context, req *pb.{{model_name}}CreateRequest) (*pb.{{model_name}}CreateReply, error) {
	data := &biz.{{model_name}}{}
	_, err := s.{{lower_case_model_name}}Business.{{model_name}}Create(ctx, data)
	if err != nil {
		return nil, err
	}
	return &pb.{{model_name}}CreateReply{}, nil
}
func (s *DemoService) {{model_name}}Delete(ctx context.Context, req *pb.{{model_name}}DeleteRequest) (*pb.{{model_name}}DeleteReply, error) {
	err := s.{{lower_case_model_name}}Business.{{model_name}}Delete(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}
	return &pb.{{model_name}}DeleteReply{}, nil
}
func (s *DemoService) {{model_name}}Update(ctx context.Context, req *pb.{{model_name}}UpdateRequest) (*pb.{{model_name}}UpdateReply, error) {
	data := &biz.{{model_name}}{}
	err := s.{{lower_case_model_name}}Business.{{model_name}}Update(ctx, int(req.Id), data)
	if err != nil {
		return nil, err
	}
	return &pb.{{model_name}}UpdateReply{}, nil
}
func (s *DemoService) {{model_name}}Detail(ctx context.Context, req *pb.{{model_name}}DetailRequest) (*pb.{{model_name}}DetailReply, error) {
	detail, err := s.{{lower_case_model_name}}Business.{{model_name}}Detail(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}
	return &pb.{{model_name}}DetailReply{
		Id: int32(detail.Id),
	}, nil
}
func (s *DemoService) {{model_name}}List(ctx context.Context, req *pb.{{model_name}}ListRequest) (*pb.{{model_name}}ListReply, error) {
	data := &biz.{{model_name}}Search{
		Page: int(req.Page),
		Size: int(req.Size),
	}

	list, total, err := s.{{lower_case_model_name}}Business.{{model_name}}List(ctx, data)
	if err != nil {
		return nil, err
	}

	result := make([]*pb.{{model_name}}ListReply_{{model_name}}, len(list))
	for k, v := range list {
		result[k] = &pb.{{model_name}}ListReply_{{model_name}}{
			Id: int32(v.Id),
		}
	}

	return &pb.{{model_name}}ListReply{
		Total: int32(total),
		List:  result,
	}, nil
}

`

func genDao(tableName string, convertFields, convertExcludeIDFields, listWhere []string) string {
	replace := map[string]string{
		"{{model_name}}":                strings.Title(tableName),
		"{{convert_fields}}":            strings.Join(convertFields, ",\n"),
		"{{convert_exclude_id_fields}}": strings.Join(convertExcludeIDFields, "\n"),
		"{{list_where}}":                strings.Join(listWhere, "\n"),
	}

	return replaceStrings(DaoTpl, replace)
}

func genBiz(tableName string, bizStruct []string) string {
	replace := map[string]string{
		"{{model_name}}": strings.Title(tableName),
		"{{biz_field}}":  strings.Join(bizStruct, "\n"),
	}

	return replaceStrings(BizTpl, replace)
}

func genModel(tableName string, sqlStruct []string) string {
	replace := map[string]string{
		"{{model_name}}": strings.Title(tableName),
		"{{gorm_field}}": strings.Join(sqlStruct, "\n"),
	}

	return replaceStrings(ModelTpl, replace)
}

func genProtobuf(tableName string, protoFields, protoFieldsExcludeID []string) string {
	replace := map[string]string{
		"{{route_prefix}}":      strings.ReplaceAll(tableName, "_", "-"),
		"{{module_name}}":       strings.Title(tableName),
		"{{fields_exclude_id}}": strings.Join(protoFieldsExcludeID, "\n"),
		"{{fields}}":            strings.Join(protoFields, "\n"),
	}

	return replaceStrings(ProtobufTpl, replace)
}

func genService(tableName string) string {
	temp := strings.Title(tableName)
	replace := map[string]string{
		"{{model_name}}":            temp,
		"{{lower_case_model_name}}": strings.ToLower(temp[:1]) + temp[1:],
	}

	return replaceStrings(ServiceTpl, replace)
}

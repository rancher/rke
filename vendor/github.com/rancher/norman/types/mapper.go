package types

import (
	"github.com/pkg/errors"
	"github.com/rancher/norman/types/definition"
)

type Mapper interface {
	FromInternal(data map[string]interface{})
	ToInternal(data map[string]interface{})
	ModifySchema(schema *Schema, schemas *Schemas) error
}

type TypeMapper struct {
	Mappers         []Mapper
	typeName        string
	subSchemas      map[string]*Schema
	subArraySchemas map[string]*Schema
}

func (t *TypeMapper) FromInternal(data map[string]interface{}) {
	for fieldName, schema := range t.subSchemas {
		if schema.Mapper == nil {
			continue
		}
		fieldData, _ := data[fieldName].(map[string]interface{})
		schema.Mapper.FromInternal(fieldData)
	}

	for fieldName, schema := range t.subArraySchemas {
		if schema.Mapper == nil {
			continue
		}
		datas, _ := data[fieldName].([]interface{})
		for _, fieldData := range datas {
			mapFieldData, _ := fieldData.(map[string]interface{})
			schema.Mapper.FromInternal(mapFieldData)
		}
	}

	for _, mapper := range t.Mappers {
		mapper.FromInternal(data)
	}

	if data != nil {
		data["type"] = t.typeName
	}
}

func (t *TypeMapper) ToInternal(data map[string]interface{}) {
	for i := len(t.Mappers) - 1; i >= 0; i-- {
		t.Mappers[i].ToInternal(data)
	}

	for fieldName, schema := range t.subArraySchemas {
		if schema.Mapper == nil {
			continue
		}
		datas, _ := data[fieldName].([]map[string]interface{})
		for _, fieldData := range datas {
			schema.Mapper.ToInternal(fieldData)
		}
	}

	for fieldName, schema := range t.subSchemas {
		if schema.Mapper == nil {
			continue
		}
		fieldData, _ := data[fieldName].(map[string]interface{})
		schema.Mapper.ToInternal(fieldData)
	}
}

func (t *TypeMapper) ModifySchema(schema *Schema, schemas *Schemas) error {
	t.subSchemas = map[string]*Schema{}
	t.subArraySchemas = map[string]*Schema{}
	t.typeName = schema.ID

	mapperSchema := schema
	if schema.InternalSchema != nil {
		mapperSchema = schema.InternalSchema
	}
	for name, field := range mapperSchema.ResourceFields {
		fieldType := field.Type
		targetMap := t.subSchemas
		if definition.IsArrayType(fieldType) {
			fieldType = definition.SubType(fieldType)
			targetMap = t.subArraySchemas
		}

		schema := schemas.Schema(&schema.Version, fieldType)
		if schema != nil {
			targetMap[name] = schema
		}
	}

	for _, mapper := range t.Mappers {
		if err := mapper.ModifySchema(schema, schemas); err != nil {
			return errors.Wrapf(err, "mapping type %s", schema.ID)
		}
	}

	return nil
}

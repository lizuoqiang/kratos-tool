package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"regexp"
	"strings"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gen:proto",
		Short: "Generate Protobuf",
		Run:   generateProtobuf,
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func generateProtobuf(cmd *cobra.Command, args []string) {
	// 处理命令行参数（args）
	var sqlPath string
	modules := make([]string, 0)
	for k, arg := range args {
		if k == 0 {
			sqlPath = arg
		}
		if k == 1 {
			modules = append(modules, strings.Split(arg, ",")...)
		}
	}
	if len(modules) == 0 {
		modules = append(modules, []string{"dao", "model", "biz", "protobuf", "service"}...)
	}

	if ok, _ := FileExists(sqlPath); !ok {
		fmt.Println("sql文件不存在")
		return
	}

	result := parseSQL(GetFileContent(sqlPath))
	for tableName, items := range result {
		var (
			protoFields          []string
			protoFieldsExcludeId []string
			excludeIdIndex       int
		)

		for index, item := range items {
			fieldDef := fmt.Sprintf("   %s %s = %d;",
				sqlToProtoType(item["type"]),
				item["key"],
				index+1,
			)
			protoFields = append(protoFields, fieldDef)

			if item["key"] != "id" {
				excludeFieldDef := fmt.Sprintf("   %s %s = %d;",
					sqlToProtoType(item["type"]),
					item["key"],
					excludeIdIndex+1,
				)
				protoFieldsExcludeId = append(protoFieldsExcludeId, excludeFieldDef)
				excludeIdIndex++
			}
		}

		sqlStruct := []string{}
		bizStruct := []string{}
		convertFields := []string{}
		convertExcludeIdFields := []string{}
		listWhere := []string{}

		for _, field := range items {
			fieldName := toCamelCase(field["key"])
			goType := sqlToGoType(field["type"], fieldName)

			sqlField := fmt.Sprintf("%s %s", fieldName, goType)
			if fieldName == "Id" {
				sqlStruct = append(sqlStruct, sqlField+" `gorm:\"column:id;primary_key\"`")
			} else if fieldName == "IsDeleted" {
				comment := field["comment"]
				sqlStruct = append(sqlStruct, sqlField+fmt.Sprintf(" `gorm:\"softDelete:flag;column:%s;comment:'%s'\"`", field["key"], comment))
				sqlField = strings.Replace(sqlField, "soft_delete.DeletedAt", "int", 1)
			} else {
				sqlStruct = append(sqlStruct, sqlField+fmt.Sprintf(" `gorm:\"column:%s;comment:'%s'\"`", field["key"], field["comment"]))
			}
			bizStruct = append(bizStruct, sqlField)

			fieldsPrefix, fieldsSuffix := "", ""
			if fieldName == "IsDeleted" {
				fieldsPrefix = "int("
				fieldsSuffix = ")"
			}
			convertFields = append(convertFields, fmt.Sprintf("%s: %smod.%s%s", fieldName, fieldsPrefix, fieldName, fieldsSuffix))

			if !inSlice(fieldName, []string{"Id", "IsDeleted", "CreatedAt", "UpdatedAt"}) && inSlice(goType, []string{"int", "string", "time.Time"}) {
				convertExcludeIdFields = append(convertExcludeIdFields, fmt.Sprintf(zeroJudgment(goType), fieldName, fieldName, fieldName))
			}

			switch goType {
			case "int":
				listWhere = append(listWhere, fmt.Sprintf(`
				if data.%s != 0 {
					query = query.Where("%s = ?", data.%s)
				}`, fieldName, field["key"], fieldName))
			case "string":
				listWhere = append(listWhere, fmt.Sprintf(`
				if data.%s != "" {
					query = query.Where("%s = ?", data.%s)
				}`, fieldName, field["key"], fieldName))
			}
		}

		modelData := genModel(tableName, sqlStruct)
		bizData := genBiz(tableName, bizStruct)
		daoData := genDao(tableName, convertFields, convertExcludeIdFields, listWhere)
		protobufData := genProtobuf(tableName, protoFields, protoFieldsExcludeId)
		serviceData := genService(tableName)

		if inSlice("biz", modules) {
			GenFile(GetOutputPath("biz/"+tableName+".go"), bizData, 0755)
		}
		if inSlice("model", modules) {
			GenFile(GetOutputPath("model/"+tableName+".go"), modelData, 0755)
		}
		if inSlice("data", modules) {
			GenFile(GetOutputPath("data/"+tableName+".go"), daoData, 0755)
		}
		if inSlice("protobuf", modules) {
			GenFile(GetOutputPath("protobuf/"+tableName+".proto"), protobufData, 0755)
		}
		if inSlice("service", modules) {
			GenFile(GetOutputPath("service/"+tableName+".go"), serviceData, 0755)
		}
	}
}

func parseSQL(sql string) map[string][]map[string]string {
	sql = strings.NewReplacer("create table", "CREATE TABLE", "comment", "COMMENT").Replace(sql)
	tableArr := strings.FieldsFunc(sql, func(r rune) bool {
		return r == ';'
	})

	result := make(map[string][]map[string]string)
	for k, tableStr := range tableArr {
		fmt.Printf("table %d, %+v \n", k+1, tableStr)
		// Parse table name
		tableNamePattern := regexp.MustCompile("`([^`]+)`\\s*\\(")
		tableNameMatch := tableNamePattern.FindStringSubmatch(tableStr)

		// Parse field name
		table := tableNamePattern.ReplaceAllString(tableStr, "")
		fieldPattern := regexp.MustCompile("`(\\w+)`\\s+(\\w+).*COMMENT\\s*'(.+)'")
		fieldMatches := fieldPattern.FindAllStringSubmatch(tableStr, -1)

		if len(tableNameMatch) < 2 || len(fieldMatches) == 0 {
			fmt.Println("Failed to parse table structure, sql:", table)
			continue
		}

		// Reorganize structure
		fields := make([]map[string]string, len(fieldMatches))
		for i, fieldMatch := range fieldMatches {
			temp := make(map[string]string, 0)
			temp["key"] = fieldMatch[1]
			temp["type"] = fieldMatch[2]
			temp["comment"] = fieldMatch[3]
			fields[i] = temp
		}

		result[tableNameMatch[1]] = fields
	}

	return result
}

func replaceStrings(input string, replacements map[string]string) string {
	for key, value := range replacements {
		input = strings.ReplaceAll(input, key, value)
	}
	return input
}

func sqlToGoType(sqlType, fieldName string) string {
	if fieldName == "IsDeleted" {
		return "soft_delete.DeletedAt"
	}

	switch strings.ToLower(sqlType) {
	case "int", "tinyint":
		return "int"
	case "varchar", "text":
		return "string"
	case "timestamp", "datetime":
		return "time.Time"
	default:
		return "interface{}"
	}
}

func sqlToProtoType(sqlType string) string {
	switch strings.ToLower(sqlType) {
	case "int", "tinyint":
		return "int32"
	case "varchar", "text", "timestamp", "datetime":
		return "string"
	default:
		return "google.protobuf.Any"
	}
}

func zeroJudgment(typ string) string {
	switch typ {
	case "int":
		return `if data.%s != 0 {
                mod.%s = data.%s
            }`
	case "string":
		return `if data.%s != "" {
                mod.%s = data.%s
            }`
	case "time.Time":
		return `if !data.%s.IsZero() {
                mod.%s = data.%s
            }`
	default:
		return ""
	}
}

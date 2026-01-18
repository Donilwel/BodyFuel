package postgres

import (
	"fmt"
	"github.com/fatih/structs"
	"reflect"
	"strings"
	"time"
)

type Config struct {
	Username        string        `yaml:"user" env:"DB_USERNAME"`
	Password        string        `yaml:"password" env:"DB_PASSWORD,unset"`
	Database        string        `yaml:"database" env:"DB_NAME,unset"`
	Host            string        `yaml:"host" env:"HOSTNAME,unset"`
	MaxOpenConns    int           `yaml:"max_open_conn"`
	MaxIdleConns    int           `yaml:"max_idle_conn"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

func constructParamsRow(sequenceNumber int, fieldsNumber int) string {
	sb := new(strings.Builder)
	position := sequenceNumber * fieldsNumber

	sb.WriteString("(")

	for j := 0; j < fieldsNumber; j++ {
		sb.WriteString("$")
		sb.WriteString(fmt.Sprint(position + j + 1))

		if j == fieldsNumber-1 {
			continue
		}

		sb.WriteString(",")
	}

	sb.WriteString("),")

	return sb.String()
}

func structToParams(f any) map[string]any {
	newMap := structs.Map(f)
	resultMap := make(map[string]interface{})

	for k, v := range newMap {
		if v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil()) {
			continue
		}

		resultMap[k] = v
	}

	return resultMap
}

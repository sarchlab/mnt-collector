package collector

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sarchlab/mnt-backend/model"
	"github.com/sarchlab/mnt-collector/config"
	log "github.com/sirupsen/logrus"
)

type CaseSetting struct {
	Title    string
	Suite    string
	Command  string
	ParamStr string

	directory string
	param     model.Param
}

func generateCaseSettings(cases []config.Case) []CaseSetting {
	var extracted []CaseSetting
	for _, c := range cases {
		for _, a := range c.Args {
			ec := CaseSetting{
				Title:     c.Title,
				Suite:     c.Suite,
				Command:   c.Command,
				directory: c.Directory,
				param:     a,
			}

			val := reflect.ValueOf(a)
			typ := reflect.TypeOf(a)

			var paramStr string
			for i := 0; i < val.NumField(); i++ {
				field := val.Field(i)
				fieldType := typ.Field(i)

				if !field.IsZero() {
					yamlTag := fieldType.Tag.Get("yaml")
					yamlKey := strings.Split(yamlTag, ",")[0]
					str := fmt.Sprintf("-%s %v", yamlKey, field.Interface())
					paramStr += str + " "
				}
			}
			ec.ParamStr = paramStr

			extracted = append(extracted, ec)
			log.WithFields(log.Fields{
				"Title":   ec.Title,
				"Suite":   ec.Suite,
				"Command": ec.Command,
				"Params":  ec.ParamStr,
			}).Debug("Case extracted")
		}
	}
	if len(extracted) == 0 {
		log.Warn("No case extracted")
	}
	return extracted
}

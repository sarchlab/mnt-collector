package collector

import (
	"fmt"
	"reflect"

	"github.com/sarchlab/mnt-collector/config"
	"github.com/sarchlab/mnt-collector/mntbackend"
	log "github.com/sirupsen/logrus"
)

type Case struct {
	Title     string
	Suite     string
	Command   string
	ParamStrs []string

	directory   string
	param       mntbackend.Param
	RepeatTimes int32
}

func Run() {
	cases := extractCases(config.C.Cases, config.C.ProfileCollect.RepeatTimes)

	if config.C.TraceCollect.Enable {
		runTraceCollect(cases)
	}
	if config.C.ProfileCollect.Enable {
		runProfileCollect(cases)
	}
}

func extractCases(cases []config.Case, repeatTimes int32) []Case {
	var extracted []Case
	for _, c := range cases {
		for _, a := range c.Args {
			ec := Case{
				Title:       c.Title,
				Suite:       c.Suite,
				Command:     c.Command,
				directory:   c.Directory,
				param:       mntbackend.Param(a),
				RepeatTimes: repeatTimes,
			}

			val := reflect.ValueOf(a)
			typ := reflect.TypeOf(a)

			var paramStrs []string
			for i := 0; i < val.NumField(); i++ {
				field := val.Field(i)
				fieldType := typ.Field(i)

				yamlTag := fieldType.Tag.Get("yaml")
				str := fmt.Sprintf("-%s %v", yamlTag, field.Interface())
				paramStrs = append(paramStrs, str)
			}
			ec.ParamStrs = paramStrs

			extracted = append(extracted, ec)
			log.WithFields(log.Fields{
				"Title":   ec.Title,
				"Suite":   ec.Suite,
				"Command": ec.Command,
				"Params":  ec.ParamStrs,
				"Repeat":  ec.RepeatTimes,
			}).Info("Case extracted")
		}
	}
	return extracted
}

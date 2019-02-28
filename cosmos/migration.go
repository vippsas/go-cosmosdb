package cosmos

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/pkg/errors"
)

// 'migrations' is indexed by a string "{fromModelName}|{toModelName}"
type migrationFunc func(from, to interface{}) error

var migrations = make(map[string]migrationFunc)

// ModelNameRegexp defines the names that are accepted in the cosmosmodel:\"\" specifier (`^[a-zA-Z_]+/[0-9]+$`)
var ModelNameRegexp = regexp.MustCompile(`^[a-zA-Z_]+/[0-9]+$`)

func checkModelName(modelName string) {
	// Check that model is "name/<version>"
	if !ModelNameRegexp.MatchString(modelName) {
		panic(errors.New("The name given in cosmosmodel:\"<...>\" must match ModelNameRegexp "))
	}
}

func syncModelField(entityPtr Model) {
	v := reflect.ValueOf(entityPtr).Elem()
	structT := v.Type()
	n := structT.NumField()
	for i := 0; i != n; i++ {
		field := structT.Field(i)
		if field.Name == "Model" {
			if field.Tag.Get("json") != "model" {
				panic(errors.New("entity's Model does not have a `json:\"model\"` tag as required"))
			}
			modelName := field.Tag.Get("cosmosmodel")
			checkModelName(modelName)
			if modelName == "" {
				panic(errors.New("Model field does not have `cosmosmodel:\"...\"` tag as required"))
			}

			v.Field(i).SetString(modelName)
			break
		}
	}
}

func lookupModelField(entityPtr Model) (tagVal, fieldVal string) {
	v := reflect.ValueOf(entityPtr).Elem()
	structT := v.Type()
	n := structT.NumField()
	for i := 0; i != n; i++ {
		field := structT.Field(i)
		if field.Name == "Model" {
			if field.Tag.Get("json") != "model" {
				panic(errors.New("entity's Model does not have a `json:\"model\"` tag as required"))
			}
			tagVal = field.Tag.Get("cosmosmodel")
			if tagVal == "" {
				panic(errors.New("Model field does not have `cosmosmodel:\"...\"` tag as required"))
			}
			fieldVal = v.Field(i).String()
			return
		}
	}
	panic(errors.New("No Model field"))
}

// CheckModel will check that the Model attribute is correctly set; also return the value.
// Pass pointer to interface.
func CheckModel(entityPtr Model) string {
	tagVal, fieldVal := lookupModelField(entityPtr)
	if tagVal != fieldVal {
		panic(errors.New("Struct has a model field that disagree with the `cosmosmodel:\"...\"` specification"))
	}
	return tagVal
}

func AddMigration(fromPrototype, toPrototype Model, convFunc migrationFunc) (dummyResult struct{}) {
	fromTag, _ := lookupModelField(fromPrototype)
	toTag, _ := lookupModelField(toPrototype)
	key := fmt.Sprintf("%s|%s", fromTag, toTag)
	_, ok := migrations[key]
	if ok {
		panic(errors.Errorf("Several migrations from %s to %s", fromTag, toTag))
	}
	migrations[key] = convFunc
	//panic(errors.New("not implemented"))
	return
}

func postGet(entityPtr Model, txn *Transaction) error {
	// Always set Model to value in spec..
	syncModelField(entityPtr)
	return entityPtr.PostGet(txn)
}

func prePut(entityPtr Model, txn *Transaction) error {
	// This is not doing much but is a hook point for future additional code postPut
	return entityPtr.PrePut(txn)
}

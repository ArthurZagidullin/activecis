package main

import (
	"flag"
	"os"
	"errors"
	"strconv"
	"reflect"
	"fmt"
)

type S struct{
	Configfile string `required:"false" name:"config" default:"/etc/daemon.conf" description:"Конфигурационный файл"`
	Auth	bool	`required:"true" name:"auth" default:"false" description:"Необходиость аутентификации"`
	Count	int64	`required:"false" name:"count" default:"0" description:"Количество"`
	//Duration, float64, int64, Uint, uint64
}

func main() {
	s := &S{}
	if(len(os.Args) > 1){
		err := GetArguments(s)
		if err != nil {
			fmt.Println(getColor(err.Error(),"r"))
			os.Exit(1)
		}
		// fmt.Printf("%v\n",s)
	} else {
		fmt.Println("Не переданы нужные аргументы")
	}
}

type A struct {
	Val string
	Type string
	Name string
	Index int
	Req bool
}
func (a *A) String() string{
	return a.Val
}
func (a *A) Set(val string) error{
	if(len(val) > 0){
		a.Val = val
		return nil
	}
	return errors.New("Short string")
}

func GetArguments(s interface{}) error {
	val := reflect.TypeOf(s).Elem()
	var as []*A
	var err error
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		ft := field.Type.String()
		fmt.Println(getColor("Характерстики поля: ","g"),ft ,field.Tag.Get("required"), field.Tag.Get("name"), field.Tag.Get("default"),field.Tag.Get("description"))

		var a A
		a.Type = ft
		a.Index = i
		a.Name = field.Tag.Get("name")
		a.Req,err = strconv.ParseBool(field.Tag.Get("required"))
		if err != nil {return err}

		a.Set(field.Tag.Get("default"))

		flag.Var(&a, field.Tag.Get("name"), field.Tag.Get("description"))
		as = append(as, &a)
	}
	flag.Parse()
	return Pull(s, as)
}
func Pull(s interface{}, as []*A) error{
	for _,a := range as {
		if !CheckRequireArg(a) { return errors.New("Не передан обязательный аргумент: "+a.Name+"!") }
		var v interface{}
		var err error
		sv := reflect.ValueOf(s).Elem().Field(a.Index)
		// fmt.Println("Преобразовываемый тип: ",a.Type)
		switch a.Type {
			case "string": v = a.Val;
			case "bool": v,err = strconv.ParseBool(a.Val)
			case "int": v,err = strconv.ParseInt(a.Val,10,0)
			case "uint":
				var foo uint64
				foo,err = strconv.ParseUint(a.Val,10,64)
				v = uint(foo)
			case "uint64": v,err = strconv.ParseUint(a.Val,10,64)
			default: err = errors.New("Низвестный тип " + a.Type + " поля: " + a.Name + "!")
		}
		if err != nil { return err }
		sv.Set(reflect.ValueOf(v))
	}
	return nil
}
func WhoCame() []*flag.Flag {
	var res []*flag.Flag
	visit := func(f *flag.Flag) {
		res = append(res, f)
	}
	flag.Visit(visit)
	return res
}
func CheckRequireArg(a *A) bool {
	if !a.Req {
		return true			
	}
	for _,c := range WhoCame() {
		if a.Name == c.Name {
			return true
		}
	}
	return false
}
func getColor(s,c string) string {
	var res string
	switch c {
		case "g": res += "\x1b[32m"+s+"\x1b[0m"
		case "r": res += "\x1b[31m"+s+"\x1b[0m"
		default : res = s
	}
	return res
}
package main

import (
	"fmt"
	"net/http"
	"reflect"
	"bytes"
	"strings"
	"strconv"
	"errors"
)
  
type MyForm struct {
	UserName  string `required:"true" field:"name" name:"Имя пользователя" type:"text"`
	UserPassword string `required:"true" field:"password" name:"Пароль пользователя" type:"password"`
	Resident  bool `field:"resident" type:"radio" radio:"1;checked" name:"Резидент РФ"`
	NoResident bool `field:"resident" type:"radio" radio:"2" name:"Не резидент РФ"`
	Gender string `field:"gender" name:"Пол" type:"select" select:"Неизвестный=3;selected,Мужской=1,Женский=2"`
	Age int64 `field:"age" name:"Возраст" type:"text" default:"true"`
	Token string `field:"token" type:"hidden" default:"true"`
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
func typeAssert(v,t string) (interface{}, error) {
	var res interface{}
	var err error
	switch t {
		case "string": res = v;
		case "bool": 
			res, err = strconv.ParseBool(v)
		case "int": 
			res, err = strconv.Atoi(v)
		case "int64":
			res, err = strconv.ParseInt(v,10,64)
		default:
			err = errors.New("Приведение типа провалилось - неизвестный тип")
	}
	return res, err
}
func FormRead(d *MyForm, r *http.Request) (err error) {
	   	r.ParseMultipartForm(1024)
	   	t := reflect.TypeOf(d).Elem()
	   	fmt.Println("\n\t" + getColor("Пользователь отправил форму!","g") + "\n\t----------------------------\n")

	   	//Проходимся по всем свойствам структуры
	   	for i := 0; i < t.NumField(); i++ {
	   		f := t.Field(i)
	   		fmt.Print(f.Tag.Get("field"),":\t")
	   		// Проверяем, пришло ли чего-нибудь из формы в поле для текущего свойства
	   		if(len(r.Form[ f.Tag.Get("field") ] ) > 0 ){
	   			formValue := r.Form[f.Tag.Get("field")][0]
	   			fmt.Print(formValue,"\n")
		   		// Заполняем структуру пришедшими данными
			   	if formValue != "" {
			   		tf := f.Type.String()
				   	fmt.Println("Нужно привести строку к типу "+tf)
				   	vf,err := typeAssert(formValue, tf)
				   	if err != nil { return err }
				   	fmt.Println(getColor("Удачно!","g")," теперь тип ",reflect.TypeOf(vf))
	   				reflect.ValueOf(d).Elem().Field(i).Set( reflect.ValueOf(vf) )
			   	} else {
			   		if f.Tag.Get("required") != "" && f.Tag.Get("required") != "false" { 
			   			return errors.New("Обязательное поле "+f.Tag.Get("field")+" не заполнено!")
			   		} else {
			   			fmt.Print("\tданных нет\n")
			   		}
			   	}
		   	}
		   	fmt.Print("**********************\n\n")
	   	}
	   	return nil
}
func FormCreate(d *MyForm) string {
	var buffer bytes.Buffer
	buffer.WriteString("<form method=\"POST\" enctype=\"multipart/form-data\">")

	t := reflect.TypeOf(d).Elem()

	for i := 0; i < t.NumField(); i++ {
	    var ent string
	    var sel string
	    f := t.Field(i)
	    //Похоже это надо в label
	    name := f.Tag.Get("name")
	    //Тип поля
	    typ := f.Tag.Get("type")
	    //Имя поля
	    field := f.Tag.Get("field")
	    if(typ == "radio"){
	    	r := strings.Split(f.Tag.Get("radio"), ";")
	    	ent = "<input type=\""+typ+"\" name=\""+field+"\" value=\""+r[0]+"\" "
	    	if(len(r)>1){ ent += r[1] }
	    	ent += ">"

	    } else if(typ == "select"){
	    	sel = "<select>"
	    	s := strings.Split(f.Tag.Get("select"), ",")
	    	for _,opt := range s {
	    		o := strings.Split(opt,"=")
	    		sel += "<option value=\""+o[1]+"\">"+o[0]+"</option>"
	    	}
	    	sel += "</select>" 
	    } else {
	    	ent = "<input type=\""+typ+"\" name=\""+field+"\""
	    	if f.Tag.Get("default") == "true" {
	    		val := reflect.ValueOf(d).Elem().Field(i).Interface()
	    		foo, _ := toString(val)

	    		// fmt.Println("Val: ",val," field: ",field)
	    		ent += " value=\""+foo+"\" "
	    	}
	    	ent += ">"
	    }
	    if typ != "hidden"{ buffer.WriteString("<label for="+field+">"+name+": </label>") }
 
	    
	    buffer.WriteString(ent)
	    buffer.WriteString(sel)
	    buffer.WriteString("<br />")
	}
	buffer.WriteString("<button type=\"submit\">Отправить</button></form >")

	return buffer.String()
}
func toString(i interface{}) (string, error) {
    switch i.(type) {
    	case string:
        	return i.(string),nil
		case int64:
			return strconv.FormatInt(i.(int64),10),nil
      	default:
        	return "",errors.New("Неизвестный тип")
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
    // fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:],"<br >")
    var d *MyForm = &MyForm{
    	UserName: "Arthur",
		Age: 18,
		Token: "345625145123451234123412342345",
	}
	if r.Method == "POST" {
		err := FormRead(d, r)
		if err != nil {
			fmt.Println(getColor(err.Error(),"r"))
		} else {
			fmt.Println("Структура заполнена пришедшими параметрами: \n",d) 
		}
    }
	
	// f := FormCreate(fd)
	var buffer bytes.Buffer
	buffer.WriteString( "<!DOCTYPE html><html><header></header><body>")
	buffer.WriteString( FormCreate(d) )
	buffer.WriteString("</body></html>")
	// res := []byte(buffer.String())
	fmt.Fprintf(w, buffer.String())
	// w.Write([]byte(FormCreate(fd)))
}

func main() {
    http.HandleFunc("/", handler)
    fmt.Println("Слушаем "+getColor("localhost:8080","g")+":\n")
    http.ListenAndServe(":8080", nil)
}


package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

var packageName string

var typeIdentifiers = []string{"byte", "uint32", "uint64", "uint8", "uint16", "string", "bool", "object"}

type ChateauType struct {
	BasicKind string
	IsArray   bool
	IsSlice   bool
	ArrLength uint64
}

var ErrDuplicateObj = fmt.Errorf("object has been dublicated")

type Object struct {
	Name       string
	TextFields []string

	Formatted bool // перегнались ли текст филды в нормальные филды
	Fields    []Field
}

type Field struct {
	Name string
	Type ChateauType // тип

	IsRef  bool
	RefObj Object // указывает на вложенный объект если такой есть
}

type ObjectStorage struct {
	obj map[string]*Object
	m   sync.Mutex
}

func NewObjectStorage() *ObjectStorage {
	return &ObjectStorage{
		obj: make(map[string]*Object, 0),
	}
}

func (s *ObjectStorage) Add(obj Object) error {
	s.lock()
	if _, exists := s.obj[obj.Name]; exists {
		return ErrDuplicateObj
	}

	s.obj[obj.Name] = &obj

	s.unlock()
	return nil
}

func (s *ObjectStorage) lock() {
	s.m.Lock()
}

func (s *ObjectStorage) unlock() {
	s.m.Unlock()
}

func main() {
	file, err := os.Open("endpoints.chateau")
	if err != nil {
		fmt.Println(err)

		return
	}

	objStorage := NewObjectStorage()

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "package") {
			packageName = (strings.Split(line, " "))[1]
			continue
		}

		if strings.ReplaceAll(line, " ", "") == "" {
			continue
		}

		lastEndIndex := strings.LastIndex(line, "{")
		line = line[:lastEndIndex+1] // обрезаем все вайтспейсы после {

		if strings.HasPrefix(line, "object ") && strings.HasSuffix(line, "{") {
			objName := line[7 : strings.LastIndex(line, "{")-1]
			objName = strings.ReplaceAll(objName, " ", "")

			err := generateObject(objName, scanner, objStorage)
			if err != nil {
				fmt.Println(err, objName)

				return
			}
		}

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("obj`s was scanned successfully")
	fmt.Println("start describing fields")

	objStorage.describe()
	objStorage.codegen()

	fmt.Println(packageName)
	fmt.Println("Generated successfully")
}

func generateObject(name string, scanner *bufio.Scanner, objStorage *ObjectStorage) error {
	newObj := Object{
		Name: name,
	}

	for scanner.Scan() {
		intoObjLine := scanner.Text()

		if strings.ReplaceAll(intoObjLine, " ", "") == "" { // пропускаем все пустые строки
			continue
		}

		if strings.ReplaceAll(scanner.Text(), " ", "") == "}" { // если это последняя строка - просто пропускаем
			break
		}

		// чистим от пробелов слева
		for strings.HasPrefix(intoObjLine, " ") || intoObjLine[0] == '\t' {
			intoObjLine = intoObjLine[1:]
		}

		// чистим от пробелов справа
		for strings.HasSuffix(intoObjLine, " ") {
			intoObjLine = intoObjLine[:len(intoObjLine)-1]
		}

		newObj.TextFields = append(newObj.TextFields, intoObjLine)
	}

	return objStorage.Add(newObj)
}

func (s *ObjectStorage) describe() error {
	for _, obj := range s.obj {
		err := obj.describeFields(s)
		if err != nil {
			return err
		}

	}

	return nil
}

func (o *Object) describeFields(s *ObjectStorage) error {
	for _, txtField := range o.TextFields {
		var field Field

		matches := strings.Split(txtField, " ")
		if len(matches) != 2 {
			return fmt.Errorf("no 2 args into %s", txtField)
		}

		field.Name = matches[1]

		var err error
		var isSlice, isArray bool
		var arrLength uint64 = 0

		fmt.Println(isArray, isSlice, arrLength)
		// check if it is arr or slice type
		tp := matches[0]

		if strings.HasPrefix(tp, "[]") {
			isSlice = true
		} else if strings.HasPrefix(tp, "[") { // проверяем что это массив
			tpStrLen := tp[1:strings.LastIndex(tp, "]")] // получаем длинну массива

			arrLength, err = strconv.ParseUint(tpStrLen, 10, 64)
			if err != nil {
				return fmt.Errorf("cant parse lenght of array: %w", err)
			}

			isArray = true
		}

		field.Type.ArrLength = arrLength
		field.Type.IsArray = isArray
		field.Type.IsSlice = isSlice

		tpName := tp[strings.Index(tp, "]")+1:]

		switch tpName {
		case "string":
			field.Type.BasicKind = tpName
		case "byte":
			field.Type.BasicKind = tpName
		case "uint32":
			field.Type.BasicKind = tpName
		case "uint64":
			field.Type.BasicKind = tpName
		case "uint8":
			field.Type.BasicKind = tpName
		case "uint16":
			field.Type.BasicKind = tpName
		case "bool":
			field.Type.BasicKind = tpName
		case "int32":
			field.Type.BasicKind = tpName
		case "int64":
			field.Type.BasicKind = tpName
		case "int8":
			field.Type.BasicKind = tpName
		default:
			obj, exists := s.obj[tpName]
			if !exists {
				return fmt.Errorf("such type does not exists %s", tpName)
			}

			field.IsRef = true
			field.RefObj = *obj
		}
		o.Fields = append(o.Fields, field)
	}

	return nil
}

func (s *ObjectStorage) codegen() {
	f, err := os.Create("out/endpoints.go")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.WriteString("package " + packageName + "\n\n")
	if err != nil {
		panic(err)
	}

	for _, obj := range s.obj {
		_, err = f.WriteString("type " + obj.Name + " struct {" + "\n")
		if err != nil {
			panic(err)
		}

		obj.writeFieldsToFile(f)

		_, err = f.WriteString("}\n\n")
		if err != nil {
			panic(err)
		}

		obj.generateEncode(f)
		obj.generateDecode(f)
	}
}

func (o *Object) writeFieldsToFile(f *os.File) {
	for _, field := range o.Fields {

		_, err := f.WriteString(field.Name + " ") // записали имя
		if err != nil {
			panic(err)
		}

		if field.Type.IsSlice {
			_, err := f.WriteString("[]") // записали имя
			if err != nil {
				panic(err)
			}
		}

		if field.Type.IsArray {
			_, err := f.WriteString(fmt.Sprintf("[%d]", field.Type.ArrLength)) // записали имя
			if err != nil {
				panic(err)
			}
		}

		if field.IsRef {
			field.Type.BasicKind = field.RefObj.Name
		}

		_, err = f.WriteString(field.Type.BasicKind + "\n") // записали имя
		if err != nil {
			panic(err)
		}
	}

}

func (o *Object) generateEncode(f *os.File) {
	_, err := f.WriteString(fmt.Sprintf("func (%s *%s) Encode() {\n", o.Name, o.Name)) // записали имя
	if err != nil {
		panic(err)
	}

	_, err = f.WriteString("}\n\n") // записали имя
	if err != nil {
		panic(err)
	}
}

func (o *Object) generateDecode(f *os.File) {
	_, err := f.WriteString(fmt.Sprintf("func (%s *%s) Decode() {\n", o.Name, o.Name)) // записали имя
	if err != nil {
		panic(err)
	}

	_, err = f.WriteString("}\n\n") // записали имя
	if err != nil {
		panic(err)
	}
}

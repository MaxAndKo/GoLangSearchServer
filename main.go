package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
)

var AllowedFieldsForSorting = []string{"Name", "Id", "Age"}

type UserRoot struct {
	UserRow []struct {
		Id        int    `xml:"id"`
		FirstName string `xml:"first_name"`
		LastName  string `xml:"last_name"`
		Age       int    `xml:"age"`
		About     string `xml:"about"`
		Gender    string `xml:"gender"`
	} `xml:"row"`
}

func main() {

	http.HandleFunc("/", handlerHelloWorld)
	http.HandleFunc("/search/", handler)

	http.ListenAndServe(":8080", nil)

	//server, err := SearchServer(SearchRequest{
	//	Query:      "cillum",
	//	Limit:      10,
	//	Offset:     0,
	//	OrderField: "Id",
	//	OrderBy:    -1,
	//})
	//
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//fmt.Println(server)
}

func handlerHelloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
}

func handler(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	orderField := r.FormValue("order_field")

	orderBy, err := toIntIfNotEmpty(r.FormValue("order_by"))
	if err != nil {
		handleConvertError(err, w)
		return
	}

	limit, err := toIntIfNotEmpty(r.FormValue("limit"))
	if err != nil {
		handleConvertError(err, w)
		return
	}

	offset, err := toIntIfNotEmpty(r.FormValue("offset"))
	if err != nil {
		handleConvertError(err, w)
		return
	}

	searchServer, err := SearchServer(SearchRequest{
		Query:      query,
		OrderField: orderField,
		OrderBy:    orderBy,
		Limit:      limit,
		Offset:     offset,
	})

	if err != nil {
		handleConvertError(err, w)
		return
	}

	res, err := json.Marshal(searchServer)
	if err != nil {
		handleConvertError(err, w)
		return
	}
	_, err = w.Write(res)
	if err != nil {
		handleConvertError(err, w)
	}
}

func toIntIfNotEmpty(val string) (int, error) {
	if val == "" {
		return 0, nil
	}
	atoi, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}
	return atoi, err
}

func handleConvertError(inputError error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte(inputError.Error()))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func chooseOrderDirectionFunc(orderDirection int) (func(int) bool, error) {
	switch orderDirection {
	case 1:
		return func(n int) bool { return n < 0 }, nil
	case -1:
		return func(n int) bool { return n > 0 }, nil
	default:
		return nil, fmt.Errorf(`invalid order direction: %d`, orderDirection)
	}
}

func chooseSortFunc(orderDirection int, orderField string, users *[]User) (func(i, j int) bool, error) {
	if len(orderField) != 0 && !slices.Contains(AllowedFieldsForSorting, orderField) {
		return nil, errors.New("wrong order field error")
	}

	if orderDirection == 0 {
		return func(i, j int) bool {
			return false
		}, nil
	}

	directionFunc, err := chooseOrderDirectionFunc(orderDirection)
	if err != nil {
		return nil, err
	}

	switch orderField {
	case "Id":
		return func(i, j int) bool {
			return directionFunc((*users)[i].Id - (*users)[j].Id)
		}, nil
	case "Age":
		return func(i, j int) bool {
			return directionFunc((*users)[i].Age - (*users)[j].Age)
		}, nil
	default:
		return func(i, j int) bool {
			return directionFunc(strings.Compare((*users)[i].Name, (*users)[j].Name))
		}, nil
	}
}

func SearchServer(req SearchRequest) ([]User, error) {
	if req.Offset < 0 {
		return []User{}, errors.New("invalid offset")
	}

	if req.Limit < 0 {
		return []User{}, errors.New("invalid limit")
	}

	root := new(UserRoot)
	file, _ := os.ReadFile("dataset.xml")
	_ = xml.Unmarshal(file, root)

	resUsers := make([]User, 0)

	for _, row := range root.UserRow {
		name := row.FirstName + " " + row.LastName
		if name == req.Query || strings.Contains(row.About, req.Query) {
			resUsers = append(resUsers, User{
				Id:     row.Id,
				Name:   name,
				Age:    row.Age,
				Gender: row.Gender,
				About:  row.About,
			})
		}
	}

	if req.Offset > len(resUsers) {
		return []User{}, errors.New("offset out of range")
	}

	sortFunc, err := chooseSortFunc(req.OrderBy, req.OrderField, &resUsers)
	if err != nil {
		return []User{}, err
	}

	sort.SliceStable(resUsers, sortFunc)

	if req.Offset != 0 {
		resUsers = resUsers[req.Offset:]
	}

	if req.Limit != 0 {
		return resUsers[:req.Limit], nil
	}

	return resUsers, nil
}

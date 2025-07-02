package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"slices"
	"sort"
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
	server, err := SearchServer(SearchRequest{
		Query:      "cillum",
		Limit:      10,
		Offset:     0,
		OrderField: "Id",
		OrderBy:    -1,
	})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(server)
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

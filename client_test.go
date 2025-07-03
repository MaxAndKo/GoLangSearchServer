package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test(t *testing.T) {
	server := startSearchServer()

	searchClient := SearchClient{
		AccessToken: "12345",
		URL:         server.URL,
	}

	searchRequest := SearchRequest{
		Query:      "cillum",
		Limit:      10,
		Offset:     0,
		OrderField: "Id",
		OrderBy:    -1,
	}

	users, err := searchClient.FindUsers(searchRequest)
	if err != nil || users == nil {
		fmt.Errorf("Error")
	}
}

func startSearchServer() *httptest.Server {
	http.HandleFunc("/search/", handler)

	return httptest.NewServer(http.HandlerFunc(handler))
	//http.ListenAndServe(":8080", nil)
}

package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
    "sort"
    "encoding/json"
)

// ---- структуры для SearchServer ----
type UserXML struct {
	Id        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

type Dataset struct {
	Users []UserXML `xml:"row"`
}

// ---- SearchServer ----
func SearchServer(w http.ResponseWriter, r *http.Request) {
	limit := 0
	offset := 0
	orderBy := 0

	if l := r.FormValue("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := r.FormValue("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}
	if ob := r.FormValue("order_by"); ob != "" {
		fmt.Sscanf(ob, "%d", &orderBy)
	}
	query := r.FormValue("query")
	orderField := r.FormValue("order_field")

	if r.Header.Get("AccessToken") == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// для таймаута
	if query == "sleep" {
		time.Sleep(2 * time.Second)
	}

	// имитация внутренней ошибки
	if query == "FileNotFound" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	file, err := os.ReadFile("dataset.xml")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var ds Dataset
	if err := xml.Unmarshal(file, &ds); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// формируем пользователей
	users := make([]User, 0, len(ds.Users))
	for _, u := range ds.Users {
		name := strings.TrimSpace(u.FirstName + " " + u.LastName)
		users = append(users, User{
			Id:     u.Id,
			Name:   name,
			Age:    u.Age,
			About:  u.About,
			Gender: u.Gender,
		})
	}

	// фильтрация по query
	filtered := []User{}
	for _, u := range users {
		if query == "" || strings.Contains(u.Name, query) || strings.Contains(u.About, query) {
			filtered = append(filtered, u)
		}
	}

	// сортировка
	if orderField == "" {
		orderField = "Name"
	}
	switch orderField {
	case "Id":
		if orderBy >= 0 {
			SortByIdAsc(filtered)
		} else {
			SortByIdDesc(filtered)
		}
	case "Age":
		if orderBy >= 0 {
			SortByAgeAsc(filtered)
		} else {
			SortByAgeDesc(filtered)
		}
	case "Name":
		if orderBy >= 0 {
			SortByNameAsc(filtered)
		} else {
			SortByNameDesc(filtered)
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"ErrorBadOrderField"}`))
		return
	}

	if offset > len(filtered) {
		offset = len(filtered)
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	result := filtered[offset:end]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ---- вспомогательные сортировки ----
func SortByIdAsc(users []User)    { sort.Slice(users, func(i, j int) bool { return users[i].Id < users[j].Id }) }
func SortByIdDesc(users []User)   { sort.Slice(users, func(i, j int) bool { return users[i].Id > users[j].Id }) }
func SortByAgeAsc(users []User)   { sort.Slice(users, func(i, j int) bool { return users[i].Age < users[j].Age }) }
func SortByAgeDesc(users []User)  { sort.Slice(users, func(i, j int) bool { return users[i].Age > users[j].Age }) }
func SortByNameAsc(users []User)  { sort.Slice(users, func(i, j int) bool { return users[i].Name < users[j].Name }) }
func SortByNameDesc(users []User) { sort.Slice(users, func(i, j int) bool { return users[i].Name > users[j].Name }) }

// ---- вспомогательная функция теста ----
func RunClientTest(t *testing.T, client *SearchClient, req SearchRequest, wantErr string, wantUsers int) {
	resp, err := client.FindUsers(req)
	if wantErr != "" {
		if err == nil || !strings.Contains(err.Error(), wantErr) {
			t.Errorf("expected error %q, got %v", wantErr, err)
		}
		return
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if len(resp.Users) != wantUsers {
		t.Errorf("expected %d users, got %d", wantUsers, len(resp.Users))
	}
}

// ---- тесты ----
func TestSearchClient_AllCases(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{AccessToken: "token", URL: ts.URL}

	tests := []struct {
		name      string
		req       SearchRequest
		wantErr   string
		wantUsers int
	}{
		{"Limit negative", SearchRequest{Limit: -1}, "limit must be > 0", 0},
		{"Offset negative", SearchRequest{Limit: 1, Offset: -1}, "offset must be > 0", 0},
		{"Bad order field", SearchRequest{Limit: 1, OrderField: "Unknown"}, "OrderFeld", 0},
		{"Valid request small limit", SearchRequest{Limit: 2, Offset: 0}, "", 2},
		{"Valid request with query", SearchRequest{Limit: 3, Offset: 0, Query: "Boyd Wolf"}, "", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RunClientTest(t, client, tt.req, tt.wantErr, tt.wantUsers)
		})
	}
}

func TestSearchClient_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.TimeoutHandler(http.HandlerFunc(SearchServer), 1*time.Second, "timeout"))
	defer ts.Close()

	client := &SearchClient{AccessToken: "token", URL: ts.URL}
	req := SearchRequest{Limit: 1, Offset: 0, Query: "sleep"}
	_, err := client.FindUsers(req)
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected timeout error, got %v", err)
	}
}

func TestSearchClient_Unauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	client := &SearchClient{URL: ts.URL}
	req := SearchRequest{Limit: 1, Offset: 0}
	_, err := client.FindUsers(req)
	if err == nil || !strings.Contains(err.Error(), "Bad AccessToken") {
		t.Errorf("expected unauthorized error, got %v", err)
	}
}

func TestSearchClient_InternalServerError(t *testing.T) {
	// временно переименуем файл dataset.xml, чтобы сервер вернул 500
	_ = os.Rename("dataset.xml", "dataset.xml.bak")
	defer os.Rename("dataset.xml.bak", "dataset.xml")

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	client := &SearchClient{AccessToken: "token", URL: ts.URL}
	req := SearchRequest{Limit: 1, Offset: 0}
	_, err := client.FindUsers(req)
	if err == nil || !strings.Contains(err.Error(), "SearchServer fatal error") {
		t.Errorf("expected internal server error, got %v", err)
	}
}
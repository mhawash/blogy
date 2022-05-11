package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

var (
	// postAPI holds the the API url from which to get posts from
	postAPI = "https://api.hatchways.io/assessment/blog/posts"
	// sortByAV is used to hold the acceptable values for the QP sortBy
	sortByAV = []string{"id", "reads", "likes", "popularity"}
	// directionAV is used to hold the acceptable values of QP (sortBy)
	directionAV = []string{"desc", "asc"}
)

// Post is used for unmarshalling and marshalling the Post json object
type Post struct {
	Author     string   `json:"author"`
	AuthorID   int      `json:"authorId"`
	ID         int      `json:"id"`
	Likes      int      `json:"likes"`
	Popularity float64  `json:"popularity"`
	Reads      int      `json:"reads"`
	Tags       []string `json:"tags"`
}

// Posts object is used to hold a collection of Post/s
type Posts struct {
	Posts []Post `json:"posts"`
}

// checkStringInList returns true if s has been found in list
func checkStringInList(s string, list []string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// sortPosts sorts a slice of Post objects based on sortBy and direction params
func sortPosts(posts []Post, sortBy, direction string) {
	sortFn := func(i, j int) bool {
		switch sortBy {
		case "id":
			return posts[i].ID < posts[j].ID
		case "reads":
			return posts[i].Reads < posts[j].Reads
		case "likes":
			return posts[i].Likes < posts[j].Likes
		case "popularity":
			return posts[i].Popularity < posts[j].Popularity
		default:
			return false
		}
	}

	sort.SliceStable(posts, func(i, j int) bool {
		switch direction {
		case "asc":
			return sortFn(i, j)
		case "desc":
			return !sortFn(i, j)
		default:
			return false
		}
	})
}

// postsFetcher takes a slice of tags and returns a posts object
func postsFetcher(tags []string) (*Posts, error) {
	// postsRecord is used to Initialize a zero length map to record grabbed posts temporarily
	postsRecord := map[int]Post{}
	// pc is the posts object which will be returned
	pc := Posts{}

	for _, tag := range tags {
		resp, err := http.Get(fmt.Sprintf("%s?tag=%s", postAPI, tag))
		if err != nil {
			return nil, err
		}

		var tpc Posts // tpc is used temporarily for unmarshalling the json response
		if err := json.NewDecoder(resp.Body).Decode(&tpc); err != nil {
			return nil, err
		}
		resp.Body.Close()

		// Checking grabbed posts for duplicates and record new ones into postsRecord
		for _, post := range tpc.Posts {
			if _, ok := postsRecord[post.ID]; !ok {
				postsRecord[post.ID] = post
			}
		}
	}

	// Compining all retrieved posts from postsRecord into pc
	for _, post := range postsRecord {
		pc.Posts = append(pc.Posts, post)
	}

	return &pc, nil
}


func postsHandler(w http.ResponseWriter, r *http.Request) {
	// Header settings
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		// Query params
		qp := r.URL.Query()

		tags := qp.Get("tags")           // Required
		sortBy := qp.Get("sortBy")       // Optional with deafult value of (id)
		direction := qp.Get("direction") // Optional with deafult value of (asc)

		// Checking required query params
		if tags == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "tags parameter is required"}`))
			return
		}
		if sortBy == "" {
			sortBy = "id"
		}
		if direction == "" {
			direction = "asc"
		}

		// Checking the legality of query params
		if !checkStringInList(sortBy, sortByAV) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "sortBy parameter is invalid"}`))
			return
		} else if !checkStringInList(direction, directionAV) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "direction parameter is invalid"}`))
			return
		}

		// Fetching posts based on tags
		pc, err := postsFetcher(strings.Split(tags, ","))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal server error"}`))
			return
		}

		// Sorting the posts collection appropriately
		sortPosts(pc.Posts, sortBy, direction)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(pc)
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}
}

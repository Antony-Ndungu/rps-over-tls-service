// Package contract defines the data structures for the request and the response
package contract

// CatsRequest defines the data structure for the request
type CatsRequest struct {
	Cursor int
	Limit  int
}

// Cat is a data structure representing a cat.
type Cat struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Weight        int    `json:"weight"`
	CreatedOn     string `json:"createdOn"`
	LastUpdatedOn string `json:"lastUpdatedOn"`
}

// Cats is a collection of references to cats.
type Cats []*Cat

// CatsResponse defines the data structure for the response
type CatsResponse struct {
	Cats Cats
}

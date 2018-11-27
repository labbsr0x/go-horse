package matchers

import (
	"fmt"

	"github.com/tidwall/gjson"
)

// JSONMatcher https://github.com/tidwall/gjson#path-syntax
// A path is a series of keys separated by a dot. A key may contain special wildcard characters '*' and '?'.
// To access an array value use the index as the key. To get the number of elements in an array or to access a child path, use the '#' character.
// The dot and wildcard characters can be escaped with '\'.
//   {
// 	"name": {"first": "Tom", "last": "Anderson"},
// 	"age":37,
// 	"children": ["Sara","Alex","Jack"],
// 	"fav.movie": "Deer Hunter",
// 	"friends": [
// 	    {"first": "Dale", "last": "Murphy", "age": 44},
// 	    {"first": "Roger", "last": "Craig", "age": 68},
// 	    {"first": "Jane", "last": "Murphy", "age": 47}
// 	  ]
//   }
//   "name.last"          >> "Anderson"
//   "age"                >> 37
//   "children"           >> ["Sara","Alex","Jack"]
//   "children.#"         >> 3
//   "children.1"         >> "Alex"
//   "child*.2"           >> "Jack"
//   "c?ildren.0"         >> "Sara"
//   "fav\.movie"         >> "Deer Hunter"
//   "friends.#.first"    >> ["Dale","Roger","Jane"]
//   "friends.1.last"     >> "Craig"
//   You can also query an array for the first match by using #[...], or find all matches with #[...]#. Queries support the ==, !=, <, <=, >, >= comparison operators and the simple pattern matching % (like) and !% (not like) operators.
//   friends.#[last=="Murphy"].first    >> "Dale"
//   friends.#[last=="Murphy"]#.first   >> ["Dale","Jane"]
//   friends.#[age>45]#.last            >> ["Craig","Murphy"]
//   friends.#[first%"D*"].last         >> "Murphy"
//   friends.#[first!%"D*"].last        >> "Craig"
type JSONMatcher struct {
	Query         string
	ExpectedValue string
}

// Match lero lero
func (matcher JSONMatcher) Match(body []byte) (matches bool) {
	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> " + string(body))
	match := gjson.Get(string(body), matcher.Query)
	fmt.Printf("------- BODY : %s\n", match.Raw)
	return true
}

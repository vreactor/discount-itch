package ditch

type Category string

// Enum for all categories that can be found on itch.io
const (
	Games Category = "games"
)

// Array containing all categories.
var Categories = []Category{
	Games,
}

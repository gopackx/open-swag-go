// Package testtypes holds types used by tests elsewhere in the module, such
// as a second type named User to exercise schema component name collisions.
package testtypes

// User deliberately shares its name with test types in other packages.
type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

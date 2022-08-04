package context

type Context struct {
	HandleError func(err error) bool
	ShowInfo func(title, message string)
}

package service

// Service represents an OS service interface
type Service interface {
	Run() error
	Stop() error
}

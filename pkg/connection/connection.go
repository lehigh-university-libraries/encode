package connection

import "fmt"

type ConnectionProvider interface {
	Authenticate() error
	FetchReport(params map[string]string) ([]map[string]string, error)
}

type AuthService[T ConnectionProvider] struct {
	Provider T
	Name     string `yaml:"name"`
}

func NewAuthService[T ConnectionProvider](provider T, name string) *AuthService[T] {
	return &AuthService[T]{
		Provider: provider,
		Name:     name,
	}
}

func (s *AuthService[T]) FetchReport(params map[string]string) ([]map[string]string, error) {
	if fetcher, ok := any(s.Provider).(ConnectionProvider); ok {
		return fetcher.FetchReport(params)
	}
	return nil, fmt.Errorf("provider does not support report fetching")
}

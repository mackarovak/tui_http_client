package types

// AuthType определяет тип аутентификации.
type AuthType string

const (
	AuthNone   AuthType = "none"
	AuthBearer AuthType = "bearer"
)

// AuthConfig — конфигурация аутентификации для запроса.
type AuthConfig struct {
	Type  AuthType `json:"type"`
	Token string   `json:"token,omitempty"`

	// TokenVisible — только UI-состояние, не сериализуется.
	// true = показывать токен открытым текстом, false = маскировать.
	TokenVisible bool `json:"-"`
}

package auth

import "context"

type AuthenticationRepository interface {
	Read(ctx context.Context) (*Auth, error)
	Save(ctx context.Context, auth *Auth) error
}

type TokenRepository interface {
	Refresh(ctx context.Context, refreshToken string) (*Auth, error)
	Exchange(ctx context.Context, code string) (*Auth, error)
}

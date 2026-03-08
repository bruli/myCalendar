package auth

import "context"

type CreateTokens struct {
	tokenRepo TokenRepository
	authRepo  AuthenticationRepository
}

func (t CreateTokens) Create(ctx context.Context, code string) error {
	auth, err := t.tokenRepo.Exchange(ctx, code)
	if err != nil {
		return err

	}
	return t.authRepo.Save(ctx, auth)
}

func NewCreateTokens(tokenRepo TokenRepository, authRepo AuthenticationRepository) *CreateTokens {
	return &CreateTokens{tokenRepo: tokenRepo, authRepo: authRepo}
}

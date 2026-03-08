package auth

import "context"

type RefreshToken struct {
	authRepo  AuthenticationRepository
	tokenRepo TokenRepository
}

func (a RefreshToken) Refresh(ctx context.Context) error {
	auth, err := a.authRepo.Read(ctx)
	if err != nil {
		return err
	}
	auth, err = a.tokenRepo.Refresh(ctx, auth.RefreshToken())
	if err != nil {
		return err
	}
	return a.authRepo.Save(ctx, auth)
}

func NewRefreshToken(authRepo AuthenticationRepository, tokenRepo TokenRepository) *RefreshToken {
	return &RefreshToken{authRepo: authRepo, tokenRepo: tokenRepo}
}

package auth

import "time"

type Auth struct {
	accessToken, refreshToken string
	expiry                    time.Time
	tokenType                 string
}

func (a Auth) AccessToken() string {
	return a.accessToken
}

func (a Auth) RefreshToken() string {
	return a.refreshToken
}

func (a Auth) Expiry() time.Time {
	return a.expiry
}

func (a Auth) TokenType() string {
	return a.tokenType
}

func (a Auth) validate() error {
	switch {
	case a.accessToken == "":
		return ErrEmptyAccessToken
	case a.refreshToken == "":
		return ErrEmptyRefreshToken
	case a.tokenType == "":
		return ErrEmptyTokenType
	case a.expiry.IsZero():
		return ErrEmptyExpiry
	default:
		return nil
	}
}

func New(accessToken, refreshToken string, expiry time.Time, tokenType string) (*Auth, error) {
	a := Auth{accessToken: accessToken, refreshToken: refreshToken, expiry: expiry, tokenType: tokenType}
	if err := a.validate(); err != nil {
		return nil, err
	}
	return &a, nil
}

package auth

import (
	"github.com/akramboussanni/gocode/internal/jwt"
	"github.com/akramboussanni/gocode/internal/mailer"
	"github.com/akramboussanni/gocode/internal/model"
	"github.com/akramboussanni/gocode/internal/utils"
)

func GenerateTokenAndSendEmail(email, templateName, subject, url string, data any) (*model.Token, error) {
	token, err := utils.GetRandomToken(16)
	if err != nil {
		return nil, err
	}

	if dataMap, ok := data.(map[string]any); ok {
		dataMap["Token"] = token.Raw
	} else {
		data = map[string]any{"Token": token.Raw}
	}

	err = mailer.Send(templateName, []string{email}, subject, data)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func GenerateLogin(jwtToken jwt.Jwt) model.LoginTokens {
	sessionToken := jwtToken.WithType(model.CredentialJwt).GenerateToken()
	refreshToken := jwtToken.WithType(model.RefreshJwt).GenerateToken()

	return model.LoginTokens{
		Session: sessionToken,
		Refresh: refreshToken,
	}
}

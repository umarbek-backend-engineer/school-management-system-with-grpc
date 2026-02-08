package interceptors

import (
	"context"
	"os"
	"school_project_grpc/pkg/utils"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func Authentication_Intercepter(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// getting the token from metadata

	// skip spacific rpc that do not need authentication check
	skipMethods := map[string]bool{
		"/main.ExecsService/Login":          true,
		"/main.ExecsService/ForgotPassword": true,
		"/main.ExecsService/ResetPassword":  true,
	}

	if skipMethods[info.FullMethod] {
		return handler(ctx, req)
	}

	m, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "meta data missing")
	}

	authH, ok := m["authorization"]
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "meta data missing")
	}

	tokenStr := strings.TrimPrefix(authH[0], "Bearer ")
	tokenStr = strings.TrimSpace(tokenStr)

	ok = utils.JwtStore.IsLoggedOut(tokenStr)
	if ok {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized Access")
	}

	jwtSecrete := os.Getenv("JWT_SECRETE_STRING")

	// parsing the token
	parsedToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "Unauthenticated Access")
		}
		return []byte(jwtSecrete), nil
	})

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized Access")
	}

	// checking if the token is valid or not (expired?)
	if !parsedToken.Valid {
		return nil, status.Error(codes.Unauthenticated, "Invalid Token")
	}

	// converting and parcing the token claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Invalid Token")
	}

	// extracting the field of the token and passing them with the request to that the information can be accessed in ather handlers
	role, ok := claims["role"].(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Invalid Token")
	}

	userID := claims["uid"].(string)
	username := claims["username"].(string)
	expTimef64 := claims["exp"].(float64) // token expiry date
	expTimeInt64 := int64(expTimef64)

	newCtx := context.WithValue(ctx, "uid", userID)
	newCtx = context.WithValue(newCtx, "role", role)
	newCtx = context.WithValue(newCtx, "username", username)
	newCtx = context.WithValue(newCtx, "exp", expTimeInt64)

	return handler(newCtx, req)

}

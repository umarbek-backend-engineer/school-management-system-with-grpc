package utils

import (
	"context"
	"errors"
)

func Authorization(ctx context.Context, alloweRoles ...string) error {
	userrole, ok := ctx.Value("role").(string)
	if !ok {
		return errors.New("User not authorized for access: role not found")
	}

	for _, alrole := range alloweRoles {
		if alrole == userrole {
			return nil
		}
	}
	return errors.New("user not authorized for access")
}

package models

import (
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID          uuid.UUID            `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Username    string               `bun:"username,notnull,unique"`
	DisplayName string               `bun:"display_name,notnull"`
	CreatedAt   time.Time            `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt   time.Time            `bun:"updated_at,notnull,default:current_timestamp"`
	Credentials []WebAuthnCredential `bun:"rel:has-many,join:id=user_id"`
}

func (u *User) WebAuthnID() []byte {
	b, _ := u.ID.MarshalBinary()
	return b
}

func (u *User) WebAuthnName() string {
	return u.Username
}

func (u *User) WebAuthnDisplayName() string {
	return u.DisplayName
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	creds := make([]webauthn.Credential, len(u.Credentials))
	for i, c := range u.Credentials {
		creds[i] = c.ToWebAuthn()
	}
	return creds
}

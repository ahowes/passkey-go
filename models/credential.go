package models

import (
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type WebAuthnCredential struct {
	bun.BaseModel `bun:"table:webauthn_credentials,alias:wc"`

	ID              int64     `bun:"id,pk,autoincrement"`
	UserID          uuid.UUID `bun:"user_id,type:uuid,notnull"`
	CredentialID    []byte    `bun:"credential_id,notnull,unique"`
	PublicKey       []byte    `bun:"public_key,notnull"`
	AttestationType string    `bun:"attestation_type,notnull"`
	AAGUID          []byte    `bun:"aaguid"`
	SignCount       uint32    `bun:"sign_count,notnull,default:0"`
	CloneWarning    bool      `bun:"clone_warning,notnull,default:false"`
	Attachment      string    `bun:"attachment,default:''"`
	Transports      []string  `bun:"transports,type:text[],array"`

	FlagUserPresent    bool `bun:"flag_user_present,notnull,default:false"`
	FlagUserVerified   bool `bun:"flag_user_verified,notnull,default:false"`
	FlagBackupEligible bool `bun:"flag_backup_eligible,notnull,default:false"`
	FlagBackupState    bool `bun:"flag_backup_state,notnull,default:false"`

	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	User *User `bun:"rel:belongs-to,join:user_id=id"`
}

func (wc *WebAuthnCredential) ToWebAuthn() webauthn.Credential {
	transports := make([]protocol.AuthenticatorTransport, len(wc.Transports))
	for i, t := range wc.Transports {
		transports[i] = protocol.AuthenticatorTransport(t)
	}

	return webauthn.Credential{
		ID:              wc.CredentialID,
		PublicKey:       wc.PublicKey,
		AttestationType: wc.AttestationType,
		Transport:       transports,
		Flags: webauthn.CredentialFlags{
			UserPresent:    wc.FlagUserPresent,
			UserVerified:   wc.FlagUserVerified,
			BackupEligible: wc.FlagBackupEligible,
			BackupState:    wc.FlagBackupState,
		},
		Authenticator: webauthn.Authenticator{
			AAGUID:       wc.AAGUID,
			SignCount:     wc.SignCount,
			CloneWarning: wc.CloneWarning,
			Attachment:   protocol.AuthenticatorAttachment(wc.Attachment),
		},
	}
}

func NewWebAuthnCredentialFromLibrary(userID uuid.UUID, cred *webauthn.Credential) *WebAuthnCredential {
	transports := make([]string, len(cred.Transport))
	for i, t := range cred.Transport {
		transports[i] = string(t)
	}

	return &WebAuthnCredential{
		UserID:             userID,
		CredentialID:       cred.ID,
		PublicKey:          cred.PublicKey,
		AttestationType:    cred.AttestationType,
		AAGUID:             cred.Authenticator.AAGUID,
		SignCount:          cred.Authenticator.SignCount,
		CloneWarning:       cred.Authenticator.CloneWarning,
		Attachment:         string(cred.Authenticator.Attachment),
		Transports:         transports,
		FlagUserPresent:    cred.Flags.UserPresent,
		FlagUserVerified:   cred.Flags.UserVerified,
		FlagBackupEligible: cred.Flags.BackupEligible,
		FlagBackupState:    cred.Flags.BackupState,
	}
}

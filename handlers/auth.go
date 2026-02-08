package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/ahowes/passkey-go/models"
)

type AuthHandler struct {
	webAuthn *webauthn.WebAuthn
	db       *bun.DB
}

func NewAuthHandler(wa *webauthn.WebAuthn, db *bun.DB) *AuthHandler {
	return &AuthHandler{webAuthn: wa, db: db}
}

func (h *AuthHandler) BeginRegistration(c *gin.Context) {
	var req struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}
	if req.DisplayName == "" {
		req.DisplayName = req.Username
	}

	user := &models.User{}
	err := h.db.NewSelect().
		Model(user).
		Where("username = ?", req.Username).
		Relation("Credentials").
		Scan(c.Request.Context())

	if err != nil {
		user = &models.User{
			ID:          uuid.New(),
			Username:    req.Username,
			DisplayName: req.DisplayName,
		}
		if _, err := h.db.NewInsert().Model(user).Exec(c.Request.Context()); err != nil {
			log.Printf("failed to create user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
	}

	options, sessionData, err := h.webAuthn.BeginRegistration(user)
	if err != nil {
		log.Printf("failed to begin registration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin registration"})
		return
	}

	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		log.Printf("failed to marshal session data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	session := sessions.Default(c)
	session.Set("webauthn_registration", string(sessionDataJSON))
	if err := session.Save(); err != nil {
		log.Printf("failed to save session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, options)
}

func (h *AuthHandler) FinishRegistration(c *gin.Context) {
	session := sessions.Default(c)
	rawSessionData := session.Get("webauthn_registration")
	if rawSessionData == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no registration in progress"})
		return
	}

	var sessionData webauthn.SessionData
	if err := json.Unmarshal([]byte(rawSessionData.(string)), &sessionData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid session data"})
		return
	}

	userID, err := uuid.FromBytes(sessionData.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID in session"})
		return
	}

	user := &models.User{}
	if err := h.db.NewSelect().
		Model(user).
		Where("u.id = ?", userID).
		Relation("Credentials").
		Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	credential, err := h.webAuthn.FinishRegistration(user, sessionData, c.Request)
	if err != nil {
		log.Printf("failed to finish registration: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "registration failed"})
		return
	}

	dbCred := models.NewWebAuthnCredentialFromLibrary(user.ID, credential)
	if _, err := h.db.NewInsert().Model(dbCred).Exec(c.Request.Context()); err != nil {
		log.Printf("failed to save credential: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save credential"})
		return
	}

	session.Delete("webauthn_registration")
	session.Set("user_id", user.ID.String())
	if err := session.Save(); err != nil {
		log.Printf("failed to save session: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *AuthHandler) BeginLogin(c *gin.Context) {
	options, sessionData, err := h.webAuthn.BeginDiscoverableLogin()
	if err != nil {
		log.Printf("failed to begin login: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin login"})
		return
	}

	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		log.Printf("failed to marshal session data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	session := sessions.Default(c)
	session.Set("webauthn_login", string(sessionDataJSON))
	if err := session.Save(); err != nil {
		log.Printf("failed to save session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, options)
}

func (h *AuthHandler) FinishLogin(c *gin.Context) {
	session := sessions.Default(c)
	rawSessionData := session.Get("webauthn_login")
	if rawSessionData == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no login in progress"})
		return
	}

	var sessionData webauthn.SessionData
	if err := json.Unmarshal([]byte(rawSessionData.(string)), &sessionData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid session data"})
		return
	}

	var resolvedUser *models.User

	handler := func(rawID, userHandle []byte) (webauthn.User, error) {
		userID, err := uuid.FromBytes(userHandle)
		if err != nil {
			return nil, err
		}

		user := &models.User{}
		if err := h.db.NewSelect().
			Model(user).
			Where("u.id = ?", userID).
			Relation("Credentials").
			Scan(c.Request.Context()); err != nil {
			return nil, err
		}

		resolvedUser = user
		return user, nil
	}

	credential, err := h.webAuthn.FinishDiscoverableLogin(handler, sessionData, c.Request)
	if err != nil {
		log.Printf("failed to finish login: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "login failed"})
		return
	}

	// Update sign count and clone warning in the database
	if _, err := h.db.NewUpdate().
		Model((*models.WebAuthnCredential)(nil)).
		Set("sign_count = ?", credential.Authenticator.SignCount).
		Set("clone_warning = ?", credential.Authenticator.CloneWarning).
		Where("credential_id = ?", credential.ID).
		Exec(c.Request.Context()); err != nil {
		log.Printf("failed to update credential: %v", err)
	}

	session.Delete("webauthn_login")
	session.Set("user_id", resolvedUser.ID.String())
	if err := session.Save(); err != nil {
		log.Printf("failed to save session: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "username": resolvedUser.Username})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		log.Printf("failed to save session: %v", err)
	}
	c.Redirect(http.StatusFound, "/")
}

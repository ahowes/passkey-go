package handlers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/ahowes/passkey-go/models"
)

type PageHandler struct {
	db *bun.DB
}

func NewPageHandler(db *bun.DB) *PageHandler {
	return &PageHandler{db: db}
}

func (h *PageHandler) Index(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	c.HTML(http.StatusOK, "index.html", gin.H{
		"LoggedIn": userID != nil,
	})
}

func (h *PageHandler) Register(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", nil)
}

func (h *PageHandler) Login(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func (h *PageHandler) Dashboard(c *gin.Context) {
	rawUserID, _ := c.Get("user_id")
	userIDStr, ok := rawUserID.(string)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	user := &models.User{}
	if err := h.db.NewSelect().
		Model(user).
		Where("u.id = ?", userID).
		Relation("Credentials").
		Scan(c.Request.Context()); err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"User":            user,
		"CredentialCount": len(user.Credentials),
	})
}

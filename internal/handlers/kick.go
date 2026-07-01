package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/twirapp/twir/apps/twitch-mock/internal/config"
)

func (s *Server) kickToken(c *gin.Context) {
	grantType := c.DefaultQuery("grant_type", c.PostForm("grant_type"))

	switch grantType {
	case "client_credentials":
		s.logger.Info("kick token request",
			slog.String("grant_type", grantType),
			slog.String("access_token", config.MockKickAppToken),
		)
		c.JSON(http.StatusOK, gin.H{
			"access_token": config.MockKickAppToken,
			"token_type":   "bearer",
			"expires_in":   99999999,
		})
	case "authorization_code", "refresh_token":
		s.logger.Info("kick token request",
			slog.String("grant_type", grantType),
			slog.String("access_token", config.MockKickUserToken),
		)
		c.JSON(http.StatusOK, gin.H{
			"access_token":  config.MockKickUserToken,
			"token_type":    "bearer",
			"expires_in":    99999999,
			"refresh_token": "mock-kick-refresh",
			"scope":         "",
		})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}
}

func (s *Server) kickAuthorize(c *gin.Context) {
	redirect := s.config.SiteBaseURL + "/kick/login"
	query := "?code=" + "mock_kick_code_1234567890"
	if stateValue := c.Query("state"); stateValue != "" {
		query += "&state=" + stateValue
	}

	c.Redirect(http.StatusFound, redirect+query)
}

func (s *Server) kickIntrospect(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"active":    true,
			"client_id": "mock",
			"token_type": "app",
			"scope":     "",
			"exp":       99999999,
		},
		"message": "OK",
	})
}

func (s *Server) kickUsers(c *gin.Context) {
	ids := c.QueryArray("id")
	if len(ids) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"data":    []gin.H{kickBroadcasterUser()},
			"message": "OK",
		})
		return
	}

	data := make([]gin.H, 0, len(ids))
	for _, idStr := range ids {
		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil {
			continue
		}

		switch id {
		case config.MockKickBroadcasterID:
			data = append(data, kickBroadcasterUser())
		case config.MockKickBotID:
			data = append(data, kickBotUser())
		default:
			data = append(data, gin.H{
				"email":           fmt.Sprintf("user%d@kick.com", id),
				"name":            fmt.Sprintf("User%d", id),
				"profile_picture": "",
				"user_id":         id,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "OK",
	})
}

func (s *Server) kickLivestreams(c *gin.Context) {
	broadcasterIDs := c.QueryArray("broadcaster_user_id")
	categoryIDParam := c.Query("category_id")
	languageParam := c.Query("language")
	limitParam := c.DefaultQuery("limit", "25")

	stream, online := s.state.StreamSnapshot()
	if !online {
		c.JSON(http.StatusOK, gin.H{
			"data":    []any{},
			"message": "OK",
		})
		return
	}

	title, gameID, gameName, viewerCount := s.state.StreamMeta()
	startedAt, _ := stream["started_at"].(string)

	catID, _ := strconv.Atoi(gameID)
	if catID == 0 {
		catID = config.MockKickCategoryID
	}

	_ = categoryIDParam
	_ = languageParam
	_ = limitParam
	_ = broadcasterIDs

	livestream := gin.H{
		"broadcaster_user_id": config.MockKickBroadcasterID,
		"channel_id":          config.MockKickBroadcasterID,
		"slug":                config.MockKickChannelSlug,
		"stream_title":        title,
		"streamer_username":   config.MockBroadcasterName,
		"profile_picture":     "",
		"thumbnail":           "",
		"viewer_count":        viewerCount,
		"language":            "en",
		"is_live":             true,
		"tags":                []string{},
		"has_mature_content":  false,
		"started_at":          startedAt,
		"custom_tags":         []string{},
		"category": gin.H{
			"id":        catID,
			"name":      gameName,
			"thumbnail": "",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    []gin.H{livestream},
		"message": "OK",
	})
}

func (s *Server) kickChannels(c *gin.Context) {
	broadcasterIDs := c.QueryArray("broadcaster_user_id")
	slugs := c.QueryArray("slug")

	title, gameID, gameName, _ := s.state.StreamMeta()
	online := s.state.IsStreamOnline()

	categoryID, _ := strconv.Atoi(gameID)
	if categoryID == 0 {
		categoryID = config.MockKickCategoryID
	}

	channel := gin.H{
		"broadcaster_user_id":        config.MockKickBroadcasterID,
		"slug":                       config.MockKickChannelSlug,
		"channel_description":        "Mock Kick channel",
		"stream_title":               title,
		"active_subscribers_count":   0,
		"canceled_subscribers_count": 0,
		"banner_picture":             "",
		"category": gin.H{
			"id":        categoryID,
			"name":      gameName,
			"thumbnail": "",
		},
		"stream": gin.H{
			"is_live":       online,
			"viewer_count":  0,
			"language":      "en",
			"thumbnail":     "",
			"start_time":    "",
			"is_mature":     false,
			"custom_tags":   []string{},
			"url":           "",
			"key":           "",
		},
	}

	if len(broadcasterIDs) > 0 {
		data := make([]gin.H, 0, len(broadcasterIDs))
		for _, idStr := range broadcasterIDs {
			id, err := strconv.Atoi(strings.TrimSpace(idStr))
			if err != nil {
				continue
			}
			ch := cloneKickChannel(channel)
			ch["broadcaster_user_id"] = id
			data = append(data, ch)
		}
		c.JSON(http.StatusOK, gin.H{"data": data, "message": "OK"})
		return
	}

	if len(slugs) > 0 {
		data := make([]gin.H, 0, len(slugs))
		for range slugs {
			data = append(data, cloneKickChannel(channel))
		}
		c.JSON(http.StatusOK, gin.H{"data": data, "message": "OK"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    []gin.H{channel},
		"message": "OK",
	})
}

func cloneKickChannel(ch gin.H) gin.H {
	clone := make(gin.H, len(ch))
	for k, v := range ch {
		clone[k] = v
	}
	return clone
}

func (s *Server) kickChat(c *gin.Context) {
	var body struct {
		BroadcasterUserID int    `json:"broadcaster_user_id"`
		Content           string `json:"content"`
		ReplyToMessageID  string `json:"reply_to_message_id,omitempty"`
		Type              string `json:"type"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Bad Request"})
		return
	}

	if body.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required", "message": "Bad Request"})
		return
	}

	if body.Type == "" {
		body.Type = "bot"
	}

	s.logger.Info("kick chat message received",
		slog.Int("broadcaster_user_id", body.BroadcasterUserID),
		slog.String("content", body.Content),
		slog.String("type", body.Type),
	)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"is_sent":    true,
			"message_id": uuid.NewString(),
		},
		"message": "OK",
	})
}

func (s *Server) kickModerationBans(c *gin.Context) {
	var body struct {
		BroadcasterUserID int    `json:"broadcaster_user_id"`
		UserID            int    `json:"user_id"`
		Duration          *int   `json:"duration,omitempty"`
		Reason            string `json:"reason,omitempty"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Bad Request"})
		return
	}

	if body.UserID == 0 {
		body.UserID = 99999
	}
	if body.BroadcasterUserID == 0 {
		body.BroadcasterUserID = config.MockKickBroadcasterID
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    gin.H{},
		"message": "OK",
	})
}

func (s *Server) kickModerationUnban(c *gin.Context) {
	var body struct {
		BroadcasterUserID int `json:"broadcaster_user_id"`
		UserID            int `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Bad Request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    gin.H{},
		"message": "OK",
	})
}

func (s *Server) kickCategories(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q is required", "message": "Bad Request"})
		return
	}

	data := []gin.H{
		{
			"id":        config.MockKickCategoryID,
			"name":      config.MockKickCategoryName,
			"thumbnail": "",
		},
		{
			"id":        2,
			"name":      "IRL",
			"thumbnail": "",
		},
		{
			"id":        3,
			"name":      "Music",
			"thumbnail": "",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "OK",
	})
}

func (s *Server) kickEventSubscriptions(c *gin.Context) {
	var body struct {
		URL        string   `json:"url"`
		EventTypes []string `json:"events"`
		Secret     string   `json:"secret,omitempty"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Bad Request"})
		return
	}

	if body.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is required", "message": "Bad Request"})
		return
	}

	sub := s.webhook.CreateSubscription(body.URL, body.EventTypes, body.Secret)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":         sub.ID,
			"url":        sub.URL,
			"events":     sub.EventTypes,
			"created_at": sub.CreatedAt.Format(time.RFC3339),
		},
		"message": "OK",
	})
}

func kickBroadcasterUser() gin.H {
	return gin.H{
		"email":           "mock@kick.com",
		"name":            "MockStreamer",
		"profile_picture": "",
		"user_id":         config.MockKickBroadcasterID,
	}
}

func kickBotUser() gin.H {
	return gin.H{
		"email":           "bot@kick.com",
		"name":            "MockBot",
		"profile_picture": "",
		"user_id":         config.MockKickBotID,
	}
}

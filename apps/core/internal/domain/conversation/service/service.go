package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/model"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/pkg/database"
	"gorm.io/gorm"
)

type ConversationService struct {
	db *gorm.DB
}

type ConversationSummary struct {
	ID            int64      `json:"id"`
	Type          string     `json:"type"`
	Title         string     `json:"title"`
	AvatarURL     string     `json:"avatar_url,omitempty"`
	LastMessage   string     `json:"last_message"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	UnreadCount   int64      `json:"unread_count"`
	IsMuted       bool       `json:"is_muted"`
}

const (
	DefaultConversationListLimit = 20
	MaxConversationListLimit     = 100
)

// ConversationListQuery controls the authenticated conversation list.
// Cursor is the last conversation id returned by the previous page.
type ConversationListQuery struct {
	Limit  int
	Cursor *int64
}

type ConversationMessage struct {
	ID             int64     `json:"id"`
	ConversationID int64     `json:"conversation_id"`
	SenderID       int64     `json:"sender_id"`
	SenderName     string    `json:"sender_name"`
	SenderAvatar   string    `json:"sender_avatar,omitempty"`
	Content        string    `json:"content"`
	MessageType    string    `json:"message_type"`
	IsRead         bool      `json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`
}

const (
	DefaultConversationMessageLimit = 50
	MaxConversationMessageLimit     = 100
)

// ConversationMessageListQuery controls message history pagination.
// Cursor is the oldest message id from the previous page; the next page
// contains older messages.
type ConversationMessageListQuery struct {
	Limit  int
	Cursor *int64
}

type CreateConversationInput struct {
	Title          string  `json:"title"`
	Type           string  `json:"type"`
	ParticipantIDs []int64 `json:"participant_ids"`
}

type SendMessageInput struct {
	Content     string `json:"content"`
	MessageType string `json:"message_type"`
}

const MaxConversationMessageLength = 5000

type UpdateConversationSettingsInput struct {
	IsMuted *bool `json:"is_muted"`
}

var ErrConversationNotFound = errors.New("conversation not found")
var ErrConversationForbidden = errors.New("conversation forbidden")
var ErrConversationInvalidInput = errors.New("conversation invalid input")
var ErrConversationAlreadyExists = errors.New("conversation already exists")

func NewConversationService(db *gorm.DB) *ConversationService {
	if db == nil {
		db = database.DB
	}
	return &ConversationService{db: db}
}

func (s *ConversationService) List(ctx context.Context, userID int64, query ConversationListQuery) ([]ConversationSummary, *int64, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = DefaultConversationListLimit
	}
	if limit > MaxConversationListLimit {
		limit = MaxConversationListLimit
	}

	memberships, nextCursor, err := s.loadMemberships(ctx, userID, query.Cursor, limit)
	if err != nil {
		return nil, nil, err
	}
	if len(memberships) == 0 {
		return []ConversationSummary{}, nil, nil
	}

	conversationIDs := make([]int64, 0, len(memberships))
	for _, membership := range memberships {
		conversationIDs = append(conversationIDs, membership.ConversationID)
	}

	var conversations []model.Conversation
	if err := s.db.WithContext(ctx).
		Preload("Participants.User.Profile").
		Where("id IN ?", conversationIDs).
		Find(&conversations).Error; err != nil {
		return nil, nil, err
	}
	conversationByID := make(map[int64]model.Conversation, len(conversations))
	for _, conversation := range conversations {
		conversationByID[conversation.ID] = conversation
	}

	latestMessages, err := s.loadLatestMessages(ctx, conversationIDs)
	if err != nil {
		return nil, nil, err
	}
	unreadCounts, err := s.loadUnreadCounts(ctx, userID, conversationIDs)
	if err != nil {
		return nil, nil, err
	}

	summaries := make([]ConversationSummary, 0, len(conversations))
	for _, membership := range memberships {
		conversation, exists := conversationByID[membership.ConversationID]
		if !exists {
			continue
		}
		summary := s.buildConversationSummary(
			userID,
			conversation,
			membership,
			latestMessages[conversation.ID],
			unreadCounts[conversation.ID],
		)
		if summary.ID == 0 {
			continue
		}
		summaries = append(summaries, summary)
	}

	return summaries, nextCursor, nil
}

func (s *ConversationService) Create(ctx context.Context, userID int64, input CreateConversationInput) (*ConversationSummary, error) {
	title := strings.TrimSpace(input.Title)
	conversationType := strings.ToLower(strings.TrimSpace(input.Type))
	if title == "" || len(input.ParticipantIDs) == 0 {
		return nil, ErrConversationInvalidInput
	}
	if conversationType == "" {
		conversationType = "group"
	}
	if conversationType != "direct" && conversationType != "group" {
		return nil, ErrConversationInvalidInput
	}
	if conversationType == "direct" && len(input.ParticipantIDs) != 1 {
		return nil, ErrConversationInvalidInput
	}

	participantIDs, err := validateParticipantIDs(userID, input.ParticipantIDs)
	if err != nil {
		return nil, err
	}

	conversation := model.Conversation{
		Type:  conversationType,
		Title: title,
	}

	var existingConversation *model.Conversation
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := validateActiveParticipants(tx, participantIDs); err != nil {
			return err
		}
		if conversationType == "direct" {
			existing, err := findExistingDirectConversation(ctx, tx, userID, participantIDs[0])
			if err != nil {
				return err
			}
			if existing != nil {
				existingConversation = existing
				return nil
			}
		}

		if err := tx.Create(&conversation).Error; err != nil {
			return err
		}

		allParticipantIDs := append(participantIDs, userID)
		participants := make([]model.ConversationParticipant, 0, len(allParticipantIDs))
		for _, participantID := range allParticipantIDs {
			role := "member"
			if participantID == userID {
				role = "owner"
			}
			participants = append(participants, model.ConversationParticipant{
				ConversationID: conversation.ID,
				UserID:         participantID,
				Role:           role,
				JoinedAt:       time.Now().UTC(),
			})
		}

		return tx.Create(&participants).Error
	}); err != nil {
		if errors.Is(err, ErrConversationInvalidInput) {
			return nil, err
		}
		return nil, err
	}
	if existingConversation != nil {
		summary := conversationSummary(*existingConversation, userID)
		return &summary, ErrConversationAlreadyExists
	}

	return &ConversationSummary{
		ID:          conversation.ID,
		Type:        conversation.Type,
		Title:       conversation.Title,
		LastMessage: "",
		UnreadCount: 0,
		IsMuted:     false,
	}, nil
}

func validateParticipantIDs(userID int64, participantIDs []int64) ([]int64, error) {
	seen := make(map[int64]struct{}, len(participantIDs))
	validated := make([]int64, 0, len(participantIDs))
	for _, participantID := range participantIDs {
		if participantID <= 0 || participantID == userID {
			return nil, ErrConversationInvalidInput
		}
		if _, exists := seen[participantID]; exists {
			return nil, ErrConversationInvalidInput
		}
		seen[participantID] = struct{}{}
		validated = append(validated, participantID)
	}
	return validated, nil
}

func validateActiveParticipants(tx *gorm.DB, participantIDs []int64) error {
	var activeCount int64
	if err := tx.Model(&model.User{}).
		Where("id IN ? AND status = ?", participantIDs, 0).
		Count(&activeCount).Error; err != nil {
		return err
	}
	if activeCount != int64(len(participantIDs)) {
		return ErrConversationInvalidInput
	}
	return nil
}

func findExistingDirectConversation(ctx context.Context, tx *gorm.DB, userID, participantID int64) (*model.Conversation, error) {
	var conversation model.Conversation
	err := tx.WithContext(ctx).
		Preload("Participants.User.Profile").
		Joins("JOIN conversation_participants AS cp_self ON cp_self.conversation_id = conversations.id AND cp_self.user_id = ?", userID).
		Joins("JOIN conversation_participants AS cp_target ON cp_target.conversation_id = conversations.id AND cp_target.user_id = ?", participantID).
		Where("conversations.type = ?", "direct").
		Where("(SELECT COUNT(*) FROM conversation_participants AS cp_count WHERE cp_count.conversation_id = conversations.id) = ?", 2).
		First(&conversation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

func conversationSummary(conversation model.Conversation, userID int64) ConversationSummary {
	title := conversation.Title
	if title == "" {
		title = defaultConversationTitle(conversation.Participants, userID)
	}
	return ConversationSummary{
		ID:          conversation.ID,
		Type:        conversation.Type,
		Title:       title,
		AvatarURL:   otherParticipantAvatar(conversation.Participants, userID),
		LastMessage: "",
		UnreadCount: 0,
		IsMuted:     participantMuted(conversation.Participants, userID),
	}
}

func participantMuted(participants []model.ConversationParticipant, userID int64) bool {
	for _, participant := range participants {
		if participant.UserID == userID {
			return participant.IsMuted
		}
	}
	return false
}

func (s *ConversationService) Messages(ctx context.Context, userID, conversationID int64, query ConversationMessageListQuery) ([]ConversationMessage, *int64, error) {
	membership, err := s.membershipForUser(ctx, userID, conversationID)
	if err != nil {
		return nil, nil, err
	}
	limit := query.Limit
	if limit <= 0 {
		limit = DefaultConversationMessageLimit
	}
	if limit > MaxConversationMessageLimit {
		limit = MaxConversationMessageLimit
	}

	var messages []model.Message
	dbQuery := s.db.WithContext(ctx).
		Preload("Sender.Profile").
		Where("conversation_id = ?", conversationID).
		Order("id desc").
		Limit(limit + 1)
	if query.Cursor != nil {
		dbQuery = dbQuery.Where("id < ?", *query.Cursor)
	}
	if err := dbQuery.Find(&messages).Error; err != nil {
		return nil, nil, err
	}

	var nextCursor *int64
	if len(messages) > limit {
		value := messages[limit-1].ID
		nextCursor = &value
		messages = messages[:limit]
	}
	for left, right := 0, len(messages)-1; left < right; left, right = left+1, right-1 {
		messages[left], messages[right] = messages[right], messages[left]
	}

	now := time.Now().UTC()
	if err := s.db.WithContext(ctx).
		Model(&model.ConversationParticipant{}).
		Where("id = ?", membership.ID).
		Update("last_read_at", now).Error; err != nil {
		return nil, nil, err
	}
	if err := s.db.WithContext(ctx).
		Model(&model.Message{}).
		Where("conversation_id = ? AND sender_id <> ?", conversationID, userID).
		Update("is_read", true).Error; err != nil {
		return nil, nil, err
	}

	result := make([]ConversationMessage, 0, len(messages))
	for _, message := range messages {
		result = append(result, mapConversationMessage(message))
	}

	return result, nextCursor, nil
}

func (s *ConversationService) SendMessage(ctx context.Context, userID, conversationID int64, input SendMessageInput) (*ConversationMessage, error) {
	content := strings.TrimSpace(input.Content)
	if content == "" || len([]rune(content)) > MaxConversationMessageLength {
		return nil, ErrConversationInvalidInput
	}

	messageType := strings.ToLower(strings.TrimSpace(input.MessageType))
	if messageType == "" {
		messageType = "text"
	}
	switch messageType {
	case "text", "image", "file":
	default:
		return nil, ErrConversationInvalidInput
	}

	if _, err := s.membershipForUser(ctx, userID, conversationID); err != nil {
		return nil, err
	}

	message := model.Message{
		ConversationID: conversationID,
		SenderID:       userID,
		Content:        content,
		MessageType:    messageType,
		IsRead:         false,
		CreatedAt:      time.Now().UTC(),
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&message).Error; err != nil {
			return err
		}
		return tx.Model(&model.Conversation{}).
			Where("id = ?", conversationID).
			Update("updated_at", message.CreatedAt).Error
	}); err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Preload("Sender.Profile").First(&message, message.ID).Error; err != nil {
		return nil, err
	}

	result := mapConversationMessage(message)
	return &result, nil
}

func (s *ConversationService) UpdateSettings(ctx context.Context, userID, conversationID int64, input UpdateConversationSettingsInput) (*ConversationSummary, error) {
	membership, err := s.membershipForUser(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}

	if input.IsMuted == nil {
		return nil, ErrConversationInvalidInput
	}
	membership.IsMuted = *input.IsMuted
	if err := s.db.WithContext(ctx).Model(&membership).Update("is_muted", membership.IsMuted).Error; err != nil {
		return nil, err
	}

	var conversation model.Conversation
	if err := s.db.WithContext(ctx).
		Preload("Participants.User.Profile").
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at desc, id desc").Limit(1)
		}).
		First(&conversation, conversationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrConversationNotFound
		}
		return nil, err
	}

	latestMessages, err := s.loadLatestMessages(ctx, []int64{conversationID})
	if err != nil {
		return nil, err
	}
	unreadCounts, err := s.loadUnreadCounts(ctx, userID, []int64{conversationID})
	if err != nil {
		return nil, err
	}
	summary := s.buildConversationSummary(userID, conversation, membership, latestMessages[conversationID], unreadCounts[conversationID])
	return &summary, nil
}

func (s *ConversationService) loadMemberships(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.ConversationParticipant, *int64, error) {
	var memberships []model.ConversationParticipant
	query := s.db.WithContext(ctx).
		Model(&model.ConversationParticipant{}).
		Select("conversation_participants.*").
		Joins("JOIN conversations ON conversations.id = conversation_participants.conversation_id").
		Where("user_id = ?", userID).
		Order("conversations.updated_at DESC, conversations.id DESC").
		Limit(limit + 1)
	if cursor != nil {
		query = query.Where(
			"(conversations.updated_at < (SELECT updated_at FROM conversations WHERE id = ?) OR (conversations.updated_at = (SELECT updated_at FROM conversations WHERE id = ?) AND conversations.id < ?))",
			*cursor,
			*cursor,
			*cursor,
		)
	}
	if err := query.Find(&memberships).Error; err != nil {
		return nil, nil, err
	}

	var nextCursor *int64
	if len(memberships) > limit {
		value := memberships[limit-1].ConversationID
		nextCursor = &value
		memberships = memberships[:limit]
	}
	return memberships, nextCursor, nil
}

func (s *ConversationService) membershipForUser(ctx context.Context, userID, conversationID int64) (model.ConversationParticipant, error) {
	var membership model.ConversationParticipant
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND conversation_id = ?", userID, conversationID).
		First(&membership).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var conversationCount int64
			if countErr := s.db.WithContext(ctx).
				Model(&model.Conversation{}).
				Where("id = ?", conversationID).
				Count(&conversationCount).Error; countErr != nil {
				return membership, countErr
			}
			if conversationCount == 0 {
				return membership, ErrConversationNotFound
			}
			return membership, ErrConversationForbidden
		}
		return membership, err
	}
	return membership, nil
}

func (s *ConversationService) buildConversationSummary(
	userID int64,
	conversation model.Conversation,
	membership model.ConversationParticipant,
	lastMessage *model.Message,
	unreadCount int64,
) ConversationSummary {
	summary := ConversationSummary{
		ID:          conversation.ID,
		Type:        conversation.Type,
		Title:       conversation.Title,
		UnreadCount: unreadCount,
		IsMuted:     membership.IsMuted,
	}

	if lastMessage != nil {
		summary.LastMessage = lastMessage.Content
		lastMessageAt := lastMessage.CreatedAt
		summary.LastMessageAt = &lastMessageAt
	}

	if summary.Title == "" {
		summary.Title = defaultConversationTitle(conversation.Participants, userID)
	}
	summary.AvatarURL = otherParticipantAvatar(conversation.Participants, userID)

	return summary
}

func (s *ConversationService) loadLatestMessages(ctx context.Context, conversationIDs []int64) (map[int64]*model.Message, error) {
	latest := make(map[int64]*model.Message, len(conversationIDs))
	if len(conversationIDs) == 0 {
		return latest, nil
	}

	var latestIDs []int64
	query := s.db.WithContext(ctx).
		Table("messages AS m").
		Select("m.id").
		Where("m.conversation_id IN ?", conversationIDs).
		Where("m.id = (SELECT m2.id FROM messages AS m2 WHERE m2.conversation_id = m.conversation_id ORDER BY m2.created_at DESC, m2.id DESC LIMIT 1)")
	if err := query.Find(&latestIDs).Error; err != nil {
		return nil, err
	}
	if len(latestIDs) == 0 {
		return latest, nil
	}

	var messages []model.Message
	if err := s.db.WithContext(ctx).
		Preload("Sender.Profile").
		Where("id IN ?", latestIDs).
		Find(&messages).Error; err != nil {
		return nil, err
	}
	for index := range messages {
		message := messages[index]
		latest[message.ConversationID] = &message
	}
	return latest, nil
}

func (s *ConversationService) loadUnreadCounts(ctx context.Context, userID int64, conversationIDs []int64) (map[int64]int64, error) {
	counts := make(map[int64]int64, len(conversationIDs))
	if len(conversationIDs) == 0 {
		return counts, nil
	}

	var rows []struct {
		ConversationID int64 `gorm:"column:conversation_id"`
		Count          int64 `gorm:"column:unread_count"`
	}
	if err := s.db.WithContext(ctx).
		Table("messages AS m").
		Select("m.conversation_id, COUNT(*) AS unread_count").
		Joins("JOIN conversation_participants AS cp ON cp.conversation_id = m.conversation_id AND cp.user_id = ?", userID).
		Where("m.conversation_id IN ? AND m.sender_id <> ?", conversationIDs, userID).
		Where("cp.last_read_at IS NULL OR m.created_at > cp.last_read_at").
		Group("m.conversation_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.ConversationID] = row.Count
	}
	return counts, nil
}

func mapConversationMessage(message model.Message) ConversationMessage {
	return ConversationMessage{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
		SenderName:     userDisplayName(message.Sender),
		SenderAvatar:   userAvatar(message.Sender),
		Content:        message.Content,
		MessageType:    message.MessageType,
		IsRead:         message.IsRead,
		CreatedAt:      message.CreatedAt,
	}
}

func defaultConversationTitle(participants []model.ConversationParticipant, currentUserID int64) string {
	for _, participant := range participants {
		if participant.UserID == currentUserID {
			continue
		}
		if participant.User != nil {
			return userDisplayName(participant.User)
		}
	}
	return "Conversation"
}

func otherParticipantAvatar(participants []model.ConversationParticipant, currentUserID int64) string {
	for _, participant := range participants {
		if participant.UserID == currentUserID {
			continue
		}
		if participant.User != nil {
			return userAvatar(participant.User)
		}
	}
	return ""
}

func userDisplayName(user *model.User) string {
	if user == nil {
		return "Unknown User"
	}
	if user.Profile != nil && strings.TrimSpace(user.Profile.Nickname) != "" {
		return user.Profile.Nickname
	}
	return "User"
}

func userAvatar(user *model.User) string {
	if user == nil || user.Profile == nil {
		return ""
	}
	return user.Profile.AvatarURL
}

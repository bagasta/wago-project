package service

import (
	"wago-backend/internal/model"
	"wago-backend/internal/repository"
	"wago-backend/internal/whatsapp"
)

type SessionService struct {
	SessionRepo *repository.SessionRepository
	ClientMgr   *whatsapp.ClientManager
}

func NewSessionService(sessionRepo *repository.SessionRepository, clientMgr *whatsapp.ClientManager) *SessionService {
	return &SessionService{
		SessionRepo: sessionRepo,
		ClientMgr:   clientMgr,
	}
}

func (s *SessionService) CreateSession(userID, sessionName, webhookURL string) (*model.Session, error) {
	session := &model.Session{
		UserID:      userID,
		SessionName: sessionName,
		WebhookURL:  webhookURL,
		Status:      model.SessionStatusDisconnected,
	}

	return s.SessionRepo.CreateSession(session)
}

func (s *SessionService) GetSessions(userID string) ([]*model.Session, error) {
	return s.SessionRepo.GetSessionsByUserID(userID)
}

func (s *SessionService) GetSession(id string) (*model.Session, error) {
	return s.SessionRepo.GetSessionByID(id)
}

func (s *SessionService) StartSession(id string) error {
	return s.ClientMgr.Connect(id)
}

func (s *SessionService) DeleteSession(id, userID string) error {
	// Disconnect first
	s.ClientMgr.Disconnect(id)
	return s.SessionRepo.DeleteSession(id, userID)
}

func (s *SessionService) UpdateSession(session *model.Session) error {
	return s.SessionRepo.UpdateSession(session)
}

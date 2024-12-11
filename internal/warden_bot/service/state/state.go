package state

import "time"

type BotState struct {
	UserStates map[int]time.Time
}

func NewBotState() *BotState {
	return &BotState{
		UserStates: make(map[int]time.Time),
	}
}

func (s *BotState) SetUserState(userID int, state time.Time) {
	s.UserStates[userID] = state
}

func (s *BotState) GetUserState(userID int) (time.Time, bool) {
	state, exists := s.UserStates[userID]
	return state, exists
}

func (s *BotState) ClearUserState(userID int) {
	delete(s.UserStates, userID)
}

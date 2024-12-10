package state

type BotState struct {
	UserStates map[int]string
}

func NewBotState() *BotState {
	return &BotState{
		UserStates: make(map[int]string),
	}
}

func (s *BotState) SetUserState(userID int, state string) {
	s.UserStates[userID] = state
}

func (s *BotState) GetUserState(userID int) (string, bool) {
	state, exists := s.UserStates[userID]
	return state, exists
}

func (s *BotState) ClearUserState(userID int) {
	delete(s.UserStates, userID)
}

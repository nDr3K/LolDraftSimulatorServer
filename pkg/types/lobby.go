package types

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Role represents the possible champion roles
type Role string

const (
	RoleTop     Role = "top"
	RoleJungle  Role = "jungle"
	RoleMid     Role = "mid"
	RoleBot     Role = "bot"
	RoleSupport Role = "support"
)

// DraftChampionStatus represents the status of a champion in draft
type DraftChampionStatus string

const (
	ChampStatusHover    DraftChampionStatus = "hover"
	ChampStatusSelected DraftChampionStatus = "selected"
	ChampStatusNone     DraftChampionStatus = "none"
	ChampStatusDisabled DraftChampionStatus = "disabled"
)

// DraftChampion represents a champion in the draft
type DraftChampion struct {
	ID     string              `json:"id"`
	Name   string              `json:"name"`
	Roles  []Role              `json:"role"`
	Status DraftChampionStatus `json:"status"`
}

// TeamState represents the state of a team in the draft
type TeamState struct {
	Name          string           `json:"name"`
	Picks         []*DraftChampion `json:"picks"`
	Bans          []string         `json:"bans"`
	PreviousPicks []string         `json:"previousPicks"`
	PreviousBans  []string         `json:"previousBans"`
}

// DraftPhase represents the current phase of the draft
type DraftPhase string

const (
	PhaseReady   DraftPhase = "ready"
	PhaseBan     DraftPhase = "ban"
	PhasePick    DraftPhase = "pick"
	PhaseEnd     DraftPhase = "end"
	PhaseRestart DraftPhase = "restart"
)

// DraftTurn represents whose turn it is in the draft
type DraftTurn string

const (
	TurnBlue DraftTurn = "blue"
	TurnRed  DraftTurn = "red"
	TurnEnd  DraftTurn = "end"
)

// DraftOptions represents the configurable options for the draft
type DraftOptions struct {
	IsFearless    bool `json:"isFearless"`
	BanPick       bool `json:"banPick"`
	KeepBan       bool `json:"keepBan"`
	TournamentBan bool `json:"tournamentBan"`
}

// DraftState represents the complete state of the draft
type DraftState struct {
	HasTimer bool         `json:"hasTimer"`
	Phase    DraftPhase   `json:"phase"`
	Turn     DraftTurn    `json:"turn"`
	Game     int          `json:"game"`
	Chat     []string     `json:"chat"`
	BlueTeam TeamState    `json:"blueTeam"`
	RedTeam  TeamState    `json:"redTeam"`
	Options  DraftOptions `json:"options"`
}

type LobbyRole string

const (
	RoleBlueTeam  LobbyRole = "blue"
	RoleRedTeam   LobbyRole = "red"
	RoleSpectator LobbyRole = "spectator"
)

type User struct {
	ID       string
	Conn     *websocket.Conn
	Role     LobbyRole
	Username string
}

type Lobby struct {
	ID         string
	Users      map[string]*User
	BlueTeam   map[string]*User
	RedTeam    map[string]*User
	Spectators map[string]*User
	Mutex      sync.RWMutex
	DraftState DraftState
}

func NewLobby(options DraftOptions, blueTeamName string, redTeamName string) *Lobby {
	return &Lobby{
		ID:         uuid.New().String(),
		Users:      make(map[string]*User),
		BlueTeam:   make(map[string]*User),
		RedTeam:    make(map[string]*User),
		Spectators: make(map[string]*User),
		DraftState: DraftState{
			HasTimer: true,
			Phase:    PhaseReady,
			Turn:     TurnBlue,
			Game:     1,
			Chat:     []string{},
			BlueTeam: TeamState{
				Name:          blueTeamName,
				Picks:         make([]*DraftChampion, 5),
				Bans:          make([]string, 3),
				PreviousPicks: []string{},
				PreviousBans:  []string{},
			},
			RedTeam: TeamState{
				Name:          redTeamName,
				Picks:         make([]*DraftChampion, 5),
				Bans:          make([]string, 3),
				PreviousPicks: []string{},
				PreviousBans:  []string{},
			},
			Options: options,
		},
	}
}

func generateLobbyID() string {
	return uuid.New().String()
}

func (l *Lobby) AddUser(User *User) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	l.Users[User.ID] = User
	switch User.Role {
	case RoleBlueTeam:
		l.BlueTeam[User.ID] = User
	case RoleRedTeam:
		l.RedTeam[User.ID] = User
	case RoleSpectator:
		l.Spectators[User.ID] = User
	}
}

func (l *Lobby) RemoveUser(UserID string) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	User, exists := l.Users[UserID]
	if !exists {
		return
	}

	delete(l.Users, UserID)
	switch User.Role {
	case RoleBlueTeam:
		delete(l.BlueTeam, UserID)
	case RoleRedTeam:
		delete(l.RedTeam, UserID)
	case RoleSpectator:
		delete(l.Spectators, UserID)
	}
}

func (l *Lobby) GetUsers() map[string]*User {
	l.Mutex.RLock()
	defer l.Mutex.RUnlock()

	UsersCopy := make(map[string]*User)
	for id, User := range l.Users {
		UsersCopy[id] = User
	}
	return UsersCopy
}

func (l *Lobby) GetUsersByRole(role LobbyRole) map[string]*User {
	l.Mutex.RLock()
	defer l.Mutex.RUnlock()

	var roleMap map[string]*User
	switch role {
	case RoleBlueTeam:
		roleMap = l.BlueTeam
	case RoleRedTeam:
		roleMap = l.RedTeam
	case RoleSpectator:
		roleMap = l.Spectators
	default:
		return nil
	}

	UsersCopy := make(map[string]*User)
	for id, User := range roleMap {
		UsersCopy[id] = User
	}
	return UsersCopy
}

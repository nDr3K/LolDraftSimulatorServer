package types

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Role string

const (
	RoleTop     Role = "top"
	RoleJungle  Role = "jungle"
	RoleMid     Role = "mid"
	RoleBot     Role = "bot"
	RoleSupport Role = "support"
)

type DraftChampionStatus string

const (
	ChampStatusHover    DraftChampionStatus = "hover"
	ChampStatusSelected DraftChampionStatus = "selected"
	ChampStatusNone     DraftChampionStatus = "none"
	ChampStatusDisabled DraftChampionStatus = "disabled"
)

type DraftChampion struct {
	ID     string              `json:"id"`
	Name   string              `json:"name"`
	Roles  []Role              `json:"role"`
	Status DraftChampionStatus `json:"status"`
}

type TeamState struct {
	Name          string           `json:"name"`
	Picks         []*DraftChampion `json:"picks"`
	Bans          []*string        `json:"bans"`
	PreviousPicks []string         `json:"previousPicks"`
	PreviousBans  []string         `json:"previousBans"`
}

type DraftPhase string

const (
	PhaseReady   DraftPhase = "ready"
	PhaseBan     DraftPhase = "ban"
	PhasePick    DraftPhase = "pick"
	PhaseEnd     DraftPhase = "end"
	PhaseRestart DraftPhase = "restart"
	PhaseOver    DraftPhase = "over"
)

type DraftTurn string

const (
	TurnBlue  DraftTurn = "blue"
	TurnRed   DraftTurn = "red"
	TurnStart DraftTurn = "start"
	TurnEnd   DraftTurn = "end"
)

type DraftOptions struct {
	IsFearless    bool `json:"isFearless"`
	BanPick       bool `json:"banPick"`
	KeepBan       bool `json:"keepBan"`
	TournamentBan bool `json:"tournamentBan"`
	HasTimer      bool `json:"hasTimer"`
}

type DraftState struct {
	HasTimer            bool         `json:"hasTimer"`
	Timer               int          `json:"timer"`
	Phase               DraftPhase   `json:"phase"`
	Turn                DraftTurn    `json:"turn"`
	Game                int          `json:"game"`
	Chat                []string     `json:"chat"`
	BlueTeam            TeamState    `json:"blueTeam"`
	RedTeam             TeamState    `json:"redTeam"`
	Options             DraftOptions `json:"options"`
	DisabledChampionIds []*string    `json:"disabledChampionIds"`
}

type LobbyRole string

const (
	RoleBlueTeam  LobbyRole = "blue"
	RoleRedTeam   LobbyRole = "red"
	RoleSpectator LobbyRole = "spectator"
)

type EventType string

const (
	Select  EventType = "SELECT"
	Hover   EventType = "HOVER"
	Message EventType = "MESSAGE"
	Start   EventType = "START"
	Timeout EventType = "TIMEOUT"
)

type Event struct {
	User    DraftTurn `json:"user"`
	Type    EventType `json:"type"`
	Payload Payload   `json:"payload"`
	Flag    bool      `json:"flag"`
}

type ChampionStatus string

const (
	Hovered  ChampionStatus = "hover"
	Selected ChampionStatus = "selected"
	None     ChampionStatus = "none"
	Disabled ChampionStatus = "disabled"
)

type Payload struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Role   []Role         `json:"role"`
	Status ChampionStatus `json:"status"`
}

type User struct {
	ID       string
	Conn     *websocket.Conn
	Role     LobbyRole
	Username string
}

type Lobby struct {
	ID               string
	Users            map[string]*User
	BlueTeam         map[string]*User
	RedTeam          map[string]*User
	Spectators       map[string]*User
	Mutex            sync.RWMutex
	DraftState       DraftState
	DraftService     DraftServiceInterface
	Champions        []*DraftChampion
	LastActivityTime time.Time
}

func NewLobby(options DraftOptions, blueTeamName string, redTeamName string, champions []*DraftChampion, disabledChampionIds []*string) *Lobby {
	var timer int
	if options.HasTimer {
		timer = 30
	} else {
		timer = 0
	}
	return &Lobby{
		ID:         uuid.New().String(),
		Users:      make(map[string]*User),
		BlueTeam:   make(map[string]*User),
		RedTeam:    make(map[string]*User),
		Spectators: make(map[string]*User),
		DraftState: DraftState{
			HasTimer: options.HasTimer,
			Timer:    timer,
			Phase:    PhaseReady,
			Turn:     TurnStart,
			Game:     1,
			Chat:     []string{},
			BlueTeam: TeamState{
				Name:          blueTeamName,
				Picks:         make([]*DraftChampion, 5),
				Bans:          make([]*string, 5),
				PreviousPicks: []string{},
				PreviousBans:  []string{},
			},
			RedTeam: TeamState{
				Name:          redTeamName,
				Picks:         make([]*DraftChampion, 5),
				Bans:          make([]*string, 5),
				PreviousPicks: []string{},
				PreviousBans:  []string{},
			},
			Options:             options,
			DisabledChampionIds: disabledChampionIds,
		},
		Champions:        champions,
		LastActivityTime: time.Now(),
	}
}

func (l *Lobby) AddUser(user *User) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	l.Users[user.ID] = user
	l.LastActivityTime = time.Now()

	switch user.Role {
	case RoleBlueTeam:
		l.BlueTeam[user.ID] = user
	case RoleRedTeam:
		l.RedTeam[user.ID] = user
	case RoleSpectator:
		l.Spectators[user.ID] = user
	}
}

func (l *Lobby) RemoveUser(userID string) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	if user, exists := l.Users[userID]; exists {
		delete(l.Users, userID)
		l.LastActivityTime = time.Now()

		switch user.Role {
		case RoleBlueTeam:
			delete(l.BlueTeam, userID)
		case RoleRedTeam:
			delete(l.RedTeam, userID)
		case RoleSpectator:
			delete(l.Spectators, userID)
		}
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

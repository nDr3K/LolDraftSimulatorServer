package service

import (
	"fmt"
	"log"
	"sync"
	"time"

	"fearlessdraft-server/pkg/types"
)

type DraftService struct {
	lobby        *types.Lobby
	turnCounter  int
	timerMutex   sync.Mutex
	timerStopper chan struct{}
}

func NewDraftService(lobby *types.Lobby) *DraftService {
	service := &DraftService{
		lobby:       lobby,
		turnCounter: 1,
	}

	return service
}

func (ds *DraftService) StartTimer(sendStateFunc func(*types.Lobby)) {
	ds.timerMutex.Lock()
	defer ds.timerMutex.Unlock()

	if ds.timerStopper != nil {
		close(ds.timerStopper)
	}

	if !ds.lobby.DraftState.HasTimer || ds.lobby.DraftState.Timer <= 0 {
		return
	}

	ds.timerStopper = make(chan struct{})
	stopper := ds.timerStopper

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ds.timerMutex.Lock()
				if ds.lobby.DraftState.Timer > 0 {
					ds.lobby.DraftState.Timer--
					sendStateFunc(ds.lobby)
				}

				// Stop timer if it reaches 0
				if ds.lobby.DraftState.Timer <= 0 {
					ds.timerMutex.Unlock()
					return
				}
				ds.timerMutex.Unlock()

			case <-stopper:
				return
			}
		}
	}()
}

func (ds *DraftService) StopTimer() {
	ds.timerMutex.Lock()
	defer ds.timerMutex.Unlock()

	if ds.timerStopper != nil {
		close(ds.timerStopper)
		ds.timerStopper = nil
	}
}

func (ds *DraftService) HandleEvent(event *types.Event, sendStateFunc func(*types.Lobby)) (bool, error) {
	if ds.lobby.DraftState.Turn != event.User &&
		ds.lobby.DraftState.Turn != types.TurnStart &&
		ds.lobby.DraftState.Turn != types.TurnEnd {
		return false, nil
	}

	if event.Type == types.Select {
		ds.StopTimer()
	}

	switch event.Type {
	case types.Start:
		return ds.handleStartEvent(event, sendStateFunc)
	case types.Hover:
		if ds.lobby.DraftState.Phase == types.PhasePick {
			return ds.handleHoverEvent(event)
		}
		return false, nil
	case types.Select:
		return ds.handleSelectEvent(event, sendStateFunc)
	case types.Timeout:
		// TODO
		ds.lobby.DraftState.Timer = 30
		sendStateFunc(ds.lobby)
		return true, nil
	case types.Message:
		// Currently not implemented
		return true, nil
	default:
		return false, fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func (ds *DraftService) handleStartEvent(event *types.Event, sendStateFunc func(*types.Lobby)) (bool, error) {
	switch ds.lobby.DraftState.Phase {
	case types.PhaseReady:
		ds.handleWaitingConfirm(event, types.TurnStart, func() {
			ds.lobby.DraftState.Turn = types.TurnBlue
			ds.lobby.DraftState.Phase = types.PhaseBan
			ds.lobby.DraftState.Timer = 30
			sendStateFunc(ds.lobby)
			ds.StartTimer(sendStateFunc)
		})
	case types.PhaseEnd:
		if event.User == types.TurnBlue {
			ds.lobby.DraftState.Turn = types.TurnRed
		} else if event.User == types.TurnRed {
			ds.lobby.DraftState.Turn = types.TurnBlue
		}
		if ds.lobby.DraftState.Game < 5 {
			ds.lobby.DraftState.Phase = types.PhaseRestart
		} else {
			ds.lobby.DraftState.Phase = types.PhaseOver
		}
	case types.PhaseRestart:
		if ds.lobby.DraftState.Game < 5 {
			ds.handleRestart(event.Flag)
		}
	}

	return true, nil
}

func (ds *DraftService) handleWaitingConfirm(event *types.Event, turn types.DraftTurn, action func()) {
	if ds.lobby.DraftState.Turn == turn {
		if event.User == types.TurnBlue {
			ds.lobby.DraftState.Turn = types.TurnRed
		} else if event.User == types.TurnRed {
			ds.lobby.DraftState.Turn = types.TurnBlue
		}
	} else {
		action()
	}
}

func (ds *DraftService) handleRestart(switchSide bool) {
	blueSide := ds.lobby.DraftState.BlueTeam
	redSide := ds.lobby.DraftState.RedTeam

	if switchSide {
		blueSide, redSide = redSide, blueSide
	}

	if ds.lobby.DraftState.Options.IsFearless {
		blueSide.PreviousPicks = append(blueSide.PreviousPicks, ds.extractPreviousPicks(blueSide.Picks)...)
		redSide.PreviousPicks = append(redSide.PreviousPicks, ds.extractPreviousPicks(redSide.Picks)...)

		if ds.lobby.DraftState.Options.KeepBan {
			blueSide.PreviousBans = append(blueSide.PreviousBans, ds.extractPreviousBans(blueSide.Bans)...)
			redSide.PreviousBans = append(redSide.PreviousBans, ds.extractPreviousBans(redSide.Bans)...)
		}
	}

	// Reset picks and bans
	blueSide.Picks = make([]*types.DraftChampion, 5)
	redSide.Picks = make([]*types.DraftChampion, 5)
	blueSide.Bans = make([]*string, 5)
	redSide.Bans = make([]*string, 5)

	ds.turnCounter = 1
	ds.lobby.DraftState.Phase = types.PhaseReady
	ds.lobby.DraftState.Game++
	ds.lobby.DraftState.Turn = types.TurnStart
	ds.lobby.DraftState.BlueTeam = blueSide
	ds.lobby.DraftState.RedTeam = redSide
}

func (ds *DraftService) extractPreviousPicks(picks []*types.DraftChampion) []string {
	previousPicks := []string{}
	for _, pick := range picks {
		if pick != nil {
			previousPicks = append(previousPicks, pick.ID)
		} else {
			previousPicks = append(previousPicks, "none")
		}
	}
	return previousPicks
}

func (ds *DraftService) extractPreviousBans(bans []*string) []string {
	previousBans := []string{}
	for _, ban := range bans {
		if ban != nil {
			previousBans = append(previousBans, *ban)
		} else {
			previousBans = append(previousBans, "none")
		}
	}
	return previousBans
}

func (ds *DraftService) handleHoverEvent(event *types.Event) (bool, error) {

	teamKey := ds.determineTeamKey(event.User)

	team := ds.getTeamState(teamKey)

	hoverChampion := &types.DraftChampion{
		ID:     event.Payload.ID,
		Name:   event.Payload.Name,
		Roles:  event.Payload.Role,
		Status: types.ChampStatusHover,
	}
	updated := ds.updateChampionArray(team.Picks, hoverChampion)

	if !updated {
		log.Println("Unable to hover the champion")
	}

	return true, nil
}

func (ds *DraftService) handleSelectEvent(event *types.Event, sendStateFunc func(*types.Lobby)) (bool, error) {

	teamKey := ds.determineTeamKey(event.User)
	team := ds.getTeamState(teamKey)

	isBanPhase := ds.lobby.DraftState.Phase == types.PhaseBan

	var updated bool
	if isBanPhase {
		updated = ds.updateStringArray(team.Bans, &event.Payload.ID)
	} else {
		selectedChampion := &types.DraftChampion{
			ID:     event.Payload.ID,
			Name:   event.Payload.Name,
			Roles:  event.Payload.Role,
			Status: types.ChampStatusSelected,
		}

		updated = ds.updateChampionArray(team.Picks, selectedChampion)
	}

	if !updated {
		log.Printf("Unable to %s the champion", map[bool]string{true: "ban", false: "select"}[isBanPhase])
		return true, nil
	}

	ds.turnCounter++
	ds.updatePhaseAndTurn()

	ds.lobby.DraftState.Timer = 30
	sendStateFunc(ds.lobby)
	ds.StartTimer(sendStateFunc)

	return true, nil
}

func (ds *DraftService) determineTeamKey(turn types.DraftTurn) string {
	if turn == types.TurnBlue {
		return "blueTeam"
	}
	return "redTeam"
}

func (ds *DraftService) getTeamState(teamKey string) *types.TeamState {
	if teamKey == "blueTeam" {
		return &ds.lobby.DraftState.BlueTeam
	}
	return &ds.lobby.DraftState.RedTeam
}

func (ds *DraftService) updateChampionArray(arr []*types.DraftChampion, value *types.DraftChampion) bool {
	for i := range arr {
		if arr[i] != nil && arr[i].Status == types.ChampStatusHover {
			arr[i] = value
			return true
		}
	}

	for i := range arr {
		if arr[i] == nil {
			arr[i] = value
			return true
		}
	}

	return false
}

func (ds *DraftService) updateStringArray(arr []*string, value *string) bool {
	for i := range arr {
		if arr[i] == nil {
			arr[i] = value
			return true
		}
	}
	return false
}

func (ds *DraftService) updatePhaseAndTurn() {
	if ds.lobby.DraftState.Options.TournamentBan {
		ds.lobby.DraftState.Phase = ds.determinePhase(ds.turnCounter)
		ds.lobby.DraftState.Turn = ds.getTurn(ds.turnCounter)
	} else {
		ds.lobby.DraftState.Phase = ds.determineStandardPhase(ds.turnCounter)
		ds.lobby.DraftState.Turn = ds.getStandardTurn(ds.turnCounter)
	}
}

func (ds *DraftService) determineStandardPhase(turnCounter int) types.DraftPhase {
	if turnCounter <= 10 {
		return types.PhaseBan
	}
	if turnCounter <= 20 {
		return types.PhasePick
	}
	return types.PhaseEnd
}

func (ds *DraftService) determinePhase(turnCounter int) types.DraftPhase {
	if turnCounter <= 6 {
		return types.PhaseBan
	}
	if turnCounter <= 12 {
		return types.PhasePick
	}
	if turnCounter <= 16 {
		return types.PhaseBan
	}
	if turnCounter <= 20 {
		return types.PhasePick
	}
	return types.PhaseEnd
}

func (ds *DraftService) getStandardTurn(turnCounter int) types.DraftTurn {
	switch turnCounter {
	//bans
	case 1, 3, 5, 7, 9:
		return types.TurnBlue
	case 2, 4, 6, 8, 10:
		return types.TurnRed
		//picks
	case 11, 14, 15, 18, 19:
		return types.TurnBlue
	case 12, 13, 16, 17, 20:
		return types.TurnRed
		//end
	case 21:
		return types.TurnEnd
	default:
		panic("Invalid turn counter")
	}
}

func (ds *DraftService) getTurn(turnCounter int) types.DraftTurn {
	switch turnCounter {
	//bans
	case 1, 3, 5:
		return types.TurnBlue
	case 2, 4, 6:
		return types.TurnRed
		//picks
	case 7, 10, 11:
		return types.TurnBlue
	case 8, 9, 12:
		return types.TurnRed
		//bans
	case 13, 15:
		return types.TurnRed
	case 14, 16:
		return types.TurnBlue
		//picks
	case 17, 20:
		return types.TurnRed
	case 18, 19:
		return types.TurnBlue
		//end
	case 21:
		return types.TurnEnd
	default:
		panic("Invalid turn counter")
	}
}

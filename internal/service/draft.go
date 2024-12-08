package service

import (
	"fmt"
	"log"

	"fearlessdraft-server/pkg/types"
)

type DraftService struct {
	draftState  *types.DraftState
	turnCounter int
}

func NewDraftService(initialDraftState *types.DraftState) *DraftService {
	service := &DraftService{
		draftState: initialDraftState,
	}

	return service
}

func (ds *DraftService) HandleEvent(event *types.Event) (bool, error) {
	switch event.Type {
	case types.Start:
		return ds.handleStartEvent(event)
	case types.Hover:
		return ds.handleHoverEvent(event)
	case types.Select:
		return ds.handleSelectEvent(event)
	case types.Timeout:
		// TODO
		return true, nil
	case types.Message:
		// Currently not implemented
		return true, nil
	default:
		return false, fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func (ds *DraftService) handleStartEvent(event *types.Event) (bool, error) {
	switch ds.draftState.Phase {
	case types.PhaseEnd:
		ds.draftState.Phase = types.PhaseRestart
	case types.PhaseRestart:
		if ds.draftState.Game < 5 {
			switchSide, ok := event.Payload.(bool)
			if !ok {
				return false, fmt.Errorf("invalid payload for start event")
			}
			ds.handleRestart(switchSide)
		}
	}

	return true, nil
}

func (ds *DraftService) handleRestart(switchSide bool) {
	blueSide := ds.draftState.BlueTeam
	redSide := ds.draftState.RedTeam

	if switchSide {
		blueSide, redSide = redSide, blueSide
	}

	if ds.draftState.Options.IsFearless {
		blueSide.PreviousPicks = append(blueSide.PreviousPicks, ds.extractPreviousPicks(blueSide.Picks)...)
		redSide.PreviousPicks = append(redSide.PreviousPicks, ds.extractPreviousPicks(redSide.Picks)...)

		if ds.draftState.Options.KeepBan {
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
	ds.draftState.Phase = types.PhaseBan
	ds.draftState.Game++
	ds.draftState.Turn = types.TurnBlue
	ds.draftState.BlueTeam = blueSide
	ds.draftState.RedTeam = redSide
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
	payload, ok := event.Payload.(types.Payload)
	if !ok {
		return false, fmt.Errorf("invalid payload for hover event")
	}

	teamKey := ds.determineTeamKey(event.User)

	hoverChampion := &types.DraftChampion{
		ID:     payload.ID,
		Name:   payload.Name,
		Roles:  payload.Role,
		Status: types.ChampStatusHover,
	}

	team := ds.getTeamState(teamKey)

	updated := ds.updateChampionArray(team.Picks, hoverChampion,
		func(item *types.DraftChampion) bool {
			return item != nil && item.Status == types.ChampStatusHover
		}) ||
		ds.updateChampionArray(team.Picks, hoverChampion,
			func(item *types.DraftChampion) bool {
				return item == nil
			})

	if !updated {
		log.Println("Unable to hover the champion")
	}

	return true, nil
}

func (ds *DraftService) handleSelectEvent(event *types.Event) (bool, error) {
	payload, ok := event.Payload.(types.Payload)
	if !ok {
		return false, fmt.Errorf("invalid payload for select event")
	}

	teamKey := ds.determineTeamKey(event.User)
	team := ds.getTeamState(teamKey)

	isBanPhase := ds.draftState.Phase == types.PhaseBan

	var updated bool
	if isBanPhase {
		updated = ds.updateStringArray(team.Bans, payload.ID,
			func(item string) bool {
				return item == ""
			})
	} else {
		selectedChampion := &types.DraftChampion{
			ID:     payload.ID,
			Name:   payload.Name,
			Roles:  payload.Role,
			Status: types.ChampStatusSelected,
		}

		updated = ds.updateChampionArray(team.Picks, selectedChampion,
			func(item *types.DraftChampion) bool {
				return item != nil && item.Status == types.ChampStatusHover
			})
	}

	if !updated {
		log.Printf("Unable to %s the champion", map[bool]string{true: "ban", false: "select"}[isBanPhase])
		return true, nil
	}

	ds.turnCounter++
	ds.updatePhaseAndTurn()
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
		return &ds.draftState.BlueTeam
	}
	return &ds.draftState.RedTeam
}

func (ds *DraftService) updateChampionArray(arr []*types.DraftChampion, value *types.DraftChampion, condition func(*types.DraftChampion) bool) bool {
	for i := range arr {
		if condition(arr[i]) {
			arr[i] = value
			return true
		}
	}
	return false
}

func (ds *DraftService) updateStringArray(arr []*string, value string, condition func(string) bool) bool {
	for i := range arr {
		if condition(*arr[i]) {
			*arr[i] = value
			return true
		}
	}
	return false
}

func (ds *DraftService) updatePhaseAndTurn() {
	if ds.draftState.Options.TournamentBan {
		ds.draftState.Phase = ds.determinePhase(ds.turnCounter)
		ds.draftState.Turn = ds.getTurn(ds.turnCounter)
	} else {
		ds.draftState.Phase = ds.determineStandardPhase(ds.turnCounter)
		ds.draftState.Turn = ds.getStandardTurn(ds.turnCounter)
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
	case 1, 3, 5, 7, 9:
		return types.TurnBlue
	case 2, 4, 6, 8, 10:
		return types.TurnRed
	case 11, 14, 18, 19:
		return types.TurnBlue
	case 12, 13, 16, 17, 20:
		return types.TurnRed
	case 21:
		return types.TurnEnd
	default:
		panic("Invalid turn counter")
	}
}

func (ds *DraftService) getTurn(turnCounter int) types.DraftTurn {
	switch turnCounter {
	case 1, 3, 5:
		return types.TurnBlue
	case 2, 4, 6:
		return types.TurnRed
	case 7, 10, 11:
		return types.TurnBlue
	case 8, 9, 12:
		return types.TurnRed
	case 13, 15:
		return types.TurnRed
	case 14, 16:
		return types.TurnBlue
	case 17, 19:
		return types.TurnRed
	case 18, 20:
		return types.TurnBlue
	case 21:
		return types.TurnEnd
	default:
		panic("Invalid turn counter")
	}
}

func (ds *DraftService) Disconnect() {
	ds.turnCounter = 0
}

package blockerstest

import (
	"testing"
	"time"

	"github.com/0xor1/tlbx/cmd/games/pkg/blockers"
	"github.com/0xor1/tlbx/cmd/games/pkg/blockers/blockerseps"
	"github.com/0xor1/tlbx/cmd/games/pkg/config"
	"github.com/0xor1/tlbx/cmd/games/pkg/game"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/logrusorgru/aurora"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	a := assert.New(t)
	r := test.NewRig(
		config.Get(),
		append(game.Eps, blockerseps.Eps...),
		false,
		me.AuthedExists,
		me.SetAuthed,
		me.GetAuthed,
		me.Del,
		nil,
		nil,
		nil,
		nil,
		false,
		ratelimit.MeMware)
	defer r.CleanUp()

	g2 := playGame(a, []*app.Client{
		r.Ali().Client(),
		r.Bob().Client(),
	})
	g3 := playGame(a, []*app.Client{
		r.Ali().Client(),
		r.Bob().Client(),
		r.Cat().Client(),
	})
	g4 := playGame(a, []*app.Client{
		r.Ali().Client(),
		r.Bob().Client(),
		r.Cat().Client(),
		r.Dan().Client(),
	})
	printGame(g2)
	printGame(g3)
	printGame(g4)

	// test abondon
	// p1 creates a new game
	g := (&blockers.New{}).
		MustDo(r.Ali().Client())
	a.NotNil(g)

	// p1 abandons the game
	(&blockers.Abandon{}).
		MustDo(r.Ali().Client())

	g = (&blockers.Get{
		Game: g.ID,
	}).MustDo(r.Ali().Client())
	a.False(g.Finished())
	a.True(g.Abandoned())

	game.DeleteOutdated(func(query string, args ...interface{}) {
		r.Data().Primary().Exec(query, args...)
	}, 0, time.Hour)
}

func playGame(a *assert.Assertions, players []*app.Client) *blockers.Game {
	var err error
	var id ID
	player := func(wrong ...bool) *app.Client {
		fns := make([]func(), len(players))
		gs := make([]*blockers.Game, len(players))
		for i := range players {
			// closure on i
			ci := i
			fns[ci] = func() {
				gs[ci] = (&blockers.Get{Game: id}).
					MustDo(players[ci])
			}
		}
		PanicOn(GoGroup(fns...))
		var p *app.Client
		for i := range gs {
			if gs[i].IsMyTurn() == (len(wrong) == 0) {
				p = players[i]
				break
			}
		}
		return p
	}

	// test no active game
	active := (&game.Active{}).
		MustDo(players[0])
	a.Nil(active)

	// p1 creates a new game
	g := (&blockers.New{}).
		MustDo(players[0])
	a.NotNil(g)
	id = g.ID

	// test active game
	active = (&game.Active{}).
		MustDo(players[0])
	a.Equal("blockers", active.Type)
	a.Equal(id, active.ID)

	// p1 fails to create another new game
	g, err = (&blockers.New{}).
		Do(players[0])
	a.Nil(g)
	a.Regexp("can not create a new game while you are still participating in an active game, id: ", err)

	// p1 fails to starts the game
	g, err = (&blockers.Start{}).
		Do(players[0])
	a.Nil(g)
	a.Regexp("game hasn't met minimum player count requirement: 2", err)

	// p2-4 joins p1s game
	for _, p := range players[1:] {
		g = (&blockers.Join{Game: id}).
			MustDo(p)
		a.NotNil(g)
	}

	// p2 fails to starts the game
	g, err = (&blockers.Start{}).
		Do(players[1])
	a.Nil(g)
	a.Regexp("only the creator can start the game", err)

	// p1 starts the game
	g = (&blockers.Start{
		RandomizePlayerOrder: true,
	}).MustDo(players[0])
	a.NotNil(g)

	// p1 fails to starts the game again
	g, err = (&blockers.Start{}).
		Do(players[0])
	a.Nil(g)
	a.Regexp("can't start a game that has already been started", err)

	// invalid - missing first corner cell
	g, err = (&blockers.TakeTurn{
		Piece:    0,
		Position: 1,
	}).Do(player())
	a.Nil(g)
	a.Regexp("first corner constraint not met", err)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    0,
		Position: 0,
	}).MustDo(player())
	a.NotNil(g)

	// valid get with no change
	g = (&blockers.Get{
		Game:         g.ID,
		UpdatedAfter: &g.UpdatedOn,
	}).MustDo(player())
	a.Nil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    16,
		Position: 17,
		Rotation: 1,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    10,
		Position: 377,
		Flip:     1,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    15,
		Position: 360,
		Flip:     1,
	}).MustDo(player())
	a.NotNil(g)

	// invalid - reuse an already placed piece
	g, err = (&blockers.TakeTurn{
		Piece:    0,
		Position: 21,
	}).Do(player())
	a.Nil(g)
	a.Regexp("invalid piece, that piece has already been used", err)

	// invalid - place outside the board boundaries
	g, err = (&blockers.TakeTurn{
		Piece:    1,
		Position: 19,
	}).Do(player())
	a.Nil(g)
	a.Regexp("piece/position/rotation combination is not contained on the board", err)

	// invalid - place on top of existing piece
	g, err = (&blockers.TakeTurn{
		Piece:    1,
		Position: 0,
	}).Do(player())
	a.Nil(g)
	a.Regexp("cell [0-9]+ already occupied", err)

	// invalid - faces touching
	g, err = (&blockers.TakeTurn{
		Piece:    1,
		Position: 20,
	}).Do(player())
	a.Nil(g)
	a.Regexp("face to face constraint not met, cell 0", err)

	// invalid - no touching diagonals
	g, err = (&blockers.TakeTurn{
		Piece:    1,
		Position: 22,
	}).Do(player())
	a.Nil(g)
	a.Regexp("corner touch constraint not met", err)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    1,
		Position: 21,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    20,
		Position: 14,
	}).MustDo(player())
	a.NotNil(g)

	// invalid - not their turn
	g, err = (&blockers.TakeTurn{
		Piece:    20,
		Position: 14,
	}).Do(player(false))
	a.Nil(g)
	a.Regexp("it's not your turn", err)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    20,
		Position: 354,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    19,
		Position: 300,
		Rotation: 3,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    17,
		Position: 3,
		Rotation: 1,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    19,
		Position: 77,
		Rotation: 1,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    19,
		Position: 316,
		Flip:     1,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    20,
		Position: 240,
		Rotation: 1,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    20,
		Position: 40,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    18,
		Position: 73,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    18,
		Position: 293,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    18,
		Position: 303,
		Flip:     1,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    19,
		Position: 103,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    17,
		Position: 134,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    17,
		Position: 250,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    17,
		Position: 203,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    18,
		Position: 166,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    15,
		Position: 129,
		Rotation: 2,
		Flip:     1,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    16,
		Position: 188,
		Rotation: 1,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    7,
		Position: 206,
		Rotation: 3,
		Flip:     1,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    16,
		Position: 121,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    14,
		Position: 176,
		Rotation: 3,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    15,
		Position: 108,
		Rotation: 3,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    3,
		Position: 248,
		Rotation: 3,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    7,
		Position: 181,
		Rotation: 1,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    4,
		Position: 256,
		Rotation: 1,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    11,
		Position: 151,
		Rotation: 2,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    6,
		Position: 270,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    6,
		Position: 223,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    12,
		Position: 334,
		Rotation: 1,
		Flip:     1,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    14,
		Position: 233,
		Flip:     1,
		Rotation: 2,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    10,
		Position: 273,
		Flip:     1,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    5,
		Position: 266,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    10,
		Position: 337,
		Rotation: 2,
	}).MustDo(player())
	a.NotNil(g)

	// end
	g = (&blockers.TakeTurn{
		End: 1,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    8,
		Position: 210,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    2,
		Position: 106,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    7,
		Position: 258,
		Rotation: 3,
	}).MustDo(player())
	a.NotNil(g)

	// pieceSet 1 ended so should
	// skip straight to pieceSet 2.
	// valid
	g = (&blockers.TakeTurn{
		Piece:    16,
		Position: 173,
		Rotation: 1,
	}).MustDo(player())
	a.NotNil(g)

	// end
	g = (&blockers.TakeTurn{
		End: 1,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    1,
		Position: 158,
	}).MustDo(player())
	a.NotNil(g)

	// valid
	g = (&blockers.TakeTurn{
		Piece:    0,
		Position: 182,
	}).MustDo(player())
	a.NotNil(g)

	// end
	g = (&blockers.TakeTurn{
		End: 1,
	}).MustDo(player())
	a.NotNil(g)

	turn := uint32(50)
	if len(players) != 3 {
		// end last piece idx
		g = (&blockers.TakeTurn{
			End: 1,
		}).MustDo(player())
		a.NotNil(g)
		turn = 52
	}

	// get
	g = (&blockers.Get{
		Game: g.ID,
	}).MustDo(players[0])
	a.NotNil(g)

	a.Equal(uint8(2), g.State)
	a.Equal(turn, g.Turn)

	return g
}

func printGame(g *blockers.Game) {
	Println()
	Println("turn", g.Turn, "state", g.State, "ended", g.PieceSetsEnded)
	Println()
	printPieceSets(g)
	Println()
	printBoard(g)
	Println()
}

func printPieceSets(g *blockers.Game) {
	for j := uint8(0); j < 4; j++ {
		row := ""
		for i := uint8(0); i < blockers.PiecesCount(); i++ {
			color := aurora.White
			switch j {
			case 0:
				color = aurora.Red
			case 1:
				color = aurora.Green
			case 2:
				color = aurora.Blue
			case 3:
				color = aurora.Yellow
			}
			if g.PieceSets[blockers.PiecesCount()*j+i] == 1 {
				row += aurora.Sprintf(color(`■ `))
			} else {
				row += aurora.Sprintf(color(`□ `))
			}
		}
		Println(row)
	}
}

func printBoard(g *blockers.Game) {
	for y := 0; y < 20; y++ {
		rowStart := 20 * y
		row := "%d "
		if rowStart < 100 {
			row += " "
		}
		if rowStart < 10 {
			row += " "
		}
		row = Strf(row, rowStart)
		for x := 0; x < 20; x++ {
			color := aurora.White
			switch g.Board[20*y+x] {
			case 0:
				color = aurora.Red
			case 1:
				color = aurora.Green
			case 2:
				color = aurora.Blue
			case 3:
				color = aurora.Yellow
			}
			row += aurora.Sprintf(color(`■ `))
		}
		row += Strf(" %d", 20*(y+1)-1)
		Println(row)
	}
}

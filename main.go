package main

import (
	"fmt"
	"strconv"
)

func main() {
	fmt.Println("Shahio 0.1")
}

const (
	White         = 'w'
	Black         = 'b'
	KingCastling  = "O-O"
	QueenCastling = "O-O-O"
	Capture       = 'x'
	Movement      = ' '
	Promotion     = '='
	AsciiOffset   = 97
)

type Move string // e.g. "Pwa2xPbb3" White pawn at a2 captures Black pawn at b3
type Piece string
type Board [][]Piece
type Game struct {
	Board     Board
	Moves     []Move
	whoseTurn byte
	checked   bool
	ended     bool
}

func NewGame() Game {
	return Game{
		Board: [][]Piece{
			{"Rb", "Nb", "Bb", "Qb", "Kb", "Bb", "Nb", "Rb"},
			{"Pb", "Pb", "Pb", "Pb", "Pb", "Pb", "Pb", "Pb"},
			{"", "", "", "", "", "", "", ""},
			{"", "", "", "", "", "", "", ""},
			{"", "", "", "", "", "", "", ""},
			{"", "", "", "", "", "", "", ""},
			{"Pw", "Pw", "Pw", "Pw", "Pw", "Pw", "Pw", "Pw"},
			{"Rw", "Nw", "Bw", "Qw", "Kw", "Bw", "Nw", "Rw"},
		},
		Moves:     []Move{},
		whoseTurn: White,
		checked:   false,
		ended:     false,
	}
}

func (g *Game) processMove(m Move) error {
	if g.ended {
		return fmt.Errorf("game has ended")
	}

	if !checkMoveFormat(m) {
		return fmt.Errorf("invalid move format")
	}

	if g.whoseTurn != m[1] {
		return fmt.Errorf("not your turn")
	}

	actionProcessor := g.getProcessorForMove(m)
	err := actionProcessor(m)
	if err != nil {
		return err
	}

	// TODO Check if your king is checked

	// TODO Check if enemy king is checked and if game ended
	return nil
}

func (g *Game) processCastling(m Move) error {
	// Check that the king didn't move
	for _, prevm := range g.Moves {
		if prevm[0] == 'K' && prevm[1] == g.whoseTurn {
			return fmt.Errorf("king not in position")
		}
	}

	// Set rooks initial position and direction
	var RookInitialPosition string
	var RookDirection int
	if m == KingCastling {
		RookDirection = 1
		if g.whoseTurn == Black {
			RookInitialPosition = "h8"
		} else {
			RookInitialPosition = "h1"
		}
	} else {
		RookDirection = -1
		if g.whoseTurn == Black {
			RookInitialPosition = "a8"
		} else {
			RookInitialPosition = "a1"
		}
	}

	// Check that the rook didn't move
	for _, prevm := range g.Moves {
		if prevm[0] == 'R' && prevm[1] == g.whoseTurn && prevm[2:4] == Move(RookInitialPosition) {
			return fmt.Errorf("rook not in position")
		}
		if prevm[4] == Capture && prevm[5] == 'R' && prevm[6] == g.whoseTurn && prevm[7:9] == Move(RookInitialPosition) {
			return fmt.Errorf("rook not in position")
		}
	}

	// Check if there are pieces between the king and the rook
	x := 4
	y, _ := strconv.Atoi(RookInitialPosition[1:])
	for {
		x += RookDirection

		pos := fmt.Sprintf("%c%d", x+AsciiOffset, y)
		if pos == RookInitialPosition {
			break
		}

		if g.Board[y][x] != "" {
			return fmt.Errorf("pieces between king and rook")
		}
	}

	// Set king and rook new positions
	g.Board[y][x+RookDirection] = Piece(fmt.Sprintf("R%c", g.whoseTurn))
	g.Board[y][x+2*RookDirection] = Piece(fmt.Sprintf("K%c", g.whoseTurn))
	// Clear old positions
	g.Board[y][x] = Piece("")
	g.Board[y][4] = Piece("")

	return nil
}

func (g *Game) getProcessorForMove(m Move) func(Move) error {
	// Castling
	if m == KingCastling || m == QueenCastling {
		return g.processCastling
	}

	// TODO Add other processors
	return nil
}

func checkMoveFormat(m Move) bool {
	// Castling
	if m == KingCastling || m == QueenCastling {
		return true
	}

	if len(m) != 9 {
		return false
	}

	// 'x' means capture; ' ' means move; '=' means promotion
	// Check if move action is correct
	if m[4] != 'x' && m[4] != ' ' && m[4] != '=' {
		return false
	}

	// Check if move notation is correct
	if (m[2] < 'a' || m[2] > 'h' || m[3] < '1' || m[3] > '8') ||
		(m[5] < 'a' || m[5] > 'h' || m[6] < '1' || m[6] > '8') {
		return false
	}

	return true
}

package main

import (
	"fmt"
)

func main() {
	fmt.Println("Shahio 0.1")
}

type Move string // e.g. "Pwa2xPbb3" White pawn at a2 captures Black pawn at b3е
type Piece string
type Position string
type Side byte
type Board [][]Piece
type Game struct {
	Board     Board
	Moves     []Move
	whoseTurn Side
	ended     bool
}

const (
	White         = Side('w')
	Black         = Side('b')
	KingCastling  = "O-O"
	QueenCastling = "O-O-O"
	Capture       = 'x'
	Movement      = ' '
	Promotion     = '='
	AsciiOffset   = 97
)

func NewGame() Game {
	return Game{
		Board: [][]Piece{
			{"Pw", "Pw", "Pw", "Pw", "Pw", "Pw", "Pw", "Pw"},
			{"Rw", "Nw", "Bw", "Qw", "Kw", "Bw", "Nw", "Rw"},
			{"", "", "", "", "", "", "", ""},
			{"", "", "", "", "", "", "", ""},
			{"", "", "", "", "", "", "", ""},
			{"", "", "", "", "", "", "", ""},
			{"Rb", "Nb", "Bb", "Qb", "Kb", "Bb", "Nb", "Rb"},
			{"Pb", "Pb", "Pb", "Pb", "Pb", "Pb", "Pb", "Pb"},
		},
		Moves:     []Move{},
		whoseTurn: White,
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

	if g.whoseTurn != Side(m[1]) {
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

func (g *Game) getProcessorForMove(m Move) func(Move) error {
	// Castling
	if m == KingCastling || m == QueenCastling {
		return g.processCastling
	}

	// TODO Add other processors
	return nil
}

func (g *Game) processCastling(m Move) error {
	// Check that the king didn't move
	for _, prevm := range g.Moves {
		if prevm[0] == 'K' && Side(prevm[1]) == g.whoseTurn {
			return fmt.Errorf("king not in position")
		}
	}

	var (
		kingRow  int
		kingCol  int = 4
		attacker Side
	)

	if g.whoseTurn == Black {
		kingRow = 8
		attacker = White
	} else {
		kingRow = 1
		attacker = Black
	}

	var (
		rookInitialPosition Position
		rookDir             int
	)
	if m == KingCastling {
		rookDir = 1
		rookInitialPosition = Position(fmt.Sprintf("h%d", kingRow))
	} else {
		rookDir = -1
		rookInitialPosition = Position(fmt.Sprintf("a%d", kingRow))
	}

	// Check that the rook didn't move
	for _, prevm := range g.Moves {
		if prevm[0] == 'R' && Side(prevm[1]) == g.whoseTurn &&
			Position(prevm[2:4]) == rookInitialPosition {
			return fmt.Errorf("rook not in position")
		}
		if prevm[4] == Capture &&
			prevm[5] == 'R' && Side(prevm[6]) == g.whoseTurn &&
			Position(prevm[7:9]) == rookInitialPosition {
			return fmt.Errorf("rook not in position")
		}
	}

	// Check if there are pieces between the king and the rook
	column := kingCol
	for {
		column += rookDir

		if Position(fmt.Sprintf("%c%d", column+AsciiOffset, kingRow)) == rookInitialPosition {
			break
		}

		if g.Board[kingRow][column] != "" {
			return fmt.Errorf("pieces between king and rook")
		}
	}

	// Check if crossover squares are attacked
	crossoverAttacked := len(g.cellAttackers(kingCol+rookDir, kingRow, attacker)) > 0
	if crossoverAttacked {
		return fmt.Errorf("crossover cells attacked")
	}

	// Check if king is checked
	kingChecked := len(g.cellAttackers(kingCol, kingRow, attacker)) > 0
	if kingChecked {
		return fmt.Errorf("king is checked")
	}

	// Set king and rook new positions
	g.Board[kingRow][kingCol+rookDir] = Piece(fmt.Sprintf("R%c", g.whoseTurn))
	g.Board[kingRow][kingCol+2*rookDir] = Piece(fmt.Sprintf("K%c", g.whoseTurn))

	// Clear old positions
	g.Board[kingRow][column] = Piece("")
	g.Board[kingRow][kingCol] = Piece("")

	return nil
}

func (g *Game) cellAttackers(cellCol, cellRow int, attacker Side) []Position {
	var res []Position

	isValidPosition := func(col, row int) bool {
		return col > -1 && col < 8 && row > -1 && row < 8
	}

	checkDirections := func(dirs [4][2]int, isAttacker func(Piece) bool) []Position {
		var resDirs []Position
		for _, dir := range dirs {
			for col, row := cellCol+dir[0], cellRow+dir[1]; isValidPosition(col, row); col, row = col+dir[0], row+dir[1] {
				if p := g.Board[row][col]; p != "" {
					if isAttacker(p) {
						resDirs = append(resDirs, Position(fmt.Sprintf("%c%d", col+AsciiOffset, row+1)))
					}
					break
				}
			}
		}

		return resDirs
	}

	// Check lines
	isFullLineAttacker := func(p Piece) bool {
		return p == Piece("R"+string(attacker)) ||
			p == Piece("Q"+string(attacker))
	}
	lineDirs := [4][2]int{
		{-1, 0}, {1, 0}, {0, -1}, {0, 1},
	}
	res = append(res, checkDirections(lineDirs, isFullLineAttacker)...)

	// Check diagonals
	isFullDiagonalAttacker := func(p Piece) bool {
		return p == Piece("B"+string(attacker)) ||
			p == Piece("Q"+string(attacker))
	}
	diagDirs := [4][2]int{
		{-1, 1}, {1, 1}, {-1, -1}, {1, -1},
	}
	res = append(res, checkDirections(diagDirs, isFullDiagonalAttacker)...)

	// Check knights
	knightMoves := [8][2]int{
		{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2},
		{1, -2}, {1, 2}, {2, -1}, {2, 1},
	}
	for _, move := range knightMoves {
		col, row := cellCol+move[0], cellRow+move[1]
		if isValidPosition(col, row) && g.Board[row][col] == Piece("N"+string(attacker)) {
			res = append(res, Position(fmt.Sprintf("%c%d", col+AsciiOffset, row+1)))
		}
	}

	// Check pawns
	pawnMoves := map[Side][2][2]int{
		Black: {{-1, 1}, {1, 1}},
		White: {{-1, -1}, {1, -1}},
	}
	for _, dir := range pawnMoves[Side(attacker)] {
		col, row := cellCol+dir[0], cellRow+dir[1]

		if isValidPosition(col, row) && g.Board[row][col] == Piece("P"+string(attacker)) {
			res = append(res, Position(fmt.Sprintf("%c%d", col+AsciiOffset, row+1)))
		}
	}

	// Check en passant
	f := g.Board[cellRow][cellCol]
	if f[0] == 'P' && len(g.Moves) > 0 {
		prevMove := g.Moves[len(g.Moves)-1]
		pawnAdv := 2
		if Side(f[1]) == Black {
			pawnAdv = -2
		}
		epMove := Move(fmt.Sprintf("P%c%c%d P%c%c%d",
			Side(f[1]), cellCol+AsciiOffset, cellRow,
			Side(f[1]), cellCol+AsciiOffset, cellRow+pawnAdv))
		if prevMove == epMove {
			for _, offset := range []int{-1, 1} {
				if adjCol := cellCol + offset; isValidPosition(adjCol, cellRow) && g.Board[cellRow][adjCol] == Piece("P"+string(attacker)) {
					res = append(res, Position(fmt.Sprintf("%c%d", adjCol+AsciiOffset, cellRow+1)))
				}
			}
		}
	}

	// Check kings
	kingMoves := [8][2]int{
		{-1, 0}, {-1, 1}, {0, 1}, {1, 1},
		{1, 0}, {1, -1}, {0, -1}, {-1, -1},
	}
	for _, dir := range kingMoves {
		col, row := cellCol+dir[0], cellRow+dir[1]
		if isValidPosition(col, row) && g.Board[row][col] == Piece("K"+string(attacker)) {
			res = append(res, Position(fmt.Sprintf("%c%d", col+AsciiOffset, row+1)))
		}
	}

	return res
}

func checkMoveFormat(m Move) bool {
	if m == KingCastling || m == QueenCastling {
		return true
	}

	if len(m) != 9 {
		return false
	}

	if m[4] != Capture && m[4] != Movement && m[4] != Promotion {
		return false
	}

	if (m[2] < 'a' || m[2] > 'h' || m[3] < '1' || m[3] > '8') ||
		(m[5] < 'a' || m[5] > 'h' || m[6] < '1' || m[6] > '8') {
		return false
	}

	return true
}

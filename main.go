package main

import (
	"fmt"
)

func main() {
	fmt.Println("Shahio 0.1")
}

type Move string // e.g. "Pwa2xPbb3" White pawn at a2 captures Black pawn at b3е
type Piece string
type Position [2]int
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

	if !isFormatCorrect(m) {
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

func (g *Game) processCastling(move Move) error {
	// Check that the king didn't move
	for _, prevm := range g.Moves {
		if prevm[0] == 'K' && Side(prevm[1]) == g.whoseTurn {
			return fmt.Errorf("king not in position")
		}
	}

	var (
		kingRow  int
		kingCol  int = 4
		opponent Side
	)

	if g.whoseTurn == Black {
		kingRow = 7
		opponent = White
	} else {
		kingRow = 0
		opponent = Black
	}

	rookPosition := map[Move]Position{KingCastling: {7, kingRow}, QueenCastling: {0, kingRow}}[move]
	rookDir := map[Move]int{KingCastling: 1, QueenCastling: -1}[move]

	// Check that the rook didn't move
	for _, prevm := range g.Moves {
		wasMoved := Piece(prevm[0:2]) == Piece(fmt.Sprintf("R%c", g.whoseTurn)) &&
			string(prevm[2:4]) == fmt.Sprintf("%c%d", rookPosition[0]+AsciiOffset, rookPosition[1])
		wasCaptured := prevm[4] == Capture &&
			Piece(prevm[5:7]) == Piece(fmt.Sprintf("R%c", g.whoseTurn)) &&
			string(prevm[7:9]) == fmt.Sprintf("%c%d", rookPosition[0]+AsciiOffset, rookPosition[1])
		if wasMoved || wasCaptured {
			return fmt.Errorf("rook not in position")
		}
	}

	// Check if there are pieces between the king and the rook
	col := kingCol
	for {
		col += rookDir

		if col == rookPosition[0] {
			break
		}

		if g.Board[kingRow][col] != "" {
			return fmt.Errorf("pieces between king and rook")
		}
	}

	// Check if crossover squares are attacked
	_, crossoverAttacked := g.getAttackingCell(Position{kingCol + rookDir, kingRow}, opponent)
	if crossoverAttacked {
		return fmt.Errorf("crossover cells attacked")
	}

	// Check if king is checked
	_, kingChecked := g.getAttackingCell(Position{kingCol, kingRow}, opponent)
	if kingChecked {
		return fmt.Errorf("king is checked")
	}

	// Set king and rook new positions
	g.Board[kingRow][kingCol+rookDir] = Piece(fmt.Sprintf("R%c", g.whoseTurn))
	g.Board[kingRow][kingCol+2*rookDir] = Piece(fmt.Sprintf("K%c", g.whoseTurn))

	// Clear old positions
	g.Board[kingRow][col] = Piece("")
	g.Board[kingRow][kingCol] = Piece("")

	return nil
}

func (g *Game) getAttackingCell(cell Position, side Side) (Position, bool) {
	checkDirections := func(dirs [4][2]int, isAttacker func(Piece) bool) (Position, bool) {
		for _, dir := range dirs {
			col, row := cell[0]+dir[0], cell[1]+dir[1]
			for ; isValidPosition(col, row); col, row = col+dir[0], row+dir[1] {
				if p := g.Board[row][col]; p != "" {
					if isAttacker(p) {
						return Position{col, row}, true
					}
					break
				}
			}
		}

		return Position{}, false
	}

	// Lines
	isFullLineAttacker := func(p Piece) bool {
		return p == Piece("R"+string(side)) ||
			p == Piece("Q"+string(side))
	}
	lineDirs := [4][2]int{
		{-1, 0}, {1, 0}, {0, -1}, {0, 1},
	}
	line, found := checkDirections(lineDirs, isFullLineAttacker)
	if found {
		return line, true
	}

	// Diagonals
	isFullDiagonalAttacker := func(p Piece) bool {
		return p == Piece("B"+string(side)) ||
			p == Piece("Q"+string(side))
	}
	diagDirs := [4][2]int{
		{-1, 1}, {1, 1}, {-1, -1}, {1, -1},
	}
	diag, found := checkDirections(diagDirs, isFullDiagonalAttacker)
	if found {
		return diag, found
	}

	// Kights
	knightMoves := [8][2]int{
		{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2},
		{1, -2}, {1, 2}, {2, -1}, {2, 1},
	}
	for _, move := range knightMoves {
		col, row := cell[0]+move[0], cell[1]+move[1]
		if isValidPosition(col, row) && g.Board[row][col] == Piece("N"+string(side)) {
			return Position{col, row}, true
		}
	}

	// Pawns
	pawnMoves := map[Side][2][2]int{
		Black: {{-1, 1}, {1, 1}},
		White: {{-1, -1}, {1, -1}},
	}
	for _, dir := range pawnMoves[Side(side)] {
		col, row := cell[0]+dir[0], cell[1]+dir[1]

		if isValidPosition(col, row) && g.Board[row][col] == Piece("P"+string(side)) {
			return Position{col, row}, true
		}
	}

	// En passant
	fig := g.Board[cell[1]][cell[0]]
	if fig[0] == 'P' && len(g.Moves) > 0 {
		prevMove := g.Moves[len(g.Moves)-1]
		pawnAdv := map[Side]int{White: 2, Black: -2}[Side(fig[1])]
		epMove := Move(fmt.Sprintf("P%c%c%d P%c%c%d",
			Side(fig[1]), cell[0]+AsciiOffset, cell[1],
			Side(fig[1]), cell[0]+AsciiOffset, cell[1]+pawnAdv))
		if prevMove == epMove {
			for _, offset := range []int{-1, 1} {
				if adjCol := cell[0] + offset; isValidPosition(adjCol, cell[1]) &&
					g.Board[cell[1]][adjCol] == Piece("P"+string(side)) {
					return Position{adjCol, cell[1]}, true
				}
			}
		}
	}

	// King
	kingMoves := [8][2]int{
		{-1, 0}, {-1, 1}, {0, 1}, {1, 1},
		{1, 0}, {1, -1}, {0, -1}, {-1, -1},
	}
	for _, dir := range kingMoves {
		col, row := cell[0]+dir[0], cell[1]+dir[1]
		if isValidPosition(col, row) && g.Board[row][col] == Piece("K"+string(side)) {
			return Position{col, row}, true
		}
	}

	return Position{}, false
}

func (g *Game) getSourceCell(cell Position, side Side) (Position, bool) {
	if g.Board[cell[0]][cell[1]] == "" {
		return Position{}, false
	}

	atkCell, found := g.getAttackingCell(cell, side)
	if found {
		return atkCell, found
	}

	// Pawn
	pawn := Piece(fmt.Sprintf("P%c", side))
	advDir := map[Side]int{White: 1, Black: -1}[side]
	pawnRow := map[Side]int{White: 1, Black: 6}[side]

	if row := cell[1] - advDir; isValidPosition(cell[0], row) &&
		g.Board[cell[0]][row] == pawn {
		return Position{cell[0], row}, true
	}

	if g.Board[cell[0]][pawnRow] == pawn &&
		g.Board[cell[0]][pawnRow+advDir] == "" {
		return Position{cell[0], pawnRow}, true
	}

	return Position{}, false
}

func isFormatCorrect(m Move) bool {
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

func isValidPosition(col, row int) bool {
	return col > -1 && col < 8 && row > -1 && row < 8
}

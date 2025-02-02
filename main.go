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
type Action byte
type Board [][]Piece
type Game struct {
	Board     Board
	Moves     []Move
	blackKing Position
	whiteKing Position
	ended     bool
}

const (
	White         = Side('w')
	Black         = Side('b')
	KingCastling  = "O-O"
	QueenCastling = "O-O-O"
	Capture       = Action('x')
	Movement      = Action(' ')
	Promotion     = Action('=')
	Enpassant     = Action('e')
	AsciiOffset   = 'a'
	IntOffset     = '0'
)

func NewGame() Game {
	return Game{
		Board: [][]Piece{
			{"Rw", "Nw", "Bw", "Qw", "Kw", "Bw", "Nw", "Rw"},
			{"Pw", "Pw", "Pw", "Pw", "Pw", "Pw", "Pw", "Pw"},
			{"", "", "", "", "", "", "", ""},
			{"", "", "", "", "", "", "", ""},
			{"", "", "", "", "", "", "", ""},
			{"", "", "", "", "", "", "", ""},
			{"Pb", "Pb", "Pb", "Pb", "Pb", "Pb", "Pb", "Pb"},
			{"Rb", "Nb", "Bb", "Qb", "Kb", "Bb", "Nb", "Rb"},
		},
		Moves:     []Move{},
		blackKing: Position{0, 4},
		whiteKing: Position{7, 4},
		ended:     false,
	}
}

func (g *Game) processMove(move Move) error {
	if g.ended {
		return fmt.Errorf("game has ended")
	}

	if !isFormatCorrect(move) {
		return fmt.Errorf("invalid move format")
	}

	actionProcessor := g.getProcessor(move)
	err := actionProcessor(move)
	if err != nil {
		return err
	}

	// Check if your king is in check
	kingAttackers, _ := g.getAttackingCells(g.sideKing(g.whoseTurn()), g.whoseTurn())
	if len(kingAttackers) > 0 {
		return fmt.Errorf("king is in check")
	}

	// Check if enemy king is in checkmate
	if g.checkGameEnded() {
		g.ended = true
		return nil
	}

	return nil
}

func (g *Game) getProcessor(move Move) func(Move) error {
	// Castling
	if move == KingCastling || move == QueenCastling {
		return g.processCastling
	}

	// TODO Add other processors
	return nil
}

func (g *Game) processCastling(move Move) error {
	// Check that king didn't move
	for _, prevm := range g.Moves {
		if prevm[0] == 'K' && Side(prevm[1]) == g.whoseTurn() {
			return fmt.Errorf("king not in position")
		}
	}

	kingRow := map[Side]int{Black: 7, White: 0}[g.whoseTurn()]
	kingCol := 4

	rookPosition := map[Move]Position{KingCastling: {7, kingRow}, QueenCastling: {0, kingRow}}[move]
	rookDir := map[Move]int{KingCastling: 1, QueenCastling: -1}[move]

	// Check that rook didn't move
	for _, prevm := range g.Moves {
		wasMoved := Piece(prevm[0:2]) == Piece(fmt.Sprintf("R%c", g.whoseTurn())) &&
			string(prevm[2:4]) == fmt.Sprintf("%c%d", rookPosition[0]+AsciiOffset, rookPosition[1])
		wasCaptured := Action(prevm[4]) == Capture &&
			Piece(prevm[5:7]) == Piece(fmt.Sprintf("R%c", g.whoseTurn())) &&
			string(prevm[7:9]) == fmt.Sprintf("%c%d", rookPosition[0]+AsciiOffset, rookPosition[1])
		if wasMoved || wasCaptured {
			return fmt.Errorf("rook not in position")
		}
	}

	// Check if there are pieces between king and rook
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

	crossoverAttackers, _ := g.getAttackingCells(Position{kingCol + rookDir, kingRow}, getOpponent(g.whoseTurn()))
	if len(crossoverAttackers) > 0 {
		return fmt.Errorf("crossover cells attacked")
	}

	// Check if king is in check
	kingAttackers, _ := g.getAttackingCells(Position{kingCol, kingRow}, getOpponent(g.whoseTurn()))
	if len(kingAttackers) > 0 {
		return fmt.Errorf("king is in check")
	}

	// Set king and rook new positions
	g.Board[kingRow][kingCol+rookDir] = Piece(fmt.Sprintf("R%c", g.whoseTurn()))
	g.Board[kingRow][kingCol+2*rookDir] = Piece(fmt.Sprintf("K%c", g.whoseTurn()))

	// Clear old positions
	g.Board[kingRow][col] = Piece("")
	g.Board[kingRow][kingCol] = Piece("")

	return nil
}

func (g *Game) checkGameEnded() bool {
	side := getOpponent(g.whoseTurn())
	king := g.sideKing(side)

	atkCells, isBlockable := g.getAttackingCells(king, side)
	if len(atkCells) == 0 {
		return g.checkStalemate()
	}

	// Check if king can run away
	if g.canKingMove(side) {
		return false
	}

	// Mate if attacked by 2 cells and can't run away
	if len(atkCells) == 2 {
		return true
	}

	atkCell := atkCells[0]
	// Check if an attacking cell can be captured
	captCells, _ := g.getAttackingCells(atkCell, side)
	if len(captCells) > 0 {
		return false
	}

	// Check if a blockable line of attack can be blocked
	if isBlockable {
		atkDir := func(a, b int) int {
			if a > b {
				return -1
			} else if a < b {
				return 1
			}
			return 0
		}

		dirCol := atkDir(king[0], atkCell[0])
		dirRow := atkDir(king[1], atkCell[1])
		for col, row := king[0]+dirCol, king[1]+dirRow; isValidPosition(col, row); {
			blkCell, canBlock := g.getSourceCell(Position{col, row}, side)
			if blkCell != king && canBlock {
				return false
			}
		}
	}

	// TODO if only 2 kings left, then game is over

	return true
}

func (g *Game) checkStalemate() bool {
	opponent := getOpponent(g.whoseTurn())
	if g.canKingMove(opponent) {
		return false
	}

	starts := map[Side]int{Black: 7, White: 0}
	dir := map[Side]int{Black: -1, White: 1}[opponent]
	for row := starts[opponent]; row != starts[g.whoseTurn()]+1; row += dir {
		for col := range g.Board[row] {
			if fig := g.Board[row][col]; fig != "" &&
				Side(fig[1]) == opponent && fig[0] != 'K' && g.canMove(Position{col, row}) {
				return true
			}
		}
	}

	return false
}

func (g *Game) canMove(cell Position) bool {
	fig := g.Board[cell[1]][cell[0]]
	side := Side(fig[1])
	king := g.sideKing(side)

	switch fig[0] {
	case 'P':
		// single move
		advDir := map[Side]int{Black: -1, White: 1}[side]
		if col, row := cell[0], cell[1]+advDir; isValidPosition(col, row) &&
			g.Board[row][col] == "" &&
			(cell[1] != king[1] || !g.moveAndCheck(g.createMove(cell, Position{col, row}, Movement), king)) {
			return true
		}

		// attack move
		atkDirs := [2]int{-1, 1}
		for _, dir := range atkDirs {
			if col, row := cell[0]+dir, cell[1]+advDir; isValidPosition(col, row) &&
				(g.Board[row][col] != "" && Side(g.Board[row][col][1]) != getOpponent(side)) &&
				!g.moveAndCheck(g.createMove(cell, Position{col, row}, Capture), king) {
				return true
			}
		}

		// en passant move
		if len(g.Moves) > 0 {
			prevMove := g.Moves[len(g.Moves)-1]
			for _, dir := range atkDirs {
				epMove := Move(fmt.Sprintf("P%c%c%d P%c%c%d",
					getOpponent(side), cell[0]+dir+AsciiOffset, cell[1]+(advDir*2),
					getOpponent(side), cell[0]+dir+AsciiOffset, cell[1]))
				if prevMove == epMove && !g.moveAndCheck(g.createMove(cell, Position{cell[0] + dir, cell[1] + advDir}, Enpassant), king) {
					return true
				}
			}
		}
	case 'Q', 'N':
		knightDirs := [8][2]int{
			{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2},
			{1, -2}, {1, 2}, {2, -1}, {2, 1},
		}
		queenDirs := [8][2]int{
			{-1, 0}, {-1, 1}, {0, 1}, {1, 1},
			{1, 0}, {1, -1}, {0, -1}, {-1, -1},
		}
		dirs := knightDirs
		if fig[0] == 'Q' {
			dirs = queenDirs
		}

		for _, dir := range dirs {
			col, row := cell[0]+dir[0], cell[1]+dir[1]
			if isValidPosition(col, row) &&
				(g.Board[row][col] == "" || Side(g.Board[row][col][1]) != side) {
				return true
			}
		}
	case 'B', 'R':
		dirs := map[byte][4][2]int{
			'B': {{-1, -1}, {1, -1}, {1, 1}, {-1, 1}},
			'R': {{0, -1}, {1, 0}, {0, 1}, {-1, 0}},
		}[fig[0]]
		for _, dir := range dirs {
			col, row := cell[0]+dir[0], cell[1]+dir[1]
			if isValidPosition(col, row) && g.moveAndCheck(g.createMove(cell, Position{col, row}, Movement), king) {
				return true
			}
		}
	}

	return false
}

func (g *Game) canKingMove(side Side) bool {
	king := g.sideKing(side)

	dirs := [8][2]int{
		{-1, 0}, {-1, 1}, {0, 1}, {1, 1},
		{1, 0}, {1, -1}, {0, -1}, {-1, -1},
	}
	for _, dir := range dirs {
		if col, row := king[0]+dir[0], king[1]+dir[1]; isValidPosition(col, row) {
			if cell := g.Board[row][col]; cell != "" && Side(cell[1]) == side {
				continue
			}

			if !g.moveAndCheck(g.createMove(king, Position{col, row}, Movement), king) {
				return true
			}
		}
	}
	return false
}

func (g *Game) moveAndCheck(move Move, check Position) bool {
	checked := false

	var (
		fCol       = move[2] - AsciiOffset
		fRow       = int(move[3] - IntOffset)
		tCol       = move[7] - AsciiOffset
		tRow       = int(move[8] - IntOffset)
		fFig, tFig = g.Board[fRow][fCol], g.Board[tRow][tCol]
		action     = Action(move[4])
	)

	advDir := map[Side]int{Black: -1, White: 1}[Side(move[1])]
	var epPawn Piece

	g.Board[tRow][tCol], g.Board[fRow][fCol] = fFig, ""
	if action == Enpassant {
		epPawn = g.Board[tRow+advDir][tCol]
		g.Board[tRow+advDir][tCol] = ""
	}
	atkCells, _ := g.getAttackingCells(check, getOpponent(Side(move[1])))
	checked = len(atkCells) > 0
	g.Board[tRow][tCol], g.Board[fRow][fCol] = tFig, fFig
	if action == Enpassant {
		g.Board[tRow+advDir][tCol] = epPawn
	}

	return checked
}

// Returns at most 2 cells that can attack the given cell
// bool return parameter indicates if the attack can be blocked
func (g *Game) getAttackingCells(cell Position, side Side) ([]Position, bool) {
	res := []Position{}

	res = append(res, g.getAttackingLines(cell, side)...)
	if len(res) == 2 {
		return res, true
	}
	isBlockable := len(res) == 1

	pawn := Piece(fmt.Sprintf("P%c", side))
	pawnMoves := map[Side][2][2]int{
		Black: {{-1, 1}, {1, 1}},
		White: {{-1, -1}, {1, -1}},
	}
	for _, dir := range pawnMoves[side] {
		col, row := cell[0]+dir[0], cell[1]+dir[1]

		if isValidPosition(col, row) && g.Board[row][col] == pawn {
			res = append(res, Position{col, row})
			if len(res) == 2 {
				return res, isBlockable
			}
		}
	}

	res = append(res, g.getAttackingKnights(cell, side)...)
	if len(res) == 2 {
		return res, isBlockable
	}

	// En passant
	if len(g.Moves) > 0 && g.Board[cell[1]][cell[0]] == Piece('P'+side) {
		prevMove := g.Moves[len(g.Moves)-1]
		pawnAdv := map[Side]int{White: 2, Black: -2}[getOpponent(side)]
		epMove := Move(fmt.Sprintf("P%c%c%d P%c%c%d",
			getOpponent(side), cell[0]+AsciiOffset, cell[1]-pawnAdv,
			getOpponent(side), cell[0]+AsciiOffset, cell[1]))
		if prevMove == epMove {
			for _, offset := range []int{-1, 1} {
				if adjCol := cell[0] + offset; isValidPosition(adjCol, cell[1]) &&
					g.Board[cell[0]][adjCol] == pawn {
					res = append(res, Position{adjCol, cell[1]})
					if len(res) == 2 {
						return res, isBlockable
					}
				}
			}
		}
	}

	king, found := g.getAttackingKing(cell, side)
	if found {
		res = append(res, king)
	}

	return res, isBlockable
}

// Returns first cell that can move to this position
// Technically castling is a move, but it's not used to block a check
// bool return parameter indicates if cell was found
func (g *Game) getSourceCell(cell Position, side Side) (Position, bool) {
	lines := g.getAttackingLines(cell, side)
	if len(lines) > 0 {
		return lines[0], true
	}

	knights := g.getAttackingKnights(cell, side)
	if len(knights) > 0 {
		return knights[0], true
	}

	if king, found := g.getAttackingKing(cell, side); found {
		return king, true
	}

	// Pawn moves by attack
	fig := g.Board[cell[1]][cell[0]]
	pawn := Piece(fmt.Sprintf("P%c", side))
	if fig != "" && Side(fig[1]) == getOpponent(side) {
		pawnMoves := map[Side][3][2]int{
			Black: {{-1, 1}, {1, 1}},
			White: {{-1, -1}, {1, -1}},
		}
		for _, dir := range pawnMoves[Side(side)] {
			col, row := cell[0]+dir[0], cell[1]+dir[1]

			if isValidPosition(col, row) && g.Board[row][col] == pawn {
				return Position{col, row}, true
			}
		}
	}

	// Pawn moves 1 cell
	advDir := map[Side]int{White: -1, Black: 1}[side]
	if row := cell[1] + advDir; isValidPosition(cell[0], row) &&
		g.Board[row][cell[0]] == pawn {
		return Position{cell[0], row}, true
	}

	// Pawn moves 2 cell
	if row := cell[1] + 2*advDir; isValidPosition(cell[0], row) &&
		g.Board[row][cell[0]] == pawn && g.Board[cell[0]][cell[1]] == "" {
		return Position{cell[0], row}, true
	}

	// Pawn moves by en passant
	if row := cell[1] + advDir; len(g.Moves) > 0 && isValidPosition(cell[1], row) &&
		g.Board[row][cell[0]] == Piece(fmt.Sprintf("P%c", getOpponent(side))) {
		prevMove := g.Moves[len(g.Moves)-1]
		epMove := Move(fmt.Sprintf("P%c%c%d P%c%c%d",
			getOpponent(side), cell[0]+AsciiOffset, row-advDir*2,
			getOpponent(side), cell[0]+AsciiOffset, row))
		if prevMove == epMove {
			for _, offset := range []int{-1, 1} {
				if adjCol := cell[0] + offset; isValidPosition(adjCol, row) &&
					g.Board[row][adjCol] == pawn {
					return Position{adjCol, cell[1]}, true
				}
			}
		}
	}

	return Position{}, false
}

func (g *Game) getAttackingLines(cell Position, side Side) []Position {
	res := []Position{}

	checkDirections := func(dirs [4][2]int, isAttacker func(Piece) bool) {
		for _, dir := range dirs {
			col, row := cell[0]+dir[0], cell[1]+dir[1]
			for ; isValidPosition(col, row); col, row = col+dir[0], row+dir[1] {
				if p := g.Board[row][col]; p != "" {
					if isAttacker(p) {
						res = append(res, Position{col, row})
						if len(res) == 2 {
							return
						}
					}
					break
				}
			}
		}
	}

	// Lines
	isFullLineAttacker := func(p Piece) bool {
		return p == Piece(fmt.Sprintf("R%c", side)) ||
			p == Piece(fmt.Sprintf("Q%c", side))
	}
	lineDirs := [4][2]int{
		{-1, 0}, {1, 0}, {0, -1}, {0, 1},
	}
	checkDirections(lineDirs, isFullLineAttacker)
	if len(res) == 2 {
		return res
	}

	// Diagonals
	isFullDiagonalAttacker := func(p Piece) bool {
		return p == Piece(fmt.Sprintf("B%c", side)) ||
			p == Piece(fmt.Sprintf("Q%c", side))
	}
	diagDirs := [4][2]int{
		{-1, 1}, {1, 1}, {-1, -1}, {1, -1},
	}
	checkDirections(diagDirs, isFullDiagonalAttacker)

	return res
}

func (g *Game) getAttackingKnights(cell Position, side Side) []Position {
	res := []Position{}

	knightMoves := [8][2]int{
		{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2},
		{1, -2}, {1, 2}, {2, -1}, {2, 1},
	}
	for _, move := range knightMoves {
		col, row := cell[0]+move[0], cell[1]+move[1]
		if isValidPosition(col, row) && g.Board[row][col] == Piece(fmt.Sprintf("N%c", side)) {
			res = append(res, Position{col, row})
			if len(res) == 2 {
				return res
			}
		}
	}

	return res
}

func (g *Game) getAttackingKing(cell Position, side Side) (Position, bool) {
	kingMoves := [8][2]int{
		{-1, 0}, {-1, 1}, {0, 1}, {1, 1},
		{1, 0}, {1, -1}, {0, -1}, {-1, -1},
	}
	for _, dir := range kingMoves {
		col, row := cell[0]+dir[0], cell[1]+dir[1]
		if isValidPosition(col, row) && g.Board[row][col] == Piece(fmt.Sprintf("K%c", side)) {
			return Position{col, row}, true
		}
	}

	return Position{}, false
}

func (g *Game) sideKing(side Side) Position {
	if side == White {
		return g.whiteKing
	}
	return g.blackKing
}

// TODO Boolean value? True White, False Black
func (g *Game) whoseTurn() Side {
	if len(g.Moves)%2 == 0 {
		return White
	}
	return Black
}

func isFormatCorrect(move Move) bool {
	if move == KingCastling || move == QueenCastling {
		return true
	}

	if len(move) != 9 {
		return false
	}

	if Action(move[4]) != Capture && Action(move[4]) != Movement && Action(move[4]) != Promotion {
		return false
	}

	if (move[2] < 'a' || move[2] > 'h' || move[3] < '1' || move[3] > '8') ||
		(move[5] < 'a' || move[5] > 'h' || move[6] < '1' || move[6] > '8') {
		return false
	}

	return true
}

func isValidPosition(col, row int) bool {
	return col > -1 && col < 8 && row > -1 && row < 8
}

func getOpponent(side Side) Side {
	if side == White {
		return Black
	}
	return White
}

func (g *Game) createMove(from Position, to Position, action Action) Move {
	fFig := g.Board[from[1]][from[0]]
	tFig := g.Board[to[1]][to[0]]
	if tFig == "" {
		tFig = "  "
	}
	return Move(fmt.Sprintf("%s%c%d%c%s%c%d",
		fFig, from[0]+AsciiOffset, from[1], action, tFig, to[0]+AsciiOffset, to[1]))
}

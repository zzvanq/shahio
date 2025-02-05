package main

import (
	"fmt"
)

func main() {
	fmt.Println("Shahio 0.1")
}

type Side byte
type Figure byte
type Piece struct {
	fig  Figure
	side Side
}
type Position struct {
	row int
	col int
}
type Cell struct {
	Piece
	Position
}
type Move struct {
	Source Cell
	Target Cell
	Action Action
}
type Action string
type Board [][]Piece
type Game struct {
	Board      Board
	Moves      []Move
	blackKing  Position
	whiteKing  Position
	blackCells int
	whiteCells int
	ended      bool
}

const (
	White         = Side('w')
	Black         = Side('b')
	KingCastling  = Action("O-O")
	QueenCastling = Action("O-O-O")
	Capture       = Action("x")
	Movement      = Action(" ")
	Promotion     = Action("=")
	Enpassant     = Action("e.p.")
)

var Empty Piece

func NewGame() Game {
	return Game{
		Board: [][]Piece{
			{{'R', 'w'}, {'N', 'w'}, {'B', 'w'}, {'Q', 'w'}, {'K', 'w'}, {'B', 'w'}, {'N', 'w'}, {'R', 'w'}},
			{{'P', 'w'}, {'P', 'w'}, {'P', 'w'}, {'P', 'w'}, {'P', 'w'}, {'P', 'w'}, {'P', 'w'}, {'P', 'w'}},
			{Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty},
			{Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty},
			{Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty},
			{Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty},
			{{'P', 'b'}, {'P', 'b'}, {'P', 'b'}, {'P', 'b'}, {'P', 'b'}, {'P', 'b'}, {'P', 'b'}, {'P', 'b'}},
			{{'R', 'b'}, {'N', 'b'}, {'B', 'b'}, {'Q', 'b'}, {'K', 'b'}, {'B', 'b'}, {'N', 'b'}, {'R', 'b'}},
		},
		Moves:      []Move{},
		whiteKing:  Position{row: 0, col: 4},
		blackKing:  Position{row: 7, col: 4},
		blackCells: 16,
		whiteCells: 16,
		ended:      false,
	}
}

func (g *Game) processMove(move Move) error {
	if g.ended {
		return fmt.Errorf("game has ended")
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
	if move.Action == KingCastling || move.Action == QueenCastling {
		return g.processCastling
	}

	// TODO Add other processors
	return nil
}

func (g *Game) processCastling(move Move) error {
	// Check that king didn't move
	for _, prevm := range g.Moves {
		if prevm.Source.fig == 'K' && prevm.Source.side == g.whoseTurn() {
			return fmt.Errorf("king not in position")
		}
	}

	kingRows := map[Side]int{Black: 7, White: 0}
	king := Cell{'K', g.whoseTurn(), kingRows[g.whoseTurn()], 4}

	rook := map[Action]Position{KingCastling: {col: 7, row: king.row}, QueenCastling: {col: 0, row: king.row}}[move.Action]

	// Check that rook didn't move
	for _, prevm := range g.Moves {
		wasMoved := prevm.Source.Piece == Piece{'R', g.whoseTurn()} && prevm.Source.Position == rook
		wasCaptured := prevm.Action == Capture &&
			prevm.Target.Piece == Piece{'R', g.whoseTurn()} && prevm.Source.Position == rook
		if wasMoved || wasCaptured {
			return fmt.Errorf("rook not in position")
		}
	}

	// Check if there are pieces between king and rook
	col := king.col
	rookDir := map[Action]int{KingCastling: 1, QueenCastling: -1}[move.Action]
	for {
		col += rookDir

		if col == rook.col {
			break
		}

		if g.Board[king.row][col] != Empty {
			return fmt.Errorf("pieces between king and rook")
		}
	}

	// Check if crossover squares are attacked

	crossoverAttackers, _ := g.getAttackingCells(Position{row: king.row, col: king.col + rookDir}, getOpponent(g.whoseTurn()))
	if len(crossoverAttackers) > 0 {
		return fmt.Errorf("crossover cells attacked")
	}

	// Check if king is in check
	kingAttackers, _ := g.getAttackingCells(Position{row: king.row, col: king.col}, getOpponent(g.whoseTurn()))
	if len(kingAttackers) > 0 {
		return fmt.Errorf("king is in check")
	}

	// Set king and rook new positions
	g.Board[king.row][king.col+rookDir] = Piece{'R', g.whoseTurn()}
	g.Board[king.row][king.col+2*rookDir] = Piece{'K', g.whoseTurn()}

	// Clear old positions
	g.Board[king.row][rook.col] = Empty
	g.Board[king.row][king.col] = Empty

	if g.whoseTurn() == White {
		g.whiteKing = Position{row: king.row, col: king.col + 2*rookDir}
	} else {
		g.blackKing = Position{row: king.row, col: king.col + 2*rookDir}
	}

	return nil
}

func (g *Game) checkGameEnded() bool {
	if g.blackCells == 1 || g.whiteCells == 1 {
		return true
	}

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

		dirCol := atkDir(king.col, atkCell.col)
		dirRow := atkDir(king.row, atkCell.row)
		for col, row := king.col+dirCol, king.row+dirRow; isValidPosition(col, row); {
			blkCell, canBlock := g.getSourceCell(Position{row: row, col: col}, side)
			if blkCell != king && canBlock {
				return false
			}
			col, row = col+dirCol, row+dirRow
		}
	}

	return true
}

func (g *Game) checkStalemate() bool {
	opponent := getOpponent(g.whoseTurn())
	if g.canKingMove(opponent) {
		return false
	}

	starts := map[Side]int{Black: 7, White: 0}
	dir := map[Side]int{Black: -1, White: 1}[opponent]
	for row := starts[opponent]; row != starts[g.whoseTurn()]; row += dir {
		for col := range g.Board[row] {
			if pic := g.Board[row][col]; pic != Empty &&
				pic.side == opponent && pic.fig != 'K' && g.canMove(Position{row: row, col: col}) {
				return true
			}
		}
	}

	return false
}

func (g *Game) canMove(cell Position) bool {
	pic := g.Board[cell.row][cell.col]
	king := g.sideKing(pic.side)

	switch pic.fig {
	case 'P':
		// single move
		advDir := map[Side]int{Black: -1, White: 1}[pic.side]
		if col, row := cell.col, cell.row+advDir; isValidPosition(col, row) && g.Board[row][col] == Empty &&
			!g.moveAndCheck(Move{Source: Cell{pic, cell}, Target: Cell{Empty, Position{row, col}}, Action: Movement}, king) {
			return true
		}

		// attack move
		atkDirs := [2]int{-1, 1}
		for _, dir := range atkDirs {
			if col, row := cell.col+dir, cell.row+advDir; isValidPosition(col, row) &&
				(g.Board[row][col] != Empty && g.Board[row][col].side == getOpponent(pic.side)) &&
				!g.moveAndCheck(Move{
					Source: Cell{pic, cell},
					Target: Cell{g.Board[row][col], Position{row: row, col: col}},
					Action: Capture,
				}, king) {
				return true
			}
		}

		// en passant move
		if len(g.Moves) > 0 {
			prevMove := g.Moves[len(g.Moves)-1]
			for _, dir := range atkDirs {
				preEpMove := Move{
					Source: Cell{Piece{'P', getOpponent(pic.side)}, Position{row: cell.row, col: cell.col + dir}},
					Target: Cell{Piece{'P', getOpponent(pic.side)}, Position{row: cell.row + (advDir * 2), col: cell.col + dir}},
					Action: Movement,
				}
				if prevMove == preEpMove &&
					!g.moveAndCheck(Move{
						Source: Cell{pic, cell},
						Target: Cell{Empty, Position{row: cell.row + advDir, col: cell.col + dir}},
						Action: Enpassant,
					}, king) {
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
		if pic.fig == 'Q' {
			dirs = queenDirs
		}

		for _, dir := range dirs {
			col, row := cell.col+dir[0], cell.row+dir[1]
			if isValidPosition(col, row) &&
				(g.Board[row][col] == Empty || g.Board[row][col].side != pic.side) &&
				!g.moveAndCheck(Move{
					Source: Cell{pic, cell},
					Target: Cell{g.Board[row][col], Position{row: row, col: col}},
					Action: Movement,
				}, king) {
				return true
			}
		}
	case 'B', 'R':
		dirs := map[Figure][4][2]int{
			'B': {{-1, -1}, {1, -1}, {1, 1}, {-1, 1}},
			'R': {{0, -1}, {1, 0}, {0, 1}, {-1, 0}},
		}[pic.fig]
		for _, dir := range dirs {
			col, row := cell.col+dir[0], cell.row+dir[1]
			if isValidPosition(col, row) &&
				!g.moveAndCheck(Move{
					Source: Cell{pic, cell},
					Target: Cell{g.Board[row][col], Position{row: row, col: col}},
					Action: Movement,
				}, king) {
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
		if col, row := king.col+dir[0], king.row+dir[1]; isValidPosition(col, row) {
			if pic := g.Board[row][col]; pic != Empty && pic.side == side {
				continue
			}

			if !g.moveAndCheck(Move{
				Source: Cell{g.Board[king.row][king.col], king},
				Target: Cell{g.Board[row][col], Position{row: row, col: col}},
				Action: Movement,
			}, king) {
				return true
			}
		}
	}
	return false
}

func (g *Game) moveAndCheck(move Move, check Position) bool {
	checked := false

	advDir := map[Side]int{Black: -1, White: 1}[move.Source.side]
	var epPawn Piece

	g.Board[move.Source.row][move.Source.col] = move.Source.Piece
	g.Board[move.Source.row][move.Source.col] = Empty
	if move.Action == Enpassant {
		epPawn = g.Board[move.Source.row+advDir][move.Source.col]
		g.Board[move.Source.row+advDir][move.Source.col] = Empty
	}
	defer func() {
		g.Board[move.Source.row][move.Source.col] = move.Target.Piece
		g.Board[move.Source.row][move.Source.col] = move.Source.Piece
		if move.Action == Enpassant {
			g.Board[move.Source.row+advDir][move.Source.col] = epPawn
		}
	}()

	atkCells, _ := g.getAttackingCells(check, getOpponent(move.Source.side))
	checked = len(atkCells) > 0
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

	pawn := Piece{'P', side}
	pawnMoves := map[Side][2][2]int{
		Black: {{-1, -1}, {1, -1}},
		White: {{-1, 1}, {1, 1}},
	}
	for _, dir := range pawnMoves[side] {
		col, row := cell.col+dir[0], cell.row+dir[1]

		if isValidPosition(col, row) && g.Board[row][col] == pawn {
			res = append(res, Position{row: row, col: col})
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
	opPawn := Piece{'P', getOpponent(side)}
	if len(g.Moves) > 0 && g.Board[cell.row][cell.col] == opPawn {
		prevMove := g.Moves[len(g.Moves)-1]
		advDir := map[Side]int{White: 1, Black: -1}[getOpponent(side)]
		if isValidPosition(cell.col, cell.row-(advDir*2)) {
			preEpMove := Move{
				Source: Cell{g.Board[cell.row-(advDir*2)][cell.col], Position{row: cell.row - (advDir * 2), col: cell.col}},
				Target: Cell{g.Board[cell.row][cell.col], Position{row: cell.row, col: cell.col}},
				Action: Movement,
			}
			if prevMove == preEpMove {
				for _, offset := range []int{-1, 1} {
					if adjCol := cell.col + offset; isValidPosition(adjCol, cell.row) &&
						g.Board[cell.row][adjCol] == pawn {
						res = append(res, Position{row: cell.row, col: adjCol})
						if len(res) == 2 {
							return res, isBlockable
						}
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
Position	if len(knights) > 0 {
		return knights[0], true
	}

	if king, found := g.getAttackingKing(cell, side); found {
		return king, true
	}

	// Pawn moves by attack
	pic := g.Board[cell.row][cell.col]
	pawn := Piece{'P', side}
	if pic != Empty && pic.side == getOpponent(side) {
		pawnMoves := map[Side][3][2]int{
			Black: {{-1, 1}, {1, 1}},
			White: {{-1, -1}, {1, -1}},
		}
		for _, dir := range pawnMoves[Side(side)] {
			col, row := cell.col+dir[0], cell.row+dir[1]

			if isValidPosition(col, row) && g.Board[row][col] == pawn {
				return Position{row: row, col: col}, true
			}
		}
	}

	// Pawn moves 1 cell
	advDir := map[Side]int{White: 1, Black: -1}[side]
	if row := cell.row - advDir; isValidPosition(cell.col, row) &&
		g.Board[row][cell.col] == pawn {
		return Position{row: row, col: cell.col}, true
	}

	// Pawn moves 2 cell
	if row := cell.row - 2*advDir; isValidPosition(cell.col, row) &&
		g.Board[row][cell.col] == pawn && g.Board[cell.row][cell.col] == Empty {
		return Position{row: row, col: cell.col}, true
	}

	// Pawn moves by en passant
	if row := cell.row - advDir; len(g.Moves) > 0 && isValidPosition(cell.col, row) &&
		g.Board[row][cell.col] == (Piece{'P', getOpponent(side)}) {
		prevMove := g.Moves[len(g.Moves)-1]
		preEpMove := Move{
			Source: Cell{g.Board[row][cell.col], Position{row: row + (advDir * 2), col: cell.col}},
			Target: Cell{Empty, Position{row: row, col: cell.col}},
			Action: Movement,
		}
		if prevMove == preEpMove {
			for _, offset := range []int{-1, 1} {
				if adjCol := cell.col + offset; isValidPosition(adjCol, row) &&
					g.Board[row][adjCol] == pawn {
					return Position{row: row, col: adjCol}, true
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
			col, row := cell.col+dir[0], cell.row+dir[1]
			for ; isValidPosition(col, row); col, row = col+dir[0], row+dir[1] {
				if p := g.Board[row][col]; p != Empty {
					if isAttacker(p) {
						res = append(res, Position{row: row, col: col})
						if len(res) == 2 {
							return
						}
					}
					break
				}
			}
		}
	}

Position	// Lines
	isFullLineAttacker := func(p Piece) bool {
		return p == Piece{'R', side} ||
			p == Piece{'Q', side}
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
		return p == Piece{'B', side} ||
			p == Piece{'Q', side}
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
		col, row := cell.col+move[0], cell.row+move[1]
		if isValidPosition(col, row) && g.Board[row][col] == (Piece{'N', side}) {
			res = append(res, Position{row: row, col: col})
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
		col, row := cell.col+dir[0], cell.row+dir[1]
		if isValidPosition(col, row) && g.Board[row][col] == (Piece{'K', side}) {
			return Position{row: row, col: col}, true
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

func (g *Game) whoseTurn() Side {
	if len(g.Moves)%2 == 0 {
		return White
	}
	return Black
}

func isValidPosition(col, row int) bool {
	return col > -1 && col < 8 && row > -1 && row < 8
}

func getOpponent(side Side) Side {
	if side == WPosition{
		return Black
	}
	return White
}

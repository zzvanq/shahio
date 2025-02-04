package main

import (
	"fmt"
)

func main() {
	fmt.Println("Shahio 0.1")
}

// "Pwa2xPbb3" White pawn at a2 captures Black pawn at b3е
type Piece string
type Position struct {
	row int
	col int
}
type Move struct {
	Source Piece
	Target Piece
	From   Position
	To     Position
	Action Action
}
type Side byte
type Action string
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
	KingCastling  = Action("O-O")
	QueenCastling = Action("O-O-O")
	Capture       = Action("x")
	Movement      = Action(" ")
	Promotion     = Action("=")
	Enpassant     = Action("e.p.")
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
		whiteKing: Position{row: 0, col: 4},
		blackKing: Position{row: 7, col: 4},
		ended:     false,
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
		if prevm.Source[0] == 'K' && Side(prevm.Source[1]) == g.whoseTurn() {
			return fmt.Errorf("king not in position")
		}
	}

	kingRow := map[Side]int{Black: 7, White: 0}[g.whoseTurn()]
	kingCol := 4

	rook := map[Action]Position{KingCastling: {col: 7, row: kingRow}, QueenCastling: {col: 0, row: kingRow}}[move.Action]

	rookDir := map[Action]int{KingCastling: 1, QueenCastling: -1}[move.Action]

	// Check that rook didn't move
	for _, prevm := range g.Moves {
		wasMoved := prevm.Source == Piece(fmt.Sprintf("R%c", g.whoseTurn())) && prevm.From == rook
		wasCaptured := prevm.Action == Capture &&
			prevm.Target == Piece(fmt.Sprintf("R%c", g.whoseTurn())) && prevm.To == rook
		if wasMoved || wasCaptured {
			return fmt.Errorf("rook not in position")
		}
	}

	// Check if there are pieces between king and rook
	col := kingCol
	for {
		col += rookDir

		if col == rook.col {
			break
		}

		if g.Board[kingRow][col] != "" {
			return fmt.Errorf("pieces between king and rook")
		}
	}

	// Check if crossover squares are attacked

	crossoverAttackers, _ := g.getAttackingCells(Position{row: kingRow, col: kingCol + rookDir}, getOpponent(g.whoseTurn()))
	if len(crossoverAttackers) > 0 {
		return fmt.Errorf("crossover cells attacked")
	}

	// Check if king is in check
	kingAttackers, _ := g.getAttackingCells(Position{row: kingRow, col: kingCol}, getOpponent(g.whoseTurn()))
	if len(kingAttackers) > 0 {
		return fmt.Errorf("king is in check")
	}

	// Set king and rook new positions
	g.Board[kingRow][kingCol+rookDir] = Piece(fmt.Sprintf("R%c", g.whoseTurn()))
	g.Board[kingRow][kingCol+2*rookDir] = Piece(fmt.Sprintf("K%c", g.whoseTurn()))

	// Clear old positions
	g.Board[kingRow][rook.col] = Piece("")
	g.Board[kingRow][kingCol] = Piece("")

	if g.whoseTurn() == White {
		g.whiteKing = Position{row: kingRow, col: kingCol + 2*rookDir}
	} else {
		g.blackKing = Position{row: kingRow, col: kingCol + 2*rookDir}
	}

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
	for row := starts[opponent]; row != starts[g.whoseTurn()]; row += dir {
		for col := range g.Board[row] {
			if fig := g.Board[row][col]; fig != "" &&
				Side(fig[1]) == opponent && fig[0] != 'K' && g.canMove(Position{row: row, col: col}) {
				return true
			}
		}
	}

	return false
}

func (g *Game) canMove(cell Position) bool {
	fig := g.Board[cell.row][cell.col]
	side := Side(fig[1])
	king := g.sideKing(side)

	switch fig[0] {
	case 'P':
		// single move
		advDir := map[Side]int{Black: -1, White: 1}[side]
		if col, row := cell.col, cell.row+advDir; isValidPosition(col, row) && g.Board[row][col] == "" &&
			!g.moveAndCheck(Move{Source: fig, Target: " ", From: cell, To: Position{row, col}, Action: Movement}, king) {
			return true
		}

		// attack move
		atkDirs := [2]int{-1, 1}
		for _, dir := range atkDirs {
			if col, row := cell.col+dir, cell.row+advDir; isValidPosition(col, row) &&
				(g.Board[row][col] != "" && Side(g.Board[row][col][1]) == getOpponent(side)) &&
				!g.moveAndCheck(Move{
					Source: fig,
					Target: g.Board[row][col],
					From:   cell,
					To:     Position{row: row, col: col},
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
					Source: Piece(fmt.Sprintf("P%c", getOpponent(side))),
					Target: Piece(fmt.Sprintf("P%c", getOpponent(side))),
					From:   Position{row: cell.row, col: cell.col + dir},
					To:     Position{row: cell.row + (advDir * 2), col: cell.col + dir},
					Action: Movement,
				}
				if prevMove == preEpMove &&
					!g.moveAndCheck(Move{
						Source: fig,
						Target: " ",
						From:   cell,
						To:     Position{row: cell.row + advDir, col: cell.col + dir},
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
		if fig[0] == 'Q' {
			dirs = queenDirs
		}

		for _, dir := range dirs {
			col, row := cell.col+dir[0], cell.row+dir[1]
			if isValidPosition(col, row) &&
				(g.Board[row][col] == "" || Side(g.Board[row][col][1]) != side) &&
				!g.moveAndCheck(Move{
					Source: fig,
					Target: g.Board[row][col],
					From:   cell,
					To:     Position{row: row, col: col},
					Action: Movement,
				}, king) {
				return true
			}
		}
	case 'B', 'R':
		dirs := map[byte][4][2]int{
			'B': {{-1, -1}, {1, -1}, {1, 1}, {-1, 1}},
			'R': {{0, -1}, {1, 0}, {0, 1}, {-1, 0}},
		}[fig[0]]
		for _, dir := range dirs {
			col, row := cell.col+dir[0], cell.row+dir[1]
			if isValidPosition(col, row) &&
				!g.moveAndCheck(Move{
					Source: fig,
					Target: g.Board[row][col],
					From:   cell,
					To:     Position{row: row, col: col},
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
			if fig := g.Board[row][col]; fig != "" && Side(fig[1]) == side {
				continue
			}

			if !g.moveAndCheck(Move{
				Source: g.Board[king.row][king.col],
				Target: g.Board[row][col],
				From:   king,
				To:     Position{row: row, col: col},
				Action: Movement,
			}, king) {
				return true
			}
		}
	}
	return false
}

func (g *Game) moveAndCheck(move Move, check Position) bool {
	// TODO Defer undo of a move
	checked := false

	advDir := map[Side]int{Black: -1, White: 1}[Side(move.Source[1])]
	var epPawn Piece

	g.Board[move.To.row][move.To.col], g.Board[move.From.row][move.From.col] = move.Source, ""
	if move.Action == Enpassant {
		epPawn = g.Board[move.To.row+advDir][move.To.col]
		g.Board[move.To.row+advDir][move.To.col] = ""
	}
	atkCells, _ := g.getAttackingCells(check, getOpponent(Side(move.Source[1])))
	checked = len(atkCells) > 0
	g.Board[move.To.row][move.To.col], g.Board[move.From.row][move.From.col] = move.Target, move.Source
	if move.Action == Enpassant {
		g.Board[move.To.row+advDir][move.To.col] = epPawn
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
	if len(g.Moves) > 0 && g.Board[cell.row][cell.col] == Piece(fmt.Sprintf("P%c", getOpponent(side))) {
		prevMove := g.Moves[len(g.Moves)-1]
		advDir := map[Side]int{White: 1, Black: -1}[getOpponent(side)]
		if isValidPosition(cell.col, cell.row-(advDir*2)) {
			preEpMove := Move{
				Source: g.Board[cell.row-(advDir*2)][cell.col],
				Target: g.Board[cell.row][cell.col],
				From:   Position{row: cell.row - (advDir * 2), col: cell.col},
				To:     Position{row: cell.row, col: cell.col},
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
	if len(knights) > 0 {
		return knights[0], true
	}

	if king, found := g.getAttackingKing(cell, side); found {
		return king, true
	}

	// Pawn moves by attack
	fig := g.Board[cell.row][cell.col]
	pawn := Piece(fmt.Sprintf("P%c", side))
	if fig != "" && Side(fig[1]) == getOpponent(side) {
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
		g.Board[row][cell.col] == pawn && g.Board[cell.row][cell.col] == "" {
		return Position{row: row, col: cell.col}, true
	}

	// Pawn moves by en passant
	if row := cell.row - advDir; len(g.Moves) > 0 && isValidPosition(cell.col, row) &&
		g.Board[row][cell.col] == Piece(fmt.Sprintf("P%c", getOpponent(side))) {
		prevMove := g.Moves[len(g.Moves)-1]
		preEpMove := Move{
			Source: g.Board[row][cell.col],
			Target: " ",
			From:   Position{row: row + (advDir * 2), col: cell.col},
			To:     Position{row: row, col: cell.col},
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
				if p := g.Board[row][col]; p != "" {
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
		col, row := cell.col+move[0], cell.row+move[1]
		if isValidPosition(col, row) && g.Board[row][col] == Piece(fmt.Sprintf("N%c", side)) {
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
		if isValidPosition(col, row) && g.Board[row][col] == Piece(fmt.Sprintf("K%c", side)) {
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

// TODO Boolean value? True White, False Black
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
	if side == White {
		return Black
	}
	return White
}

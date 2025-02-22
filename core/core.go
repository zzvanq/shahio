package core

import (
	"fmt"
	"slices"
)

func main() {
	fmt.Println("Shahio 0.1")
}

type (
	Outcome uint8
	Side    byte
	Figure  byte
	Piece   struct {
		fig  Figure
		side Side
	}
)

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
type (
	Action string
	Board  [][]Piece
	Game   struct {
		Board      Board
		Moves      []Move
		blackKing  Position
		whiteKing  Position
		blackCells int
		whiteCells int
		outcome    Outcome
	}
)

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
const (
	Checkmate = iota
	Stalemate
	NoOutcome
)

var (
	PawnDirs   = map[Side][][2]int{Black: {{0, -1}, {0, -2}}, White: {{0, 1}, {0, 2}}}
	DiagDirs   = [][2]int{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}}
	LineDirs   = [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	KnightDirs = [][2]int{
		{-2, -1},
		{-2, 1},
		{-1, -2},
		{-1, 2},
		{1, -2},
		{1, 2},
		{2, -1},
		{2, 1},
	}
	KingDirs = [][2]int{
		{-1, 0},
		{-1, 1},
		{0, 1},
		{1, 1},
		{1, 0},
		{1, -1},
		{0, -1},
		{-1, -1},
	}
	PicDirs = map[Figure][][2]int{
		'B': DiagDirs, 'R': LineDirs,
		'Q': append(DiagDirs, LineDirs...), 'N': KnightDirs,
		'K': KingDirs,
	}
	PawnAtkDirs = map[Side][][2]int{Black: {{-1, -1}, {1, -1}}, White: {{-1, 1}, {1, 1}}}
	AdvDirs     = map[Side]int{Black: -1, White: 1}
	Empty       Piece
)

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
		outcome:    NoOutcome,
	}
}

func (g *Game) processMove(move Move) error {
	if g.outcome != NoOutcome {
		return fmt.Errorf("game has ended")
	}

	actionProcessor := g.getProcessor(move)
	if actionProcessor == nil {
		return fmt.Errorf("invalid move action: %v", move.Action)
	}

	err := actionProcessor(move)
	if err != nil {
		return err
	}

	kingAttackers, _ := g.getAttackingCells(g.sideKing(g.whoseTurn()), getOpponent(g.whoseTurn()))
	if len(kingAttackers) > 0 {
		return fmt.Errorf("king is in check")
	}

	g.outcome = g.checkGameStatus()

	g.Moves = append(g.Moves, move)

	return nil
}

func (g *Game) getProcessor(move Move) func(Move) error {
	switch move.Action {
	case KingCastling, QueenCastling:
		return g.processCastling
	case Movement:
		return g.processMovement
	case Capture:
		return g.processCapture
	case Promotion:
		return g.processPromotion
	case Enpassant:
		return g.processEnpassant
	}
	return nil
}

func (g *Game) processCastling(move Move) error {
	kingRows := map[Side]int{Black: 7, White: 0}
	king := Position{kingRows[g.whoseTurn()], 4}
	rook := map[Action]Position{KingCastling: {col: 7, row: king.row}, QueenCastling: {col: 0, row: king.row}}[move.Action]

	for _, prevm := range g.Moves {
		// Check that king didn't move
		if prevm.Source.fig == 'K' && prevm.Source.side == g.whoseTurn() {
			return fmt.Errorf("king not in position")
		}

		// Check that rook didn't move
		wasMoved := prevm.Source.Piece == Piece{'R', g.whoseTurn()} && prevm.Source.Position == rook
		wasCaptured := prevm.Action == Capture &&
			prevm.Target.Piece == Piece{'R', g.whoseTurn()} && prevm.Target.Position == rook
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
		return fmt.Errorf("crossover cell attacked")
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

func (g *Game) processMovement(move Move) error {
	target := g.Board[move.Target.row][move.Target.col]

	if target.side == g.whoseTurn() {
		return fmt.Errorf("cell is occupied")
	}

	// Check if line is blocked
	switch move.Source.fig {
	case 'Q', 'B', 'R', 'P':
		stepRow, stepCol := 0, 0
		if move.Target.row > move.Source.row {
			stepRow = 1
		} else if move.Target.row < move.Source.row {
			stepRow = -1
		}

		if move.Target.col > move.Source.col {
			stepCol = 1
		} else if move.Target.col < move.Source.col {
			stepCol = -1
		}

		for col, row := move.Source.col, move.Source.row; ; {
			col, row = col+stepCol, row+stepRow
			if !isValidPosition(col, row) {
				break
			}

			if g.Board[row][col] != Empty {
				return fmt.Errorf("move is blocked")
			}

			// Target position reached
			if move.Target.row == row && move.Target.col == col {
				break
			}
		}
	}

	if err := checkMoveDir(move); err != nil {
		return err
	}

	g.moveCell(move.Source, move.Target)

	return nil
}

func (g *Game) processCapture(move Move) error {
	target := g.Board[move.Target.row][move.Target.col]
	if target == Empty {
		return fmt.Errorf("cell is empty")
	}

	if target.side == g.whoseTurn() {
		return fmt.Errorf("cell is occupied")
	}

	if move.Source.fig != 'P' {
		return g.processMovement(move)
	}

	if err := checkAtkDir(move); err != nil {
		return err
	}

	g.moveCell(move.Source, move.Target)

	return nil
}

func (g *Game) processEnpassant(move Move) error {
	if g.Board[move.Target.row][move.Target.col] != Empty {
		return fmt.Errorf("cell is not empty")
	}

	if move.Source.fig != 'P' {
		return fmt.Errorf("invalid move")
	}

	col, row := move.Source.col+(move.Target.col-move.Source.col), move.Source.row
	epCapPic := g.Board[row][col]
	if len(g.Moves) > 0 && epCapPic == (Piece{'P', getOpponent(g.whoseTurn())}) {
		preEpMove := Move{
			Source: Cell{Piece{'P', epCapPic.side}, Position{row - (AdvDirs[epCapPic.side] * 2), col}},
			Target: Cell{Piece{'P', epCapPic.side}, Position{row, col}},
			Action: Movement,
		}
		if prevMove := g.Moves[len(g.Moves)-1]; prevMove == preEpMove {
			g.moveCell(move.Source, move.Target)
			g.Board[row][col] = Empty
			return nil
		}
	}

	return fmt.Errorf("invalid move")
}

func (g *Game) processPromotion(move Move) error {
	if move.Source.fig != 'P' || move.Target.side != move.Source.side {
		return fmt.Errorf("invalid move")
	}

	var err error
	if move.Target.col == move.Source.col {
		err = g.processMovement(move)
	} else {
		err = g.processCapture(move)
	}
	if err != nil {
		return err
	}

	g.Board[move.Target.row][move.Target.col] = move.Target.Piece

	return nil
}

func (g *Game) checkGameStatus() Outcome {
	if g.blackCells == 1 || g.whiteCells == 1 {
		return Stalemate
	}

	side := getOpponent(g.whoseTurn())
	king := g.sideKing(side)

	// If king is not attacked, check for stalemate
	atkCells, isBlockable := g.getAttackingCells(king, side)
	if len(atkCells) == 0 && g.checkStalemate() {
		return Stalemate
	}

	// Check if checked king can run away
	if g.canKingMove(side) {
		return Checkmate
	}

	// Mate if attacked by 2 cells and can't run away
	if len(atkCells) == 2 {
		return Checkmate
	}

	// Check if the attacking cell can be captured
	atkCell := atkCells[0]
	captCells, _ := g.getAttackingCells(atkCell, side)
	if len(captCells) > 0 {
		return NoOutcome
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
			blockerCell, found := g.getSourceCell(Position{row: row, col: col}, side)
			if found && blockerCell != king {
				return NoOutcome
			}
			col, row = col+dirCol, row+dirRow
		}
	}

	return Checkmate
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
				pic.side == opponent && g.canMove(Position{row: row, col: col}) {
				return false
			}
		}
	}

	return true
}

func (g *Game) canMove(cell Position) bool {
	pic := g.Board[cell.row][cell.col]
	king := g.sideKing(pic.side)

	switch pic.fig {
	case 'P':
		// single move
		if col, row := cell.col, cell.row+AdvDirs[g.whoseTurn()]; isValidPosition(col, row) && g.Board[row][col] == Empty &&
			!g.moveAndCheck(Move{Source: Cell{pic, cell}, Target: Cell{Empty, Position{row, col}}, Action: Movement}, king) {
			return true
		}

		// attack move
		for _, dir := range PawnAtkDirs[g.whoseTurn()] {
			if col, row := cell.col+dir[0], cell.row+dir[1]; isValidPosition(col, row) &&
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
			for _, dir := range PawnAtkDirs[g.whoseTurn()] {
				preEpMove := Move{
					Source: Cell{Piece{'P', getOpponent(pic.side)}, Position{row: cell.row, col: cell.col + dir[0]}},
					Target: Cell{Piece{'P', getOpponent(pic.side)}, Position{row: cell.row + (AdvDirs[g.whoseTurn()] * 2), col: cell.col + dir[0]}},
					Action: Movement,
				}
				if prevMove == preEpMove &&
					!g.moveAndCheck(Move{
						Source: Cell{pic, cell},
						Target: Cell{Empty, Position{row: cell.row + AdvDirs[g.whoseTurn()], col: cell.col + dir[0]}},
						Action: Enpassant,
					}, king) {
					return true
				}
			}
		}
	case 'Q', 'N':
		dirs := PicDirs['N']
		if pic.fig == 'Q' {
			dirs = PicDirs['Q']
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
		for _, dir := range PicDirs[pic.fig] {
			col, row := cell.col+dir[0], cell.row+dir[1]
			if isValidPosition(col, row) && g.Board[row][col].side != pic.side &&
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

	for _, dir := range PicDirs['K'] {
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
	advDir := AdvDirs[g.whoseTurn()]
	var epPawn Piece

	// Make move
	g.moveCell(move.Source, move.Target)
	if move.Action == Enpassant {
		epPawn = g.Board[move.Source.row+advDir][move.Source.col]
		g.Board[move.Source.row+advDir][move.Source.col] = Empty
	}

	// Defer restoring of state
	defer func() {
		g.cancelMoveCell(move.Source, move.Target)
		if move.Action == Enpassant {
			g.Board[move.Source.row+advDir][move.Source.col] = epPawn
		}
	}()

	// Check for check
	atkCells, _ := g.getAttackingCells(check, getOpponent(move.Source.side))

	return len(atkCells) > 0
}

// Returns at most 2 cells that can attack the given cell
// bool return parameter indicates if the attack can be blocked
func (g *Game) getAttackingCells(cell Position, side Side) ([]Position, bool) {
	res := make([]Position, 0, 2)

	res = append(res, g.getAttackingLines(cell, side)...)
	if len(res) == 2 {
		return res, true
	}
	isBlockable := len(res) == 1

	// Pawn attacks
	pawn := Piece{'P', side}
	for _, dir := range PawnAtkDirs[side] {
		col, row := cell.col-dir[0], cell.row-dir[1]

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
	epPawn := Piece{'P', getOpponent(side)}
	if len(g.Moves) > 0 && g.Board[cell.row][cell.col] == epPawn {
		prevMove := g.Moves[len(g.Moves)-1]
		advDir := AdvDirs[getOpponent(side)]
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
	if len(knights) > 0 {
		return knights[0], true
	}

	if king, found := g.getAttackingKing(cell, side); found {
		return king, true
	}

	// Pawn moves by attack
	pic := g.Board[cell.row][cell.col]
	pawn := Piece{'P', side}
	if pic != Empty && pic.side == getOpponent(side) {
		for _, dir := range PawnAtkDirs[getOpponent(side)] {
			col, row := cell.col+dir[0], cell.row+dir[1]

			if isValidPosition(col, row) && g.Board[row][col] == pawn {
				return Position{row: row, col: col}, true
			}
		}
	}

	// Pawn moves 1 cell
	if row := cell.row - AdvDirs[side]; isValidPosition(cell.col, row) &&
		g.Board[row][cell.col] == pawn {
		return Position{row: row, col: cell.col}, true
	}

	// Pawn moves 2 cell
	if row := cell.row - 2*AdvDirs[side]; isValidPosition(cell.col, row) &&
		g.Board[row][cell.col] == pawn && g.Board[cell.row][cell.col] == Empty {
		return Position{row: row, col: cell.col}, true
	}

	// Pawn moves by en passant
	if row := cell.row - AdvDirs[side]; len(g.Moves) > 0 && isValidPosition(cell.col, row) &&
		g.Board[row][cell.col] == (Piece{'P', getOpponent(side)}) {
		prevMove := g.Moves[len(g.Moves)-1]
		preEpMove := Move{
			Source: Cell{g.Board[row][cell.col], Position{row: row + (AdvDirs[side] * 2), col: cell.col}},
			Target: Cell{Empty, Position{row: row, col: cell.col}},
			Action: Movement,
		}
		if prevMove == preEpMove {
			for _, offset := range PawnAtkDirs[side] {
				if adjCol := cell.col + offset[0]; isValidPosition(adjCol, row) &&
					g.Board[row][adjCol] == pawn {
					return Position{row: row, col: adjCol}, true
				}
			}
		}
	}

	return Position{}, false
}

func (g *Game) getAttackingLines(cell Position, side Side) []Position {
	res := make([]Position, 0, 2)

	checkDirections := func(dirs [][2]int, isAttacker func(Piece) bool) {
		for _, dir := range dirs {
			col, row := cell.col+dir[0], cell.row+dir[1]
			for ; isValidPosition(col, row); col, row = col+dir[0], row+dir[1] {
				p := g.Board[row][col]
				// Dir is blocked
				if p != Empty && p.side == getOpponent(side) {
					break
				}
				if isAttacker(p) {
					res = append(res, Position{row: row, col: col})
					if len(res) == 2 {
						return
					}
				}
			}
		}
	}

	// Lines
	isFullLineAttacker := func(p Piece) bool {
		return p == Piece{'R', side} ||
			p == Piece{'Q', side}
	}
	checkDirections(PicDirs['R'], isFullLineAttacker)
	if len(res) == 2 {
		return res
	}

	// Diagonals
	isFullDiagonalAttacker := func(p Piece) bool {
		return p == Piece{'B', side} ||
			p == Piece{'Q', side}
	}
	checkDirections(PicDirs['B'], isFullDiagonalAttacker)

	return res
}

func (g *Game) getAttackingKnights(cell Position, side Side) []Position {
	res := make([]Position, 0, 2)

	for _, move := range PicDirs['N'] {
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
	for _, move := range PicDirs['K'] {
		col, row := cell.col+move[0], cell.row+move[1]
		if isValidPosition(col, row) && g.Board[row][col] == (Piece{'K', side}) {
			return Position{row: row, col: col}, true
		}
	}

	return Position{}, false
}

func (g *Game) moveCell(source, target Cell) {
	g.Board[target.row][target.col] = g.Board[source.row][source.col]
	g.Board[source.row][source.col] = Empty

	if target.Piece != Empty {
		switch getOpponent(source.side) {
		case Black:
			g.blackCells--
		case White:
			g.whiteCells--
		}
	}

	if source.fig == 'K' {
		if source.side == White {
			g.whiteKing = Position{row: target.row, col: target.col}
		} else {
			g.blackKing = Position{row: target.row, col: target.col}
		}
	}
}

func (g *Game) cancelMoveCell(source, target Cell) {
	g.Board[target.row][target.col] = target.Piece
	g.Board[source.row][source.col] = source.Piece

	switch getOpponent(source.side) {
	case Black:
		g.blackCells++
	case White:
		g.whiteCells++
	}

	if source.fig == 'K' {
		if g.whoseTurn() == White {
			g.whiteKing = Position{row: source.row, col: source.col}
		} else {
			g.blackKing = Position{row: source.row, col: source.col}
		}
	}
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

func checkAtkDir(move Move) error {
	if move.Source.fig != 'P' {
		err := checkMoveDir(move)
		if err != nil {
			return fmt.Errorf("invalid attack")
		}
		return err
	}

	atkDir := [2]int{move.Target.col - move.Source.col, move.Target.row - move.Source.row}
	if !slices.Contains(PawnAtkDirs[move.Source.side], atkDir) {
		return fmt.Errorf("invalid attack")
	}

	return nil
}

func checkMoveDir(move Move) error {
	var dir [2]int
	switch move.Source.fig {
	case 'Q', 'B', 'R':
		drow, dcol := 0, 0

		if move.Target.row > move.Source.row {
			drow = 1
		} else if move.Target.row < move.Source.row {
			drow = -1
		}

		if move.Target.col > move.Source.col {
			dcol = 1
		} else if move.Target.col < move.Source.col {
			dcol = -1
		}
		dir = [2]int{dcol, drow}
	default:
		dir = [2]int{move.Target.col - move.Source.col, move.Target.row - move.Source.row}
	}

	dirs := append(PicDirs[move.Source.fig], PawnDirs[move.Source.side]...)
	if !slices.Contains(dirs, dir) {
		return fmt.Errorf("invalid move")
	}

	return nil
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

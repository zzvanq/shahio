package core

import (
	"testing"
)

func TestProcessMovement_CellIsOccupied(t *testing.T) {
	game := NewGame()

	move := Move{
		Source: Cell{Piece{'R', White}, Position{0, 0}},
		Target: Cell{Piece{'P', White}, Position{1, 0}},
		Action: Movement,
	}

	if err := game.processMovement(move); err.Error() != "cell is occupied" {
		t.Fatal(err)
	}
}

func TestProcessMovement_MoveIsBlocked(t *testing.T) {
	game := NewGame()

	// Block the pawn from moving 2 cells
	game.Board[2][0] = Piece{'P', Black}

	moves := []Move{
		{
			Source: Cell{Piece{'P', White}, Position{1, 0}},
			Target: Cell{Position: Position{3, 0}},
			Action: Movement,
		},
		{
			Source: Cell{Piece{'Q', Black}, Position{7, 3}},
			Target: Cell{Position: Position{6, 4}},
			Action: Movement,
		},
		{
			Source: Cell{Piece{'Q', White}, Position{0, 3}},
			Target: Cell{Position: Position{2, 3}},
			Action: Movement,
		},
	}

	for _, move := range moves {
		if err := game.processMovement(move); err.Error() != "move is blocked" {
			t.Fatalf("Error '%s' for move %v", err, move)
		}
	}
}

func TestProcessCapture_CellIsOccupied(t *testing.T) {
	game := NewGame()

	move := Move{
		Source: Cell{Piece{'R', White}, Position{0, 0}},
		Target: Cell{Piece{'P', White}, Position{1, 0}},
		Action: Movement,
	}

	if err := game.processMovement(move); err.Error() != "cell is occupied" {
		t.Fatal(err)
	}
}

func TestProcessCapture_CellIsEmpty(t *testing.T) {
	game := NewGame()

	move := Move{
		Source: Cell{Piece{'N', White}, Position{0, 1}},
		Target: Cell{Position: Position{2, 2}},
		Action: Capture,
	}

	if err := game.processCapture(move); err.Error() != "cell is empty" {
		t.Fatal(err)
	}
}

func TestProcessCapture_InvalidAttack(t *testing.T) {
	game := NewGame()

	move := Move{
		Source: Cell{Piece{'P', White}, Position{1, 0}},
		Target: Cell{Piece{'P', Black}, Position{6, 0}},
		Action: Capture,
	}

	if err := game.processCapture(move); err.Error() != "invalid attack" {
		t.Fatal(err)
	}
}

func TestProcessCapture(t *testing.T) {
	game := NewGame()

	capturesWithCells := []struct {
		move  Move
		enemy Cell
	}{
		{
			move: Move{
				Source: Cell{Piece{'P', White}, Position{1, 0}},
				Target: Cell{Piece{'P', Black}, Position{2, 1}},
				Action: Capture,
			},
			enemy: Cell{Piece{'P', Black}, Position{2, 1}},
		},
		{
			move: Move{
				Source: Cell{Piece{'P', White}, Position{1, 1}},
				Target: Cell{Piece{'P', Black}, Position{2, 0}},
				Action: Capture,
			},
			enemy: Cell{Piece{'P', Black}, Position{2, 0}},
		},
	}

	for _, cur := range capturesWithCells {
		// Set the enemy cell
		game.Board[cur.enemy.row][cur.enemy.col] = cur.enemy.Piece

		// Capture it
		if err := game.processCapture(cur.move); err != nil {
			t.Fatal(err)
		}
	}
}

func TestProcessEnpassant_CellIsNotEmpty(t *testing.T) {
	game := NewGame()

	// Block the e.p. move cell
	game.Board[2][1] = Piece{'P', Black}

	move := Move{
		Source: Cell{Piece{'P', White}, Position{1, 0}},
		Target: Cell{Position: Position{2, 1}},
		Action: Enpassant,
	}

	if err := game.processEnpassant(move); err.Error() != "cell is not empty" {
		t.Fatal(err)
	}
}

func TestProcessEnpassant_InvalidMove_NotAPawn(t *testing.T) {
	game := NewGame()

	move := Move{
		Source: Cell{Piece{'P', White}, Position{0, 1}},
		Target: Cell{Position: Position{2, 1}},
		Action: Enpassant,
	}

	if err := game.processEnpassant(move); err.Error() != "invalid move" {
		t.Fatal(err)
	}
}

func TestProcessEnpassant_InvalidMove_NoPreEnpassant(t *testing.T) {
	game := NewGame()

	// Set the enemy pawn
	game.Board[1][1] = Piece{'P', White}

	move := Move{
		Source: Cell{Piece{'P', Black}, Position{0, 1}},
		Target: Cell{Position: Position{2, 1}},
		Action: Enpassant,
	}

	if err := game.processEnpassant(move); err.Error() != "invalid move" {
		t.Fatal(err)
	}
}

func TestProcessEnpassant(t *testing.T) {
	game := NewGame()

	enemyPawn := Position{3, 0}
	// Add Pre-e.p. move
	preEpMove := Move{
		Source: Cell{Piece{'P', White}, Position{1, 0}},
		Target: Cell{Piece{'P', White}, enemyPawn},
		Action: Movement,
	}
	if err := game.processMove(preEpMove); err != nil {
		t.Fatal(err)
	}

	// Set Black pawn for it's e.p. attack
	game.Board[3][1] = Piece{'P', Black}

	epMove := Move{
		Source: Cell{Piece{'P', Black}, Position{3, 1}},
		Target: Cell{Position: Position{enemyPawn.row - 1, enemyPawn.col}}, // 1 cell below enemy pawn
		Action: Enpassant,
	}
	err := game.processEnpassant(epMove)
	if err != nil {
		print(err.Error())
	}

	// Check if white pre-e.p. pawn was captured
	if game.Board[enemyPawn.row][enemyPawn.col] != Empty {
		t.Fatalf("Enemy pawn wasn't captured during e.p. processing")
	}
}

func TestProcessPromotion_InvalidMove_WrongSide(t *testing.T) {
	game := NewGame()

	// Set a pawn for a promotion
	game.Board[6][0] = Piece{'P', White}

	move := Move{
		Source: Cell{Piece{'P', White}, Position{6, 0}},
		Target: Cell{Piece{'Q', Black}, Position{7, 1}},
		Action: Promotion,
	}

	if err := game.processPromotion(move); err.Error() != "invalid move" {
		t.Fatal(err)
	}
}
func TestProcessPromotion_InvalidMove_WrongPiece(t *testing.T) {
	game := NewGame()

	pic := Piece{'R', White}
	game.Board[6][0] = pic
	move := Move{
		Source: Cell{pic, Position{6, 0}},
		Target: Cell{Piece{'Q', White}, Position{7, 1}},
		Action: Promotion,
	}

	if err := game.processPromotion(move); err.Error() != "invalid move" {
		t.Fatal(err)
	}
}

func TestProcessPromotion(t *testing.T) {
	game := NewGame()

	game.Board[7][0] = Empty
	moves := []Move{
		{
			Source: Cell{Piece{'P', White}, Position{6, 0}},
			Target: Cell{Piece{'Q', White}, Position{7, 0}},
			Action: Promotion,
		},
		{
			Source: Cell{Piece{'P', White}, Position{6, 7}},
			Target: Cell{Piece{'Q', White}, Position{7, 6}},
			Action: Promotion,
		},
	}

	for _, move := range moves {
		game.processPromotion(move)

		if game.Board[move.Target.row][move.Target.col] != move.Target.Piece {
			t.Fatalf("Promotion error for move '%v'", move)
		}
	}
}

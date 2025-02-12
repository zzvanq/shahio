package core

import (
	"testing"
)

func TestProcessMovement_CellIsOccupied(t *testing.T) {
	game := NewGame()

	move := Move{
		Source: Cell{
			Piece{'R', White},
			Position{0, 0},
		},
		Target: Cell{
			Piece{'P', White},
			Position{1, 0},
		},
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
			Source: Cell{Piece{'Q', White}, Position{0, 3}},
			Target: Cell{Position: Position{2, 5}},
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
		Source: Cell{
			Piece{'R', White},
			Position{0, 0},
		},
		Target: Cell{
			Piece{'P', White},
			Position{1, 0},
		},
		Action: Movement,
	}

	if err := game.processMovement(move); err.Error() != "cell is occupied" {
		t.Fatal(err)
	}
}

func TestProcessCapture_CellIsEmpty(t *testing.T) {
	game := NewGame()

	move := Move{
		Source: Cell{
			Piece{'N', White},
			Position{0, 1},
		},
		Target: Cell{
			Position: Position{2, 2},
		},
		Action: Capture,
	}

	if err := game.processCapture(move); err.Error() != "cell is empty" {
		t.Fatal(err)
	}
}

func TestProcessCapture_InvalidAttack(t *testing.T) {
	game := NewGame()

	move := Move{
		Source: Cell{
			Piece{'P', White},
			Position{1, 0},
		},
		Target: Cell{
			Piece{'P', Black},
			Position{6, 0},
		},
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
				Source: Cell{
					Piece{'P', White},
					Position{1, 0},
				},
				Target: Cell{
					Piece{'P', Black},
					Position{2, 1},
				},
				Action: Capture,
			},
			enemy: Cell{Piece{'P', Black}, Position{2, 1}},
		},
		{
			move: Move{
				Source: Cell{
					Piece{'P', White},
					Position{1, 7},
				},
				Target: Cell{
					Piece{'P', Black},
					Position{2, 6},
				},
				Action: Capture,
			},
			enemy: Cell{Piece{'P', Black}, Position{2, 1}},
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
		Source: Cell{
			Piece{'P', White},
			Position{1, 0},
		},
		Target: Cell{
			Position: Position{2, 1},
		},
		Action: Enpassant,
	}

	if err := game.processEnpassant(move); err.Error() != "cell is not empty" {
		t.Fatal(err)
	}
}

func TestProcessEnpassant_InvalidMove_NotAPawn(t *testing.T) {
	game := NewGame()

	move := Move{
		Source: Cell{
			Piece{'P', White},
			Position{0, 1},
		},
		Target: Cell{
			Position: Position{2, 1},
		},
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
		Source: Cell{
			Piece{'P', Black},
			Position{0, 1},
		},
		Target: Cell{
			Position: Position{2, 1},
		},
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
		Source: Cell{
			Piece{'P', White},
			Position{1, 0},
		},
		Target: Cell{Piece{'P', White}, enemyPawn},
		Action: Movement,
	}
	game.processMove(preEpMove)

	// Set Black pawn for it's e.p. attack
	game.Board[3][1] = Piece{'P', Black}

	epMove := Move{
		Source: Cell{Piece{'P', Black}, Position{3, 1}},
		Target: Cell{Position: Position{enemyPawn.row - 1, enemyPawn.col}}, // 1 cell below enemy pawn
		Action: Enpassant,
	}
	game.processEnpassant(epMove)

	// Check if white pre-e.p. move was captured
	if game.Board[enemyPawn.row][enemyPawn.col] != Empty {
		t.Fatalf("Enemy pawn wasn't captured during e.p. processing")
	}
}

package core

import (
	"testing"
)

func TestProcessMovement_CellIsOccupied(t *testing.T) {
	game := NewGame()

	// Block the a3 cell
	game.Board[2][0] = Piece{'P', White}

	move := Move{
		Source: Cell{
			Piece{'P', White},
			Position{1, 0},
		},
		Target: Cell{
			Piece{'P', White},
			Position{2, 0},
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

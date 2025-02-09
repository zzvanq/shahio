package core

import (
	"testing"
)

func TestProcessCastlingQueenCastling(t *testing.T) {
	game := NewGame()
	// Empty between
	game.Board[0][3], game.Board[0][2], game.Board[0][1] = Empty, Empty, Empty

	move := Move{Action: QueenCastling}
	if err := game.processCastling(move); err != nil {
		t.Fatal(err)
	}

	king, rook := game.Board[0][2], game.Board[0][3]
	if king.fig != 'K' || rook.fig != 'R' {
		t.Fatal("incorrect castling")
	}
}

func TestProcessCastlingKingCastling(t *testing.T) {
	game := NewGame()
	// Empty between
	game.Board[0][5], game.Board[0][6] = Empty, Empty

	move := Move{Action: KingCastling}
	if err := game.processCastling(move); err != nil {
		t.Fatal(err)
	}

	king, rook := game.Board[0][6], game.Board[0][5]
	if king.fig != 'K' || rook.fig != 'R' {
		t.Fatal("incorrect castling")
	}
}

func TestProcessCastlingPiecesBetween(t *testing.T) {
	game := NewGame()

	move := Move{Action: KingCastling}

	expectedErr := "pieces between king and rook"
	if err := game.processCastling(move); err.Error() != expectedErr {
		t.Fatalf("Expected '%s', got '%v'", expectedErr, err)
	}
}

func TestProcessCastlingKingNotInPosition(t *testing.T) {
	game := NewGame()

	move := Move{Action: KingCastling}

	kingMove := Move{
		Source: Cell{
			Piece: Piece{'K', Black},
		},
	}
	game.Moves = append(game.Moves, kingMove)
	expectedErr := "king not in position"
	if err := game.processCastling(move); err.Error() != expectedErr {
		t.Fatalf("Expected '%s', got '%v'", expectedErr, err)
	}
}

func TestProcessCastlingRookNotInPositionWasMoved(t *testing.T) {
	game := NewGame()

	move := Move{Action: KingCastling}

	game.Moves = []Move{
		{
			Source: Cell{
				Piece:    Piece{'R', Black},
				Position: Position{row: 7, col: 7},
			},
			Action: Movement,
		},
	}
	expectedErr := "rook not in position"
	if err := game.processCastling(move); err.Error() != expectedErr {
		t.Fatalf("Expected '%s', got '%v'", expectedErr, err)
	}
}

func TestProcessCastlingRookNotInPositionWasCaptured(t *testing.T) {
	game := NewGame()

	move := Move{Action: KingCastling}

	game.Moves = []Move{
		{
			Target: Cell{
				Piece:    Piece{'R', Black},
				Position: Position{row: 7, col: 7},
			},
			Action: Capture,
		},
	}
	expectedErr := "rook not in position"
	if err := game.processCastling(move); err.Error() != expectedErr {
		t.Fatalf("Expected '%s', got '%v'", expectedErr, err)
	}
}

func TestProcessCastlingCrossoverCellsAttacked(t *testing.T) {
	game := NewGame()
	// Empty between
	game.Board[0][5], game.Board[0][6] = Empty, Empty
	// 1 cell below crossover cell
	game.Board[1][5] = Piece{'R', Black}

	move := Move{
		Action: KingCastling,
	}
	expectedErr := "crossover cell attacked"
	if err := game.processCastling(move); err.Error() != expectedErr {
		t.Fatalf("Expected '%s', got '%v'", expectedErr, err)
	}
}

func TestProcessCastlingKingIsInCheck(t *testing.T) {
	game := NewGame()
	// Empty between
	game.Board[0][5], game.Board[0][6] = Empty, Empty
	// 1 cell below king
	game.Board[1][4] = Piece{'R', Black}

	move := Move{
		Action: KingCastling,
	}
	expectedErr := "king is in check"
	if err := game.processCastling(move); err.Error() != expectedErr {
		t.Fatalf("Expected '%s', got '%v'", expectedErr, err)
	}
}

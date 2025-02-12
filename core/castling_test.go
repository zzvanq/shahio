package core

import (
	"testing"
)

func TestProcessCastling_QueenCastling(t *testing.T) {
	game := NewGame()
	// Empty between
	game.Board[0][3], game.Board[0][2], game.Board[0][1] = Empty, Empty, Empty

	move := Move{Action: QueenCastling}
	if err := game.processCastling(move); err != nil {
		t.Fatal(err)
	}

	king, rook := game.Board[0][2], game.Board[0][3]
	if king.fig != 'K' || rook.fig != 'R' {
		t.Fatalf("incorrect castling")
	}
}

func TestProcessCastling_KingCastling(t *testing.T) {
	game := NewGame()
	// Empty between
	game.Board[0][5], game.Board[0][6] = Empty, Empty

	move := Move{Action: KingCastling}
	if err := game.processCastling(move); err != nil {
		t.Fatal(err)
	}

	king, rook := game.Board[0][6], game.Board[0][5]
	if king.fig != 'K' || rook.fig != 'R' {
		t.Fatalf("incorrect castling")
	}
}

func TestProcessCastling_PiecesBetween(t *testing.T) {
	game := NewGame()

	move := Move{Action: KingCastling}

	if err := game.processCastling(move); err.Error() != "pieces between king and rook" {
		t.Fatal(err)
	}
}

func TestProcessCastling_KingNotInPosition(t *testing.T) {
	game := NewGame()

	move := Move{Action: KingCastling}

	kingMove := Move{
		Source: Cell{
			Piece: Piece{'K', Black},
		},
	}
	game.Moves = append(game.Moves, kingMove)
	if err := game.processCastling(move); err.Error() != "king not in position" {
		t.Fatal(err)
	}
}

func TestProcessCastling_RookNotInPositionWasMoved(t *testing.T) {
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
	if err := game.processCastling(move); err.Error() != "rook not in position" {
		t.Fatal(err)
	}
}

func TestProcessCastling_RookNotInPositionWasCaptured(t *testing.T) {
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
	if err := game.processCastling(move); err.Error() != "rook not in position" {
		t.Fatal(err)
	}
}

func TestProcessCastling_CrossoverCellsAttacked(t *testing.T) {
	game := NewGame()
	// Empty between
	game.Board[0][5], game.Board[0][6] = Empty, Empty
	// 1 cell below crossover cell
	game.Board[1][5] = Piece{'R', Black}

	move := Move{
		Action: KingCastling,
	}
	if err := game.processCastling(move); err.Error() != "crossover cell attacked" {
		t.Fatal(err)
	}
}

func TestProcessCastling_KingIsInCheck(t *testing.T) {
	game := NewGame()
	// Empty between
	game.Board[0][5], game.Board[0][6] = Empty, Empty
	// 1 cell below king
	game.Board[1][4] = Piece{'R', Black}

	move := Move{
		Action: KingCastling,
	}
	if err := game.processCastling(move); err.Error() != "king is in check" {
		t.Fatal(err)
	}
}

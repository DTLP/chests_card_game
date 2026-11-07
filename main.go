package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

const (
	logMessageLimit = 30
	layoutWidth     = 115
)

var stdin = bufio.NewScanner(os.Stdin)

// readInput reads a single line from standard input and trims surrounding
// whitespace.
func readInput() string {
	if !stdin.Scan() {
		fmt.Println("\nInput closed. Exiting.")
		os.Exit(0)
	}
	return strings.TrimSpace(stdin.Text())
}

// isAllowedRank reports whether rank is one of the game's valid ranks.
func isAllowedRank(rank string) bool {
	for _, r := range ranks {
		if r == rank {
			return true
		}
	}
	return false
}

// promptPlayerChoice asks the human player which opponent to talk to and
// repeats the prompt until they enter 2, 3 or 4.
func promptPlayerChoice() int {
	for {
		fmt.Print("Which player do you want to talk to? (2, 3 or 4): ")
		choice, err := strconv.Atoi(readInput())
		if err == nil && choice >= 2 && choice <= 4 {
			return choice
		}
		fmt.Println("Invalid player choice. Please enter 2, 3 or 4.")
	}
}

// promptRank asks the human player for a card rank and repeats the prompt
// until they enter a valid rank.
func promptRank() string {
	for {
		fmt.Print("Enter the card rank you want to ask about (A, 2, 3, ..., K): ")
		rank := strings.ToUpper(readInput())
		if isAllowedRank(rank) {
			return rank
		}
		fmt.Println("Invalid rank. Please enter one of: A, 2-10, J, Q or K.")
	}
}

// promptCount asks the human player how many cards of rank the target player
// has and repeats the prompt until they enter a number between 1 and 4.
func promptCount(rank string, targetID int) int {
	for {
		fmt.Printf("How many %s cards do you think Player %d has? ", rank, targetID)
		count, err := strconv.Atoi(readInput())
		if err == nil && count >= 1 && count <= 4 {
			return count
		}
		fmt.Println("Invalid count. Please enter a number between 1 and 4.")
	}
}

// clearScreen clears the terminal screen.
func clearScreen() {
	fmt.Print("\033[H\033[2J") // ANSI escape code to clear screen and move cursor to top
}

func appendMessageLog(log *[]string, message string) []string {
	if len(*log) >= logMessageLimit {
		*log = (*log)[1:]
	}

	*log = append(*log, message)

	return *log
}

// printLayout prints the updated game layout after each turn.
func printLayout(p *players, messageLog []string) {
	clearScreen()

	// Print score board (top)
	fmt.Println("Score board")
	fmt.Printf("Player 1: %d\n", p.player[0].score)
	fmt.Printf("Player 2: %d\n", p.player[1].score)
	fmt.Printf("Player 3: %d\n", p.player[2].score)
	fmt.Printf("Player 4: %d\n", p.player[3].score)
	fmt.Println(strings.Repeat("=", layoutWidth))

	// Print chat messages (middle)
	for _, message := range messageLog {
		fmt.Println(message)
	}

	fmt.Println(strings.Repeat("=", layoutWidth))
}

// isGameOver checks if the sum of all players' scores equals 13, indicating the game is over.
func isGameOver(player1Score, player2Score, player3Score, player4Score int) bool {
	totalScore := player1Score + player2Score + player3Score + player4Score
	return totalScore == 13
}

// main game loop
func main() {
	// Generate and shuffle the deck
	deck := generateDeck()

	// var messageLog []string
	messageLog := make([]string, logMessageLimit)

	// Distribute cards to players
	hand1, hand2, hand3, hand4 := distributeCards(deck)

	// Create an instance of players with 4 players
	p := players{
		player: []player{
			{id: 1, hand: hand1, score: 0},
			{id: 2, hand: hand2, score: 0},
			{id: 3, hand: hand3, score: 0},
			{id: 4, hand: hand4, score: 0},
		},
	}

	// Game loop
	round := 0
	for {
		round++
		if round > 1 {
			appendMessageLog(&messageLog, strings.Repeat("-", layoutWidth))
			appendMessageLog(&messageLog, fmt.Sprintf("Round %d", round))
			appendMessageLog(&messageLog, strings.Repeat("-", layoutWidth))
		}

		// Players taking turns
		for i := range p.player {

			currentPlayer := &p.player[i]
			appendMessageLog(&messageLog, fmt.Sprintf("Game Master: Player %d's turn", currentPlayer.id))

			printLayout(&p, messageLog)

			var selectedPlayer *player
			var rankGuess string
			var countGuess int

			// Handle turn logic based on player
			switch currentPlayer.id {
			case 1:
				// Player 1's turn - human player

				// Print each player's hand
				printHand("Your hand", p.player[0].hand) // Human player

				playerChoice := promptPlayerChoice()
				selectedPlayer = &p.player[playerChoice-1]

				rankGuess = promptRank()
				countGuess = promptCount(rankGuess, selectedPlayer.id)
			default:
				// Bot randomly chooses which player to talk to
				selectedPlayerIndex := selectAnotherRandomPlayer(currentPlayer.id)
				selectedPlayer = &p.player[selectedPlayerIndex]

				// Bot randomly chooses a rank to ask about
				rankGuess = ranks[rand.Intn(len(ranks))]

				// Bot randomly guesses the count
				countGuess = rand.Intn(3) + 1 // Guess between 1 and 3
			}

			actualCount := countCardsOfRank(selectedPlayer.hand, rankGuess)
			appendMessageLog(&messageLog, fmt.Sprintf("   Player %d: Player %d, do you have any cards of rank %s?", currentPlayer.id, selectedPlayer.id, rankGuess))

			if actualCount == 0 {
				appendMessageLog(&messageLog, fmt.Sprintf("   Player %d: No", selectedPlayer.id))
				continue
			}
			appendMessageLog(&messageLog, fmt.Sprintf("   Player %d: Yes", selectedPlayer.id))

			appendMessageLog(&messageLog, fmt.Sprintf("   Player %d: Player %d, do you have %d of these cards?", currentPlayer.id, selectedPlayer.id, countGuess))
			if countGuess != actualCount {
				appendMessageLog(&messageLog, fmt.Sprintf("   Player %d: No", selectedPlayer.id))
				continue
			}

			appendMessageLog(&messageLog, fmt.Sprintf("   Player %d: Yes", selectedPlayer.id))

			appendMessageLog(&messageLog, fmt.Sprintf("Game Master: Player %d takes %d cards of rank %s from Player %d", currentPlayer.id, actualCount, rankGuess, selectedPlayer.id))
			moveCards(&selectedPlayer.hand, &currentPlayer.hand, rankGuess)

			if checkForCompleteSet(&currentPlayer.hand, rankGuess) {
				currentPlayer.score++
				appendMessageLog(&messageLog, fmt.Sprintf("Game Master: Player %d has collected all 4 cards of rank %s and scored 1 point!", currentPlayer.id, rankGuess))
			}
		}

		// Check if the game is over
		if isGameOver(p.player[0].score, p.player[1].score, p.player[2].score, p.player[3].score) {
			break
		}
	}

	// Display final score
	fmt.Println("Game over!\n\nFinal Scores:")
	fmt.Printf("Player 1: %d\n", p.player[0].score)
	fmt.Printf("Player 2: %d\n", p.player[1].score)
	fmt.Printf("Player 3: %d\n", p.player[2].score)
	fmt.Printf("Player 4: %d\n", p.player[3].score)
}

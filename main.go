package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	logMessageLimit = 30
	logViewHeight   = 18
	layoutWidth     = 115
	winningScore    = 5
	typeDelay       = 15 * time.Millisecond
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

// gameLog keeps the recent game messages and tracks how many have been shown.
type gameLog struct {
	messages []string
	shown    int
}

// add appends a message to the log, dropping the oldest when the log is full.
func (l *gameLog) add(message string) {
	if len(l.messages) >= logMessageLimit {
		l.messages = l.messages[1:]
		if l.shown > 0 {
			l.shown--
		}
	}
	l.messages = append(l.messages, message)
}

// view returns the messages currently visible in the log pane and how many of
// them have already been shown. Only the most recent logViewHeight messages
// are visible, so older messages scroll off the top.
func (l *gameLog) view() (window []string, shown int) {
	window = l.messages
	if len(window) > logViewHeight {
		window = window[len(window)-logViewHeight:]
	}
	shown = l.shown - (len(l.messages) - len(window))
	if shown < 0 {
		shown = 0
	}
	if shown > len(window) {
		shown = len(window)
	}
	return window, shown
}

// typewrite prints a message one character at a time at a readable pace.
func typewrite(message string) {
	for _, r := range message {
		fmt.Printf("%c", r)
		time.Sleep(typeDelay)
	}
	fmt.Println()
}

// moveCursorUp moves the cursor up n lines.
func moveCursorUp(n int) {
	if n > 0 {
		fmt.Printf("\033[%dA", n)
	}
}

// moveCursorDown moves the cursor down n lines.
func moveCursorDown(n int) {
	if n > 0 {
		fmt.Printf("\033[%dB", n)
	}
}

// handHeight returns the number of lines printHand outputs for a hand.
func handHeight(cards [][]string) int {
	blocks := (len(cards) + 12) / 13
	height := 1 + blocks*5 + 1
	if len(cards) > 13 {
		height++
	}
	return height
}

// printLayout prints the updated game layout after each turn. The score and
// the player's hand stay on screen while only the new log messages are typed
// out in the log area between the separators.
func printLayout(p *players, log *gameLog, hand [][]string) {
	clearScreen()

	// Print score board (top) on a single line.
	fmt.Printf("Score - Player 1: %d | Player 2: %d | Player 3: %d | Player 4: %d\n",
		p.player[0].score, p.player[1].score, p.player[2].score, p.player[3].score)
	fmt.Println(strings.Repeat("=", layoutWidth))

	// Show only the most recent messages so old lines scroll off the top.
	window, shown := log.view()
	newMessages := window[shown:]

	// Print already-seen messages instantly and reserve a line for each new
	// message so the hand below stays in place.
	for _, message := range window[:shown] {
		fmt.Println(message)
	}
	for range newMessages {
		fmt.Println()
	}

	fmt.Println(strings.Repeat("=", layoutWidth))

	// Print the player's hand (bottom), which remains visible throughout.
	printHand("Your hand", hand)

	// Type the new messages into their reserved lines, then return the cursor
	// below the hand.
	moveCursorUp(len(window) + 1 + handHeight(hand) - shown)
	for _, message := range newMessages {
		typewrite(message)
	}
	log.shown = len(log.messages)
	moveCursorDown(1 + handHeight(hand))
}

// isGameOver checks if a player has collected winningScore chests or all 13
// sets have been collected, indicating the game is over.
func isGameOver(scores ...int) bool {
	totalScore := 0
	for _, score := range scores {
		if score >= winningScore {
			return true
		}
		totalScore += score
	}
	return totalScore == 13
}

// awardCompleteSets scores and removes any complete sets a player already
// holds, such as ones dealt at the start of the game.
func awardCompleteSets(p *player, log *gameLog) {
	for _, rank := range ranks {
		if checkForCompleteSet(&p.hand, rank) {
			p.score++
			log.add(fmt.Sprintf("Game Master: Player %d was dealt all 4 cards of rank %s and scored 1 point!", p.id, rank))
		}
	}
}

// main game loop
func main() {
	// Generate and shuffle the deck
	deck := generateDeck()

	log := &gameLog{}

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

	// Score any complete sets players were dealt up front.
	for i := range p.player {
		awardCompleteSets(&p.player[i], log)
	}

	// Game loop
	round := 0
	gameOver := false
	for !gameOver {
		round++
		if round > 1 {
			log.add(strings.Repeat("-", layoutWidth))
		}

		// Players taking turns
		for i := 0; i < len(p.player); i++ {

			currentPlayer := &p.player[i]
			log.add(fmt.Sprintf("Game Master: Player %d's turn", currentPlayer.id))

			printLayout(&p, log, p.player[0].hand)

			var selectedPlayer *player
			var rankGuess string
			var countGuess int

			// Handle turn logic based on player
			switch currentPlayer.id {
			case 1:
				// Player 1's turn - human player

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
			log.add(fmt.Sprintf("   Player %d: Player %d, do you have any cards of rank %s?", currentPlayer.id, selectedPlayer.id, rankGuess))

			if actualCount == 0 {
				log.add(fmt.Sprintf("   Player %d: No", selectedPlayer.id))
				continue
			}
			log.add(fmt.Sprintf("   Player %d: Yes", selectedPlayer.id))

			log.add(fmt.Sprintf("   Player %d: Player %d, do you have %d of these cards?", currentPlayer.id, selectedPlayer.id, countGuess))
			if countGuess != actualCount {
				log.add(fmt.Sprintf("   Player %d: No", selectedPlayer.id))
				continue
			}

			log.add(fmt.Sprintf("   Player %d: Yes", selectedPlayer.id))

			log.add(fmt.Sprintf("Game Master: Player %d takes %d cards of rank %s from Player %d", currentPlayer.id, actualCount, rankGuess, selectedPlayer.id))
			moveCards(&selectedPlayer.hand, &currentPlayer.hand, rankGuess)

			if checkForCompleteSet(&currentPlayer.hand, rankGuess) {
				currentPlayer.score++
				log.add(fmt.Sprintf("Game Master: Player %d has collected all 4 cards of rank %s and scored 1 point!", currentPlayer.id, rankGuess))
			}

			if isGameOver(p.player[0].score, p.player[1].score, p.player[2].score, p.player[3].score) {
				gameOver = true
				break
			}

			log.add(fmt.Sprintf("Game Master: Player %d guessed correctly and takes another turn", currentPlayer.id))
			i-- // Same player takes another turn
		}
	}

	// Render any messages added on the final turn.
	printLayout(&p, log, p.player[0].hand)

	// Display final score
	fmt.Println("Game over!\n\nFinal Scores:")
	fmt.Printf("Player 1: %d\n", p.player[0].score)
	fmt.Printf("Player 2: %d\n", p.player[1].score)
	fmt.Printf("Player 3: %d\n", p.player[2].score)
	fmt.Printf("Player 4: %d\n", p.player[3].score)

	maxScore := 0
	for _, pl := range p.player {
		if pl.score > maxScore {
			maxScore = pl.score
		}
	}
	fmt.Printf("\nPlayer(s) with %d chests win:\n", maxScore)
	for _, pl := range p.player {
		if pl.score == maxScore {
			fmt.Printf("Player %d\n", pl.id)
		}
	}
}

package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

const (
	reset   = "\033[0m"  // Reset color
	whiteBg = "\033[47m" // White background
	blackFg = "\033[30m" // Black foreground (text)
	redFg   = "\033[31m" // Red foreground (text)
)

// Card represents a playing card with a rank, suit, and color
type Card struct {
	rank  string
	suit  string
	color string
}

// generateDeck creates a shuffled deck of 52 cards
func generateDeck() []Card {
	var deck []Card
	for _, suit := range suits {
		for _, rank := range ranks {
			color := blackFg
			if suit == "❤" || suit == "♦" {
				color = redFg
			}
			deck = append(deck, Card{rank: rank, suit: suit, color: color})
		}
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
	return deck
}

// generateCard creates a string representation of a card
func generateCard(card Card) []string {
	color := blackFg
	if card.suit == "❤" || card.suit == "♦" {
		color = redFg
	}

	// Card representation as a slice of strings
	return []string{
		fmt.Sprintf("%s%s%-2s     ", whiteBg, color, card.rank), // Rank in top-left corner
		"       ",
		fmt.Sprintf("   %s%s   ", color, card.suit),
		"       ",
		fmt.Sprintf("     %s%s%2s", whiteBg, color, card.rank),
	}
}

// distributeCards distributes 13 cards to each player from the shuffled deck
func distributeCards(deck []Card) ([][]string, [][]string, [][]string, [][]string) {
	var player1, player2, player3, player4 [][]string
	for i := 0; i < 13; i++ {
		player1 = append(player1, generateCard(deck[i]))
		player2 = append(player2, generateCard(deck[i+13]))
		player3 = append(player3, generateCard(deck[i+26]))
		player4 = append(player4, generateCard(deck[i+39]))
	}
	return player1, player2, player3, player4
}

// Available suits and ranks
var (
	suits = []string{"♠", "❤", "♦", "♣"}
	ranks = []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
)

// drawCardRow creates a row for each card line in a horizontal format
func drawCardRow(cards [][]string, row int) string {
	rowString := ""
	for _, card := range cards {
		rowString += whiteBg + card[row] + reset + "  " // Add space between cards
	}
	return rowString
}

// printHand displays a hand of cards horizontally
// If the hand contains more than 13 cards, it will be printed in multiple lines with a blank line between the two parts
func printHand(player string, cards [][]string) {
	fmt.Println(player)
	numCards := len(cards)
	numLines := (numCards + 12) / 13 // Determine how many lines to print

	// Print first 13 cards
	for line := 0; line < numLines; line++ {
		start := line * 13
		end := (line + 1) * 13
		if end > numCards {
			end = numCards
		}
		printCardsInLine := cards[start:end]                  // Slice the hand for the current line
		for row := 0; row < len(printCardsInLine[0]); row++ { // Draw each row across the cards
			fmt.Println(drawCardRow(printCardsInLine, row))
		}
		if line == 0 && numCards > 13 {
			// Add a blank line after the first 13 cards
			fmt.Println()
		}
	}
	fmt.Println()
}

// ExtractRank extracts the rank of the card without color codes or extra spaces
func ExtractRank(card string) string {
	re := regexp.MustCompile(`\033\[[0-9;]*m`)                   // This regex matches any ANSI escape sequence
	cardRank := re.ReplaceAllString(strings.TrimSpace(card), "") // Remove color codes and spaces
	return cardRank
}

// countCardsOfRank counts how many cards of a given rank a player has
func countCardsOfRank(cards [][]string, rank string) int {
	count := 0
	for _, card := range cards {
		// Extract the rank from card[0] without affecting the card itself
		cardRank := ExtractRank(card[0])
		// Directly compare the extracted card rank to the requested rank
		if cardRank == rank {
			count++
		}
	}
	return count
}

// moveCards moves cards of a specific rank from one player's hand to another player's hand
func moveCards(from *[][]string, to *[][]string, rank string) {
	// Find and remove all cards of the given rank from the 'from' player
	var cardsToMove [][]string
	var remainingCards [][]string

	// Loop over the hand and separate the cards of the specific rank
	for _, card := range *from {
		// Extract the rank from card[0]
		cardRank := ExtractRank(card[0]) // Extract rank without formatting issues
		// If the card matches the desired rank, move it to the "cardsToMove" list
		if cardRank == rank {
			cardsToMove = append(cardsToMove, card)
		} else {
			remainingCards = append(remainingCards, card)
		}
	}
	// Update the 'from' player hand by keeping only the remaining cards
	*from = remainingCards
	// Add the found cards to the 'to' player's hand
	*to = append(*to, cardsToMove...)
}

// checkForCompleteSet checks if a player has collected all 4 cards of a rank
func checkForCompleteSet(player *[][]string, rank string) bool {
	count := countCardsOfRank(*player, rank)
	if count == 4 {
		// Remove all 4 cards of the rank from the player's hand
		var remainingCards [][]string
		for _, card := range *player {
			if ExtractRank(card[0]) != rank {
				remainingCards = append(remainingCards, card)
			}
		}
		*player = remainingCards
		return true
	}
	return false
}

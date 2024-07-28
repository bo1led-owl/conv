package main

// A simple example that shows how to retrieve a value from a Bubble Tea
// program after the Bubble Tea has exited.

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
)

type radix int

const (
	Binary radix = iota
	Octal
	Decimal
	Hexadecimal
)

type errMsg struct {
	msg string
}

func (e errMsg) Error() string {
	return e.msg
}

type model struct {
	input     [4]string
	mode      radix
	cursor    cursor.Model
	cursorPos int
}

func initialModel() model {
	c := cursor.New()
	c.SetChar("0")
	cursor.Blink()
	c.Focus()

	return model{
		input:     [4]string{"", "", "", ""},
		mode:      Decimal,
		cursor:    c,
		cursorPos: 0,
	}
}

func (m model) Init() tea.Cmd {
	return cursor.Blink
}

func clamp[T int | radix](v, low, high T) T {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}

func min[T int | radix](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func max[T int | radix](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func (m *model) updateCursor(newPos int) {
	m.cursorPos = clamp(newPos, 0, len(m.input[m.mode]))
	if m.cursorPos < len(m.input[m.mode]) {
		m.cursor.SetChar(string(m.input[m.mode][m.cursorPos]))
	} else {
		if len(m.input[m.mode]) == 0 {
			m.cursor.SetChar("0")
		} else {
			m.cursor.SetChar(" ")
		}
	}
}

func (m *model) updateInput() {
	var i uint64
	switch m.mode {
	case Binary:
		i = parseInt(m.input[m.mode], 2)
	case Octal:
		i = parseInt(m.input[m.mode], 8)
	case Decimal:
		i = parseInt(m.input[m.mode], 10)
	case Hexadecimal:
		i = parseInt(m.input[m.mode], 16)
	}

	if i == 0 {
		for mode := Binary; mode <= Hexadecimal; mode++ {
			m.input[mode] = ""
		}
		return
	} else {
		for mode := Binary; mode <= Hexadecimal; mode++ {
			switch mode {
			case Binary:
				m.input[mode] = fmt.Sprintf("%b", i)
			case Octal:
				m.input[mode] = fmt.Sprintf("%o", i)
			case Decimal:
				m.input[mode] = fmt.Sprintf("%d", i)
			case Hexadecimal:
				m.input[mode] = strings.ToUpper(fmt.Sprintf("%x", i))
			}
		}
	}
}

func isValidDigit(c rune, r radix) bool {
	switch r {
	case Binary:
		return c == '0' || c == '1'
	case Octal:
		return '0' <= c && c <= '7'
	case Decimal:
		return '0' <= c && c <= '9'
	case Hexadecimal:
		return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f')
	}

	return false
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	oldPos := m.cursorPos
	oldMode := m.mode

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		if len(key) == 1 && isValidDigit(unicode.ToLower(rune(key[0])), m.mode) {
			if key[0] == '0' && m.cursorPos == 0 {
				break
			}
			m.input[m.mode] = m.input[m.mode][:m.cursorPos] + key + m.input[m.mode][m.cursorPos:]
			m.updateInput()
			m.updateCursor(m.cursorPos + 1)
		} else {
			switch key {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "left", "h":
				if m.cursorPos > 0 {
					m.updateCursor(m.cursorPos - 1)
				}
			case "right", "l":
				if m.cursorPos < len(m.input[m.mode]) {
					m.updateCursor(m.cursorPos + 1)
				}
			case "up", "k":
				m.mode = clamp(m.mode-1, Binary, Hexadecimal)
				m.updateCursor(m.cursorPos)
			case "down", "j":
				m.mode = clamp(m.mode+1, Binary, Hexadecimal)
				m.updateCursor(m.cursorPos)
			case "backspace":
				if m.cursorPos > 0 {
					newPos := m.cursorPos - 1
					newInput := m.input[m.mode][:newPos]
					if m.cursorPos < len(m.input[m.mode]) {
						newInput += m.input[m.mode][m.cursorPos:]
					}

					m.input[m.mode] = newInput
					m.updateCursor(m.cursorPos - 1)
					m.updateInput()
				}
			}
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.cursor, cmd = m.cursor.Update(msg)
	cmds = append(cmds, cmd)

	if (oldMode != m.mode || oldPos != m.cursorPos) && m.cursor.Mode() == cursor.CursorBlink {
		m.cursor.Blink = false
		cmds = append(cmds, m.cursor.BlinkCmd())
	}

	return m, tea.Batch(cmds...)
}

func parseInt(s string, base int) uint64 {
	if len(s) == 0 {
		return 0
	}

	i, err := strconv.ParseUint(s, base, 64)

	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(2)
	}
	return i
}

func formatMode(mode radix) string {
	switch mode {
	case Binary:
		return "bin"
	case Octal:
		return "oct"
	case Decimal:
		return "dec"
	case Hexadecimal:
		return "hex"
	}

	return ""
}

func (m model) View() string {
	b := strings.Builder{}

	for r := Binary; r <= Hexadecimal; r++ {
		if r != m.mode {
			var view string
			if len(m.input[r]) == 0 {
				view = "0"
			} else {
				view = m.input[r]
			}

			b.WriteString(fmt.Sprintf("%s: %s\n", formatMode(r), view))
		} else {
			view := m.input[r][:m.cursorPos] + m.cursor.View()

			if m.cursorPos < len(m.input[r]) {
				view += m.input[r][m.cursorPos+1:]
			}

			b.WriteString(fmt.Sprintf("%s: %s\n", formatMode(r), view))

		}
	}

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error occured: %v", err)
		os.Exit(1)
	}
}

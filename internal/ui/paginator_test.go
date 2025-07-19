package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewPaginator(t *testing.T) {
	tests := []struct {
		name       string
		items      []string
		pageSize   int
		wantPages  int
	}{
		{
			name:      "empty items",
			items:     []string{},
			pageSize:  10,
			wantPages: 1,
		},
		{
			name:      "single page",
			items:     []string{"item1", "item2", "item3"},
			pageSize:  10,
			wantPages: 1,
		},
		{
			name:      "exact pages",
			items:     []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
			pageSize:  5,
			wantPages: 2,
		},
		{
			name:      "partial last page",
			items:     []string{"1", "2", "3", "4", "5", "6", "7"},
			pageSize:  3,
			wantPages: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPaginator(tt.items, tt.pageSize)
			
			if p.TotalPages != tt.wantPages {
				t.Errorf("TotalPages = %d, want %d", p.TotalPages, tt.wantPages)
			}
			
			if p.CurrentPage != 1 {
				t.Errorf("CurrentPage = %d, want 1", p.CurrentPage)
			}
			
			if p.PageSize != tt.pageSize {
				t.Errorf("PageSize = %d, want %d", p.PageSize, tt.pageSize)
			}
		})
	}
}

func TestPaginatorNavigation(t *testing.T) {
	items := make([]string, 25)
	for i := 0; i < 25; i++ {
		items[i] = fmt.Sprintf("Item %d", i+1)
	}
	
	p := NewPaginator(items, 10)
	
	tests := []struct {
		name        string
		keyMsg      string
		wantPage    int
		description string
	}{
		{
			name:        "right arrow increases page",
			keyMsg:      "right",
			wantPage:    2,
			description: "should go to page 2",
		},
		{
			name:        "left arrow on page 1 stays",
			keyMsg:      "left",
			wantPage:    1,
			description: "should stay on page 1",
		},
		{
			name:        "h key goes left",
			keyMsg:      "h",
			wantPage:    1,
			description: "vim style navigation",
		},
		{
			name:        "l key goes right",
			keyMsg:      "l",
			wantPage:    2,
			description: "vim style navigation",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to page 1
			p.CurrentPage = 1
			
			// Navigate to page 2 if needed for left arrow test
			if tt.keyMsg == "left" || tt.keyMsg == "h" {
				p.CurrentPage = 2
			}
			
			// Send key message
			model, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.keyMsg)})
			updatedP := model.(PaginatorModel)
			
			if updatedP.CurrentPage != tt.wantPage {
				t.Errorf("After %s: CurrentPage = %d, want %d", tt.keyMsg, updatedP.CurrentPage, tt.wantPage)
			}
		})
	}
}

func TestPaginatorBoundaries(t *testing.T) {
	items := []string{"1", "2", "3", "4", "5"}
	p := NewPaginator(items, 2) // 3 pages total
	
	// Test can't go past last page
	p.CurrentPage = 3
	model, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("right")})
	updatedP := model.(PaginatorModel)
	
	if updatedP.CurrentPage != 3 {
		t.Errorf("Should stay on last page: CurrentPage = %d, want 3", updatedP.CurrentPage)
	}
	
	// Test can't go before first page
	p.CurrentPage = 1
	model, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("left")})
	updatedP = model.(PaginatorModel)
	
	if updatedP.CurrentPage != 1 {
		t.Errorf("Should stay on first page: CurrentPage = %d, want 1", updatedP.CurrentPage)
	}
}

func TestPaginatorView(t *testing.T) {
	tests := []struct {
		name         string
		items        []string
		pageSize     int
		currentPage  int
		wantContains []string
	}{
		{
			name:         "empty items",
			items:        []string{},
			pageSize:     10,
			currentPage:  1,
			wantContains: []string{"No items to display"},
		},
		{
			name:         "first page",
			items:        []string{"Apple", "Banana", "Cherry", "Date", "Elderberry"},
			pageSize:     3,
			currentPage:  1,
			wantContains: []string{"1. Apple", "2. Banana", "3. Cherry", "Page 1 of 2"},
		},
		{
			name:         "second page",
			items:        []string{"Apple", "Banana", "Cherry", "Date", "Elderberry"},
			pageSize:     3,
			currentPage:  2,
			wantContains: []string{"4. Date", "5. Elderberry", "Page 2 of 2"},
		},
		{
			name:         "navigation help shown for multiple pages",
			items:        []string{"1", "2", "3", "4", "5"},
			pageSize:     2,
			currentPage:  1,
			wantContains: []string{"← →: navigate", "q: quit"},
		},
		{
			name:         "no navigation for single page",
			items:        []string{"1", "2"},
			pageSize:     10,
			currentPage:  1,
			wantContains: []string{"q: quit"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPaginator(tt.items, tt.pageSize)
			p.CurrentPage = tt.currentPage
			
			view := p.View()
			
			for _, want := range tt.wantContains {
				if !strings.Contains(view, want) {
					t.Errorf("View missing %q\nGot: %s", want, view)
				}
			}
		})
	}
}

func TestGetCurrentItems(t *testing.T) {
	items := []string{"A", "B", "C", "D", "E", "F", "G"}
	p := NewPaginator(items, 3)
	
	tests := []struct {
		page      int
		wantItems []string
	}{
		{
			page:      1,
			wantItems: []string{"A", "B", "C"},
		},
		{
			page:      2,
			wantItems: []string{"D", "E", "F"},
		},
		{
			page:      3,
			wantItems: []string{"G"},
		},
	}
	
	for _, tt := range tests {
		t.Run(fmt.Sprintf("page_%d", tt.page), func(t *testing.T) {
			p.CurrentPage = tt.page
			got := p.GetCurrentItems()
			
			if len(got) != len(tt.wantItems) {
				t.Errorf("GetCurrentItems() returned %d items, want %d", len(got), len(tt.wantItems))
			}
			
			for i, item := range got {
				if item != tt.wantItems[i] {
					t.Errorf("Item[%d] = %s, want %s", i, item, tt.wantItems[i])
				}
			}
		})
	}
}

func TestGetCurrentRange(t *testing.T) {
	items := []string{"A", "B", "C", "D", "E", "F", "G"}
	p := NewPaginator(items, 3)
	
	tests := []struct {
		page      int
		wantStart int
		wantEnd   int
	}{
		{
			page:      1,
			wantStart: 0,
			wantEnd:   3,
		},
		{
			page:      2,
			wantStart: 3,
			wantEnd:   6,
		},
		{
			page:      3,
			wantStart: 6,
			wantEnd:   7,
		},
	}
	
	for _, tt := range tests {
		t.Run(fmt.Sprintf("page_%d", tt.page), func(t *testing.T) {
			p.CurrentPage = tt.page
			start, end := p.GetCurrentRange()
			
			if start != tt.wantStart {
				t.Errorf("Start = %d, want %d", start, tt.wantStart)
			}
			if end != tt.wantEnd {
				t.Errorf("End = %d, want %d", end, tt.wantEnd)
			}
		})
	}
}

func TestWindowResize(t *testing.T) {
	p := NewPaginator([]string{"item"}, 10)
	
	// Test window resize message
	model, _ := p.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	updatedP := model.(PaginatorModel)
	
	if updatedP.Width != 120 {
		t.Errorf("Width = %d, want 120", updatedP.Width)
	}
	if updatedP.Height != 40 {
		t.Errorf("Height = %d, want 40", updatedP.Height)
	}
}

func TestQuitCommands(t *testing.T) {
	p := NewPaginator([]string{"item"}, 10)
	
	// Test q key
	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Error("Expected quit command for 'q' key")
	}
	
	// Test ctrl+c
	_, cmd = p.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("Expected quit command for ctrl+c")
	}
}


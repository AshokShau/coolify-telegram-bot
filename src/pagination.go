package src

import "fmt"

type PageButton struct {
	Text string
	Data string
}

// Paginate calculates the start and end indices for a slice of items and generates navigation buttons.
func Paginate(totalItems int, currentPage int, pageSize int, callbackPrefix string) (startIndex, endIndex int, buttons []PageButton) {
	if totalItems == 0 {
		return 0, 0, nil
	}

	totalPages := (totalItems + pageSize - 1) / pageSize
	if currentPage < 1 {
		currentPage = 1
	}
	if currentPage > totalPages {
		currentPage = totalPages
	}

	startIndex = (currentPage - 1) * pageSize
	endIndex = startIndex + pageSize
	if endIndex > totalItems {
		endIndex = totalItems
	}

	if currentPage > 1 {
		buttons = append(buttons, PageButton{Text: "< Prev", Data: fmt.Sprintf("%s%d", callbackPrefix, currentPage-1)})
	}

	startPage := currentPage - 1
	endPage := currentPage + 1

	if startPage < 1 {
		startPage = 1
		endPage = 3
	}

	if endPage > totalPages {
		endPage = totalPages
		startPage = totalPages - 2
		if startPage < 1 {
			startPage = 1
		}
	}

	for p := startPage; p <= endPage; p++ {
		text := fmt.Sprintf("%d", p)
		if p == currentPage {
			text = fmt.Sprintf("· %d ·", p)
		}
		buttons = append(buttons, PageButton{Text: text, Data: fmt.Sprintf("%s%d", callbackPrefix, p)})
	}

	if currentPage < totalPages {
		buttons = append(buttons, PageButton{Text: "Next >", Data: fmt.Sprintf("%s%d", callbackPrefix, currentPage+1)})
	}

	return startIndex, endIndex, buttons
}

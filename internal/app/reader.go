package app

func (m *Model) scrollReaderDown(n int) {
	maxOffset := m.readerMaxOffset()
	m.ReaderOffset += n
	if m.ReaderOffset > maxOffset {
		m.ReaderOffset = maxOffset
	}
}

func (m *Model) scrollReaderUp(n int) {
	m.ReaderOffset -= n
	if m.ReaderOffset < 0 {
		m.ReaderOffset = 0
	}
}

func (m Model) readerPageSize() int {
	size := m.Height - 8
	if size < 5 {
		return 5
	}
	return size
}

func (m Model) readerMaxOffset() int {
	lines := m.readerLines()
	pageSize := m.readerPageSize()

	maxOffset := len(lines) - pageSize
	if maxOffset < 0 {
		return 0
	}

	return maxOffset
}

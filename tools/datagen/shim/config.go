package shim

type Config struct {
	SoftNodes  int  // SoftNodes is the search soft node count.
	HardNodes  int  // HardNodes is the search hard node count.
	Draw       bool // Draw enables draw adjudication.
	DrawAfter  int  // DrawAfter determines how many moves have to be played before considering draw adjudication.
	DrawMargin int  // DrawScore determines the margin for draw adjudication.
	DrawCount  int  // DrawCount is the minimum number of back to back positions for draw adjudication.
	Win        bool // Win enables win adjudication.
	WinAfter   int  // WinAfter determines how many moves have to be played before considering win adjudication.
	WinMargin  int  // WinScore determines the margin for win adjudication.
	WinCount   int  // WinCount is the minimum number of back to back positions for win adjudication.
}

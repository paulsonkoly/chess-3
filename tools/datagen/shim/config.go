package shim

type Config struct {
	SoftNodes  int  // softNodes is the search soft node count.
	HardNodes  int  // hardNodes is the search hard node count.
	Draw       bool // draw enables draw adjudication.
	DrawAfter  int  // draw after determines how many moves have to be played before considering draw adjudication.
	DrawMargin int  // drawScore determines the margin for draw adjudication.
	DrawCount  int  // drawCount is the minimum number of back to back positions for draw adjudication.
	Win        bool // win enables win adjudication.
	WinAfter   int  // winAfter determines how many moves have to be played before considering win adjudication.
	WinMargin  int  // winScore determines the margin for win adjudication.
	WinCount   int  // winCount is the minimum number of back to back positions for win adjudication.
}

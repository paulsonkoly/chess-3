package tui

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	tcell "github.com/gdamore/tcell/v2"
	"github.com/paulsonkoly/chess-3/tools/tuner/app"
)

// QueueDepth is the channel depth for tui updates.
const QueueDepth = 10

const (
	jobsTop   = 1
	jobsBot   = jobsTop + 16
	jobsLeft  = 50
	jobsRight = jobsLeft + 2 + jobWidth*3
	jobWidth  = 10

	infoLeft  = 15
	infoRight = 80

	epochTop   = 1
	epochBot   = epochTop + 1
	epochLeft  = infoLeft
	epochRight = infoRight

	batchTop   = epochBot
	batchBot   = batchTop + 1
	batchLeft  = infoLeft
	batchRight = infoRight

	batchTimeTop   = batchBot
	batchTimeBot   = batchTimeTop + 1
	batchTimeLeft  = infoLeft
	batchTimeRight = infoRight

	mseTop   = batchTimeBot
	mseBot   = mseTop + 1
	mseLeft  = infoLeft
	mseRight = infoRight

	lrTop   = mseBot
	lrBot   = lrTop + 1
	lrLeft  = infoLeft
	lrRight = infoRight

	kTop   = lrBot
	kBot   = kTop + 1
	kLeft  = infoLeft
	kRight = infoRight

	hostTop   = kBot
	hostBot   = hostTop + 1
	hostLeft  = infoLeft
	hostRight = infoRight
)

type Update interface {
	Draw(s tcell.Screen)
	Log()
}

type JobUpdate struct {
	ChunkIx   int
	JobIx     int
	StartTime time.Time
	TTL       time.Duration
}

func (u JobUpdate) Draw(s tcell.Screen) {
	y1 := jobsTop + u.ChunkIx
	y2 := min(jobsBot, y1+1)
	drawText(s, jobsLeft, y1, jobsLeft+2, y2, tcell.StyleDefault, "P")
	// seconds since the start of the hour
	elapsed := u.StartTime.Minute()*60 + u.StartTime.Second()
	drawText(
		s,
		jobsLeft+2+u.JobIx*jobWidth,
		y1,
		min(jobsRight, jobsLeft+2+(u.JobIx+1)*jobWidth),
		y2,
		tcell.StyleDefault,
		fmt.Sprintf("%d %d", elapsed, int(u.TTL.Seconds())),
	)
}

func (u JobUpdate) Log() {}

type ResultUpdate struct {
	ChunkIx int
	JobIx   int
}

func (u ResultUpdate) Draw(s tcell.Screen) {
	y1 := jobsTop + u.ChunkIx
	y2 := min(jobsBot, y1+1)
	drawText(s, jobsLeft, y1, jobsLeft+2, y2, tcell.StyleDefault.Foreground(tcell.ColorGreen), "D")
	drawText(
		s,
		jobsLeft+2+u.JobIx*jobWidth,
		y1,
		min(jobsRight, jobsLeft+2+(u.JobIx+1)*jobWidth),
		y2,
		tcell.StyleDefault,
		"recvd",
	)
}

func (u ResultUpdate) Log() {}

type EpochUpdate struct{ Epoch int }

func (u EpochUpdate) Draw(s tcell.Screen) {
	drawText(s, epochLeft, epochTop, epochRight, epochBot, tcell.StyleDefault, strconv.Itoa(u.Epoch))
}

func (u EpochUpdate) Log() { slog.Info("new epoch", "epoch", u.Epoch) }

type BatchUpdate struct{ Start, End int }

func (u BatchUpdate) Draw(s tcell.Screen) {
	drawText(
		s,
		batchLeft,
		batchTop,
		batchRight,
		batchBot,
		tcell.StyleDefault,
		fmt.Sprintf("%d - %d", u.Start, u.End),
	)
}

func (u BatchUpdate) Log() {}

type BatchTimeUpdate struct{ Duration time.Duration }

func (u BatchTimeUpdate) Draw(s tcell.Screen) {
	drawText(
		s,
		batchTimeLeft,
		batchTimeTop,
		batchTimeRight,
		batchTimeBot,
		tcell.StyleDefault,
		fmt.Sprintf("%d(s)", int(u.Duration.Seconds())),
	)
}

func (u BatchTimeUpdate) Log() {}

type MSEUpdate struct{ MSE float64 }

func (u MSEUpdate) Draw(s tcell.Screen) {
	drawText(s, mseLeft, mseTop, mseRight, mseBot, tcell.StyleDefault, fmt.Sprintf("%f", u.MSE))
}

func (u MSEUpdate) Log() { slog.Info("MSE updated", "MSE", u.MSE) }

type LRUpdate struct{ LR float64 }

func (u LRUpdate) Draw(s tcell.Screen) {
	drawText(s, lrLeft, lrTop, lrRight, lrBot, tcell.StyleDefault, fmt.Sprintf("%f", u.LR))
}

func (u LRUpdate) Log() { slog.Info("LR updated", "LR", u.LR) }

type KUpdate struct {
	K    float64
	Step float64
}

func (u KUpdate) Draw(s tcell.Screen) {
	drawText(s, kLeft, kTop, kRight, kBot, tcell.StyleDefault, fmt.Sprintf("%f (%f)", u.K, u.Step))
}

func (u KUpdate) Log() { slog.Info("K updated", "K", u.K, "step", u.Step) }

type HostUpdate struct {
	Host string
	Port int
}

func (u HostUpdate) Draw(s tcell.Screen) {
	drawText(s, hostLeft, hostTop, hostRight, hostBot, tcell.StyleDefault, fmt.Sprintf("%s:%d", u.Host, u.Port))
}

func (u HostUpdate) Log() {
	slog.Info("listening for incoming connections", "host", u.Host, "port", u.Port)
}

type MsgUpdate struct {
	Msg  string
	Args []any
}

func (u MsgUpdate) Draw(s tcell.Screen) {
	xSize, ySize := s.Size()
	if len(u.Args) == 0 {
		drawText(s, 0, ySize-1, xSize, ySize, tcell.StyleDefault, fmt.Sprintf("%s", u.Msg))
	} else {
		drawText(s, 0, ySize-1, xSize, ySize, tcell.StyleDefault, fmt.Sprintf("%s %v", u.Msg, u.Args))
	}
}

func (u MsgUpdate) Log() {
	slog.Info(u.Msg, u.Args...)
}

func Run(useTui bool, updates <-chan Update) {
	if useTui {
		runWithTui(updates)
	} else {
		runWithoutTui(updates)
	}
}

// runWithoutTui just debug logs a few updates
func runWithoutTui(updates <-chan Update) {
	for update := range updates {
		update.Log()
	}
}

func runWithTui(updates <-chan Update) {
	s, err := tcell.NewScreen()
	if err != nil {
		slog.Error("tui screen", "error", err)
		os.Exit(app.ExitFailure)
	}

	if err := s.Init(); err != nil {
		slog.Error("tui screen", "error", err)
		os.Exit(app.ExitFailure)
	}

	s.Clear()

	drawText(s, 1, epochTop, epochLeft, epochBot, tcell.StyleDefault, "epoch")
	drawText(s, 1, batchTop, batchLeft, batchBot, tcell.StyleDefault, "batch")
	drawText(s, 1, batchTimeTop, batchTimeLeft, batchTimeBot, tcell.StyleDefault, "batch time")
	drawText(s, 1, mseTop, mseLeft, mseBot, tcell.StyleDefault, "mse")
	drawText(s, 1, lrTop, lrLeft, lrBot, tcell.StyleDefault, "lr")
	drawText(s, 1, kTop, kLeft, kBot, tcell.StyleDefault, "k")
	drawText(s, 1, hostTop, hostLeft, hostBot, tcell.StyleDefault, "host")

	mu := sync.Mutex{}

	go func() {
		for update := range updates {
			mu.Lock()
			update.Draw(s)
			s.Show()
			mu.Unlock()
		}
	}()

	go func() {
		for {
			ev := s.PollEvent()
			switch ev := ev.(type) {

			case *tcell.EventResize:
				mu.Lock()
				s.Sync()
				mu.Unlock()

			case *tcell.EventKey:
				if ev.Key() == tcell.KeyCtrlC {
					s.Fini()
					return
				}
			}
		}
	}()
}

func drawText(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	sIx := 0
	runes := []rune(text)
	for x := x1; x < x2; x++ {
		for y := y1; y < y2; y++ {
			c := ' '
			if sIx < len(runes) {
				c = runes[sIx]
			}
			s.SetContent(x, y, c, nil, style)
			sIx++
		}
	}
}

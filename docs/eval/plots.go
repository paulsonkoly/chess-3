package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"text/template"

	"github.com/paulsonkoly/chess-3/eval"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

const chessboardTemplate = `# Set up the heatmap
set terminal pngcairo enhanced font "Arial,12" size 800,600
set output '{{.OutputFile}}'
set title "{{.Title}} (Chess Coordinates)"
set xlabel "File (a-h)"
set ylabel "Rank (1-8)"
set xrange [0.5:8.5]
set yrange [0.5:8.5]
set xtics ("a" 1, "b" 2, "c" 3, "d" 4, "e" 5, "f" 6, "g" 7, "h" 8)
set ytics ("1" 1, "2" 2, "3" 3, "4" 4, "5" 5, "6" 6, "7" 7, "8" 8)
set palette defined (-70 "blue", 0 "white", 96 "red")
set cbrange [-70:100]
set pm3d map
set size square

# Your original data (first row = Rank 1, last row = Rank 8)
$DATA << EOD
{{.Data}}
EOD

# Plot with perfect chess alignment
plot $DATA matrix using ($1+1):(8-$2):3 with image notitle, \
     $DATA matrix using ($1+1):(8-$2):(sprintf("%d", $3)) with labels font ",8" notitle
`

const linePlotTemplate = `# Set up the plot
set terminal pngcairo enhanced font "Arial,12" size 800,600
set output '{{.OutputFile}}'
set title "{{.Title}}"
set xlabel "X-axis (0-{{.MaxX}})"
set ylabel "Values"
set xrange [0:{{.MaxX}}]
set grid

# Define the data
$DATA1 << EOD
{{.Data1}}
EOD

$DATA2 << EOD
{{.Data2}}
EOD

# Plot with connected points (no smoothing) and visible markers
plot $DATA1 using 1:2 with linespoints lt 1 pt 7 ps 1.5 lw 2 title "Middle Game", \
     $DATA2 using 1:2 with linespoints lt 2 pt 7 ps 1.5 lw 2 title "End Game"
`

const singleLinePlotTemplate = `# Set up the plot
set terminal pngcairo enhanced font "Arial,12" size 800,600
set output '{{.OutputFile}}'
set title "{{.Title}}"
set xlabel "X-axis (0-{{.MaxX}})"
set ylabel "Values"
set xrange [0:{{.MaxX}}]
set grid

# Define the data
$DATA << EOD
{{.Data}}
EOD

# Plot with connected points (no smoothing) and visible markers
plot $DATA using 1:2 with linespoints lt 1 pt 7 ps 1.5 lw 2 title "Values"
`

const barChartTemplate = `# Set up the bar chart
set terminal pngcairo enhanced font "Arial,12" size 600,400
set output '{{.OutputFile}}'
set title "{{.Title}}"
set style data histogram
set style histogram cluster gap 1
set style fill solid border -1
set boxwidth 0.9
set xtics ("Middle Game" 0, "End Game" 1)
set ylabel "Score Value"
set grid y

# Define the data
$DATA << EOD
0 {{.MGValue}}
1 {{.EGValue}}
EOD

# Plot the bar chart
plot $DATA using 2:xtic(1) with boxes lc rgb "#3070B3" notitle
`

const markdownTemplate = `# Chess Evaluation Coefficients Visualization

{{range .Sections}}
## {{.Title}}

![{{.Title}}]({{.ImagePath}})

{{if .Description}}
{{.Description}}
{{end}}
{{end}}
`

type ChessboardPlot struct {
	Title      string
	OutputFile string
	Data       string
}

type LinePlot struct {
	Title      string
	OutputFile string
	Data1      string
	Data2      string
	MaxX       int
}

type SingleLinePlot struct {
	Title      string
	OutputFile string
	Data       string
	MaxX       int
}

type BarChart struct {
    Title      string
    OutputFile string
    MGValue    int
    EGValue    int
}

type MarkdownSection struct {
	Title       string
	ImagePath   string
	Description string
}

func main() {
	var sections []MarkdownSection
	val := reflect.ValueOf(eval.Coefficients)
	typ := val.Type()

	for i := range val.NumField() {
		field := val.Field(i)
		fieldType := typ.Field(i)
		fieldName := fieldType.Name

		switch field.Kind() {
		case reflect.Array:
			// Check if it's a chessboard (8x8 array)
			if isChessboard(field) {
				processChessboard(field, fieldName, &sections)
			} else if isPairedArray(field) {
				processPairedArray(field, fieldName, &sections)
			} else {
				fmt.Printf("Unhandled array type for field %s\n", fieldName)
			}
		default:
			fmt.Printf("Unhandled field type %s for field %s\n", field.Kind(), fieldName)
		}
	}

	// Generate markdown document
	generateMarkdown(sections)
}

func isChessboard(field reflect.Value) bool {
	if field.Type().Kind() != reflect.Array {
		return false
	}

	// Check for [12][64]Score (PSqT)
	if field.Type().Len() == 12 {
		inner := field.Index(0)
		return inner.Type().Kind() == reflect.Array && inner.Type().Len() == 64
	}

	// Check for [64]Score (single chessboard)
	return field.Type().Len() == 64
}

func isPairedArray(field reflect.Value) bool {
	if field.Type().Kind() != reflect.Array {
		return false
	}

	// Check for arrays that should be plotted (either paired or single sequences)
	return true
}

func processChessboard(field reflect.Value, name string, sections *[]MarkdownSection) {
	if field.Type().Len() == 12 {
		// Handle PSqT which has 12 chessboards (6 pieces × 2 phases)
		pieceNames := [...]string{
			"Pawn", "Knight", "Bishop", "Rook", "Queen", "King",
		}
		phaseNames := [...]string{
			"MiddleGame", "EndGame", 
		}

		for i := range field.Type().Len() {
			chessboard := field.Index(i)
			pieceName := pieceNames[i / 2]
			phaseName := phaseNames[i % 2]
			outputFile := fmt.Sprintf("psqt_%s_%s.png", strings.ToLower(pieceName), strings.ToLower(phaseName))
			title := fmt.Sprintf("PSqT %s %s", pieceName, phaseName)
			generateChessboardPlot(chessboard, title, outputFile)
			*sections = append(*sections, MarkdownSection{
				Title:     title,
				ImagePath: outputFile,
			})
		}
	} else {
		// Handle single chessboard case (though we don't have any in our struct)
		outputFile := fmt.Sprintf("%s.png", strings.ToLower(name))
		generateChessboardPlot(field, name, outputFile)
		*sections = append(*sections, MarkdownSection{
			Title:     name,
			ImagePath: outputFile,
		})
	}
}

func generateChessboardPlot(field reflect.Value, title, outputFile string) {
	var buf bytes.Buffer
	for i := range 8 {
		for j := range 8 {
			// Chessboards are stored with a1 in the first position, but we want to display
			// with rank 1 at the bottom, so we reverse the rows
			idx := i*8 + j
			val := field.Index(idx).Interface().(Score)
			buf.WriteString(fmt.Sprintf("%d ", val))
		}
		buf.WriteString("\n")
	}

	data := ChessboardPlot{
		Title:      title,
		OutputFile: outputFile,
		Data:       buf.String(),
	}

	tmpl, err := template.New("chessboard").Parse(chessboardTemplate)
	if err != nil {
		fmt.Printf("Error creating template: %v\n", err)
		return
	}

	var script bytes.Buffer
	if err := tmpl.Execute(&script, data); err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		return
	}

	runGnuplot(script.String(), outputFile)
}

func generateTruncatedChessboardPlot(field reflect.Value, title, outputFile string, startRank, endRank int) {
	var buf bytes.Buffer

	// Add the actual data for ranks 3-7
	for rank := range 5 {
		for file := range 8 {
			idx := rank*8 + file
			val := field.Index(idx).Interface().(Score)
			buf.WriteString(fmt.Sprintf("%d ", val))
		}
		buf.WriteString("\n")
	}
	// Fill ranks 1-2 with zeros
	for range 3 {
		for range 8 {
			buf.WriteString("0 ")
		}
		buf.WriteString("\n")
	}

	data := ChessboardPlot{
		Title:      title,
		OutputFile: outputFile,
		Data:       buf.String(),
	}

	// Use the original chessboard template
	tmpl, err := template.New("chessboard").Parse(chessboardTemplate)
	if err != nil {
		fmt.Printf("Error creating template: %v\n", err)
		return
	}

	var script bytes.Buffer
	if err := tmpl.Execute(&script, data); err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		return
	}

	runGnuplot(script.String(), outputFile)
}

func processPairedArray(field reflect.Value, name string, sections *[]MarkdownSection) {
	// Special case for KnightOutpost
	if name == "KnightOutpost" {
		processKnightOutpost(field, name, sections)
		return
	}

	// Check if this is a paired array (middle game/end game)
	if field.Type().Len() == 2 && field.Type().Elem().Kind() != reflect.Array {
		mg := field.Index(0)
		eg := field.Index(1)
		outputFile := fmt.Sprintf("%s.png", strings.ToLower(name))
		processSingleValuePair(mg, eg, name, outputFile, sections)
		return
	}

	// Handle single sequence arrays (like BishopPair)
	if field.Type().Len() != 2 {
		outputFile := fmt.Sprintf("%s.png", strings.ToLower(name))
		processSingleSequence(field, name, outputFile, sections)
		return
	}

	// Handle regular paired arrays
	mg := field.Index(0)
	eg := field.Index(1)
	outputFile := fmt.Sprintf("%s.png", strings.ToLower(name))

	switch {
	case mg.Kind() == reflect.Array && mg.Len() > 0 && mg.Index(0).Kind() == reflect.Array:
		// Nested arrays like [2][9]Score (MobilityKnight)
		processNestedPairedArray(mg, eg, name, outputFile, sections)
	case mg.Kind() == reflect.Array:
		// Simple arrays like [2][7]Score (PieceValues)
		processSimplePairedArray(mg, eg, name, outputFile, sections)
	default:
		// Single values like [2]Score (ConnectedRooks)
		processSingleValuePair(mg, eg, name, outputFile, sections)
	}
}

func processSingleSequence(field reflect.Value, name, outputFile string, sections *[]MarkdownSection) {
	var data bytes.Buffer
	maxX := field.Len() - 1

	for i := range field.Len() {
		val := field.Index(i).Interface().(Score)
		data.WriteString(fmt.Sprintf("%d %d\n", i, val))
	}

	plotData := SingleLinePlot{
		Title:      name,
		OutputFile: outputFile,
		Data:       data.String(),
		MaxX:       maxX,
	}

	tmpl, err := template.New("singlelineplot").Parse(singleLinePlotTemplate)
	if err != nil {
		fmt.Printf("Error creating template: %v\n", err)
		return
	}

	var script bytes.Buffer
	if err := tmpl.Execute(&script, plotData); err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		return
	}

	runGnuplot(script.String(), outputFile)

	*sections = append(*sections, MarkdownSection{
		Title:     name,
		ImagePath: outputFile,
	})
}

func processNestedPairedArray(mg, eg reflect.Value, name, outputFile string, sections *[]MarkdownSection) {
	var data1, data2 bytes.Buffer
	maxX := mg.Len() - 1

	for i := range mg.Len() {
		mgVal := mg.Index(i).Interface().(Score)
		egVal := eg.Index(i).Interface().(Score)
		data1.WriteString(fmt.Sprintf("%d %d\n", i, mgVal))
		data2.WriteString(fmt.Sprintf("%d %d\n", i, egVal))
	}

	data := LinePlot{
		Title:      name,
		OutputFile: outputFile,
		Data1:      data1.String(),
		Data2:      data2.String(),
		MaxX:       maxX,
	}

	tmpl, err := template.New("lineplot").Parse(linePlotTemplate)
	if err != nil {
		fmt.Printf("Error creating template: %v\n", err)
		return
	}

	var script bytes.Buffer
	if err := tmpl.Execute(&script, data); err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		return
	}

	runGnuplot(script.String(), outputFile)

	*sections = append(*sections, MarkdownSection{
		Title:     name,
		ImagePath: outputFile,
	})
}

func processSimplePairedArray(mg, eg reflect.Value, name, outputFile string, sections *[]MarkdownSection) {
	var data1, data2 bytes.Buffer
	maxX := mg.Len() - 1

	for i := range mg.Len() {
		mgVal := mg.Index(i).Interface().(Score)
		egVal := eg.Index(i).Interface().(Score)
		data1.WriteString(fmt.Sprintf("%d %d\n", i, mgVal))
		data2.WriteString(fmt.Sprintf("%d %d\n", i, egVal))
	}

	data := LinePlot{
		Title:      name,
		OutputFile: outputFile,
		Data1:      data1.String(),
		Data2:      data2.String(),
		MaxX:       maxX,
	}

	tmpl, err := template.New("lineplot").Parse(linePlotTemplate)
	if err != nil {
		fmt.Printf("Error creating template: %v\n", err)
		return
	}

	var script bytes.Buffer
	if err := tmpl.Execute(&script, data); err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		return
	}

	runGnuplot(script.String(), outputFile)

	*sections = append(*sections, MarkdownSection{
		Title:     name,
		ImagePath: outputFile,
	})
}

func processSingleValuePair(mg, eg reflect.Value, name, outputFile string, sections *[]MarkdownSection) {
    mgVal := mg.Interface().(Score)
    egVal := eg.Interface().(Score)

    data := BarChart{
        Title:      name,
        OutputFile: outputFile,
        MGValue:    int(mgVal),
        EGValue:    int(egVal),
    }

    tmpl, err := template.New("barchart").Parse(barChartTemplate)
    if err != nil {
        fmt.Printf("Error creating template: %v\n", err)
        return
    }

    var script bytes.Buffer
    if err := tmpl.Execute(&script, data); err != nil {
        fmt.Printf("Error executing template: %v\n", err)
        return
    }

    runGnuplot(script.String(), outputFile)

    *sections = append(*sections, MarkdownSection{
        Title:     name,
        ImagePath: outputFile,
        Description: fmt.Sprintf("Middle Game: %d | End Game: %d", mgVal, egVal),
    })
}

func processKnightOutpost(field reflect.Value, name string, sections *[]MarkdownSection) {
	mg := field.Index(0)
	eg := field.Index(1)

	// Process middle game
	mgOutput := fmt.Sprintf("%s_mg.png", strings.ToLower(name))
	generateTruncatedChessboardPlot(mg, name+" (Middle Game)", mgOutput, 3, 7)

	// Process end game
	egOutput := fmt.Sprintf("%s_eg.png", strings.ToLower(name))
	generateTruncatedChessboardPlot(eg, name+" (End Game)", egOutput, 3, 7)

	*sections = append(*sections, MarkdownSection{
		Title:     name + " (Middle Game)",
		ImagePath: mgOutput,
	})

	*sections = append(*sections, MarkdownSection{
		Title:     name + " (End Game)",
		ImagePath: egOutput,
	})
}

func runGnuplot(script, outputFile string) {
	cmd := exec.Command("gnuplot")
	cmd.Stdin = strings.NewReader(script)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running gnuplot for %s: %v\n", outputFile, err)
	}
}

func generateMarkdown(sections []MarkdownSection) {
	tmpl, err := template.New("markdown").Parse(markdownTemplate)
	if err != nil {
		fmt.Printf("Error creating markdown template: %v\n", err)
		return
	}

	file, err := os.Create("readme.md")
	if err != nil {
		fmt.Printf("Error creating markdown file: %v\n", err)
		return
	}
	defer file.Close()

	data := struct {
		Sections []MarkdownSection
	}{
		Sections: sections,
	}

	if err := tmpl.Execute(file, data); err != nil {
		fmt.Printf("Error generating markdown: %v\n", err)
	}
}

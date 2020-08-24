package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/cheggaaa/pb.v1"
)

var totalLines int

type ProgressBar struct {
	okPb    *pb.ProgressBar
	errorPb *pb.ProgressBar
	pool    *pb.Pool
}

func NewProgressBar(total int) *ProgressBar {
	totalLines = total
	okPb := makeProgressBar(total, "ok")
	errorPb := makeProgressBar(total, "error")

	return &ProgressBar{
		okPb,
		errorPb,
		nil,
	}
}

func (p *ProgressBar) IncrementOk() {
	p.okPb.Add(1)
}

func (p *ProgressBar) IncrementError() {
	p.errorPb.Add(1)
}

func (p *ProgressBar) Start() {
	pool, err := pb.StartPool(p.okPb, p.errorPb)
	if err != nil {
		panic(err)
	}
	p.pool = pool
	p.okPb.Start()
}

func (p *ProgressBar) Stop() {
	wg := new(sync.WaitGroup)
	for _, bar := range []*pb.ProgressBar{p.okPb, p.errorPb} {
		wg.Add(1)
		go func(cb *pb.ProgressBar) {
			cb.Finish()
			wg.Done()
		}(bar)
	}
	wg.Wait()
	// close pool
	_ = p.pool.Stop()
	processParametersCuttings()
}

func processParametersCuttings() {
	logFileErrorTotal, _ := os.Open(logFileErrorNameTotal)
	logFileErrorTotal.Seek(0, 0)
	scanner := bufio.NewScanner(logFileErrorTotal)

	var countError = 0

	for scanner.Scan() {
		line := scanner.Text()
		for _, parameter := range opts.parametersCutting {
			if strings.Contains(line, parameter) {
				file := mapFiles[parameter+"-error"]
				w := bufio.NewWriter(file)
				fmt.Fprintln(w, line)
				w.Flush()
			}
		}
		countError++
	}

	dir := fileLogDir + "/error-summary.csv"
	fileSummary, _ := os.Create(dir)
	w := bufio.NewWriter(fileSummary)
	fmt.Fprintln(w, fmt.Sprintln("CORTE,CORRECTOS,INCORRECTOS,%CORRECTOS,%INCCORRECTOS"))
	percentageOk, percentageError := getPercentage(countError, totalLines)
	fmt.Fprintln(w, fmt.Sprintln("TOTAL,", totalLines, ",", countError, ",", percentageOk, ",", percentageError))
	w.Flush()

	for _, parameter := range opts.parametersCutting {
		countTotal := getCountRows(parameter + "-src")
		countTotalError := getCountRows(parameter + "-error")
		percentageOk, percentageError := getPercentage(countTotalError, countTotal)
		fmt.Fprintln(w, fmt.Sprintln(strings.ToUpper(parameter), ",", countTotal, ",", countTotalError, ",", percentageOk, ",", percentageError))
		w.Flush()
	}

	table, _ := tablewriter.NewCSV(os.Stdout, dir, true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)   // Set Alignment
	table.Render()

}

func getPercentage(totalError int, total int) (string, string) {
	var totalOk = total - totalError
	var percentageError float64 = 0
	var percentageOk float64 = 0

	if totalError != 0 {
		percentageError = float64(totalError) * 100 / float64(total)
	}

	if totalOk != 0 {
		percentageOk = float64(totalOk) * 100 / float64(total)
	}

	return fmt.Sprintf("%.2f", percentageOk), fmt.Sprintf("%.2f", percentageError)
}

func getCountRows(keyFile string) int {
	var count int = 0
	file := mapFiles[keyFile]
	file.Seek(0, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		count++
	}
	return count
}

func makeProgressBar(total int, prefix string) *pb.ProgressBar {
	bar := pb.New(total)
	bar.Prefix(prefix)
	bar.SetMaxWidth(120)
	bar.ShowElapsedTime = true
	bar.ShowTimeLeft = false
	return bar
}

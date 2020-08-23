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

func (p *ProgressBar) Stop(fileLogName string) {
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
	processParametersCuttings(fileLogName)
}

func processParametersCuttings(fileLogName string) {
	file, err := os.Open(fileLogName)
	if err != nil {
		panic("Error in " + fileLogName)
	}
	dir := fileLogDir + "/total.error"
	fileErrorTotal, _ := os.Create(dir)
	scanner := bufio.NewScanner(file)

	var countError = 0

	for scanner.Scan() {
		lineTemp := scanner.Text()
		line := lineTemp[47 : len(lineTemp)-2]

		w := bufio.NewWriter(fileErrorTotal)
		fmt.Fprintln(w, line)
		w.Flush()

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

	dir = fileLogDir + "/error-summary.csv"
	fileSummary, _ := os.Create(dir)
	w := bufio.NewWriter(fileSummary)
	fmt.Fprintln(w, fmt.Sprintln("CORTE,CORRECTOS,INCORRECTOS,%CORRECTOS,%INCCORRECTOS"))
	percentageOk, percentageError := getPercentage(countError, countTotal)
	fmt.Fprintln(w, fmt.Sprintln("TOTAL,", countTotal, ",", countError, ",", percentageOk, ",", percentageError))
	w.Flush()

	for _, parameter := range opts.parametersCutting {
		countTotalOk := getCountRows(parameter + "-src")
		countTotalError := getCountRows(parameter + "-error")
		percentageOk, percentageError := getPercentage(countTotalError, countTotalOk)
		fmt.Fprintln(w, fmt.Sprintln(strings.ToUpper(parameter), ",", countTotalOk, ",", countTotalError, ",", percentageOk, ",", percentageError))
		w.Flush()
	}

	table, _ := tablewriter.NewCSV(os.Stdout, dir, true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)   // Set Alignment
	table.Render()

}

func getPercentage(totalError int, total int) (string, string) {
	var percentageError float64
	var percentageOk float64
	if totalError != 0 && total != 0 {
		percentageError = float64(totalError) * 100 / float64(total)
		percentageOk = float64(100) - percentageError
	}else if totalError == 0 && total != 0 {
		percentageError = float64(0)
		percentageOk = float64(100) - percentageError
	}else if totalError == 0 && total == 0 {
		percentageError = float64(0)
		percentageOk = float64(0)
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

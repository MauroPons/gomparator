package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

type consumer struct {
	statusCodeOnly bool
	bar            *ProgressBar
	log            *logrus.Logger
	exclude        string
}

func NewConsumer(statusCodeOnly bool, bar *ProgressBar, log *logrus.Logger, exclude string) Consumer {
	return &consumer{
		statusCodeOnly: statusCodeOnly,
		bar:            bar,
		log:            log,
		exclude:        exclude,
	}
}

func appendMsgError(url string){
	fmt.Fprintln(writeFinalErrorFile, url)
	writeFinalErrorFile.Flush()
}

func (c *consumer) Consume(val HostsPair) {

	if val.HasErrors() {
		c.bar.IncrementError()
		for _, v := range val.Errors {
			//logFileTotal
			c.log.Errorln(v)
		}
		appendMsgError(val.RelURL)
		return
	}

	if val.EqualStatusCode() && c.statusCodeOnly {
		c.bar.IncrementOk()
		return
	}

	if !val.EqualStatusCode() {
		c.bar.IncrementError()
		c.log.Warnf(val.RelURL)
		appendMsgError(val.RelURL)
		return
	}

	leftJSON, err := unmarshal(val.Left.Body)
	if err != nil {
		c.bar.IncrementError()
		c.log.Errorf(val.RelURL, err)
		appendMsgError(val.RelURL)
		return
	}

	rightJSON, err := unmarshal(val.Right.Body)
	if err != nil {
		c.bar.IncrementError()
		c.log.Errorf(val.RelURL)
		appendMsgError(val.RelURL)
		return
	}

	if c.exclude != "" {
		Remove(leftJSON, c.exclude)
		Remove(rightJSON, c.exclude)
	}

	if !Equal(leftJSON, rightJSON) {
		c.bar.IncrementError()
		c.log.Warnf(val.RelURL)
		appendMsgError(val.RelURL)
		return
	}

	c.bar.IncrementOk()
}

func unmarshal(b []byte) (interface{}, error) {
	j, err := Unmarshal(b)
	if err != nil {
		return nil, err
	}

	return j, nil
}

package main

import (
	"io"
	"os"
	"sync"

	. "github.com/xaxys/oasis/api"
)

const PrintBufferSize = 50
const BytePoolSize = 512

var formatterLock sync.Mutex
var consolePrinter *oasisConsolePrinter

type oasisConsolePrinter struct {
	wg            sync.WaitGroup
	formatterList []Formatter
	printBuffer   chan []byte
	bytePool      sync.Pool
	output        io.Writer
}

func getConsolePrinter() *oasisConsolePrinter {
	if consolePrinter == nil {
		formatterLock.Lock()
		if consolePrinter == nil {
			consolePrinter = newConsolePrinter()
		}
		formatterLock.Unlock()
	}
	return consolePrinter
}

func newConsolePrinter() *oasisConsolePrinter {
	p := &oasisConsolePrinter{
		bytePool: sync.Pool{
			New: func() interface{} {
				b := make([]byte, BytePoolSize)
				return &b
			},
		},
		output: os.Stdout,
	}
	p.startPrinter()
	return p
}

func (p *oasisConsolePrinter) startPrinter() {
	p.printBuffer = make(chan []byte, PrintBufferSize)
	go func() {
		p.wg.Add(1)
		for {
			b, ok := <-p.printBuffer
			if !ok {
				break
			}
			s := string(b)
			b = b[0:0]
			p.bytePool.Put(&b)

			formatterLock.Lock()
			for _, f := range p.formatterList {
				s = f.Format(s)
			}
			formatterLock.Unlock()

			s = "\r" + s + "> "
			p.output.Write([]byte(s))
		}
		p.wg.Done()
	}()
}

func (p *oasisConsolePrinter) RegisterFormatter(f Formatter) {
	formatterLock.Lock()
	p.formatterList = append(p.formatterList, f)
	formatterLock.Unlock()
}

func (p *oasisConsolePrinter) ClearFormatter() {
	formatterLock.Lock()
	p.formatterList = []Formatter{}
	formatterLock.Unlock()
}

func (p *oasisConsolePrinter) Write(b []byte) (int, error) {
	bc := *(p.bytePool.Get().(*[]byte))
	bc = append(bc, b...)
	p.printBuffer <- bc
	return len(b), nil
}

func (p *oasisConsolePrinter) Stop() {
	close(p.printBuffer)
	p.wg.Wait()
}

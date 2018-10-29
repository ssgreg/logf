package logf

import (
	"io/ioutil"
)

type MultiAppender struct {
	delegatee []Appender
}

func (a *MultiAppender) Append(entry Entry) error {
	return a.forEach(func(d Appender) error {
		return d.Append(entry)
	})
}

func (a *MultiAppender) Flush() error {
	return a.forEach(func(d Appender) error {
		return d.Flush()
	})
}

func (a *MultiAppender) Close() error {
	return a.forEach(func(d Appender) error {
		return d.Close()
	})
}

func (a *MultiAppender) forEach(fn func(Appender) error) error {
	var err error
	for _, d := range a.delegatee {
		if errd := fn(d); errd != nil && err != nil {
			err = errd
		}
	}
	return err
}

// type Writer2Appender struct {
// 	writer    io.Writer
// 	formatter Formatter
// }

// func (a *Writer2Appender) Append(entry *Entry) error {
// 	return a.formatter.Format(a.writer, entry)
// }

// func (a *Writer2Appender) Flush() error {
// 	return nil
// }

// func (a *Writer2Appender) Close() error {
// 	return nil
// }

type DiscardAppender struct {
}

func (a *DiscardAppender) Append(entry Entry) error {
	return nil
}

func (a *DiscardAppender) Flush() error {
	return nil
}

func (a *DiscardAppender) Close() error {
	return nil
}

// type FileAppender struct {
// 	file      *os.File
// 	formatter Formatter
// 	buf       *Buffer
// }

// func (a *FileAppender) Append(entry Entry) error {
// 	return a.formatter.Format(a.buf, entry)
// }

// func (a *FileAppender) Flush() error {
// 	a.buf.Flush()
// 	return a.buf.Error()
// }

// func (a *FileAppender) Close() error {
// 	a.buf.Flush()
// 	err := a.file.Close()
// 	if a.buf.Error() != nil {
// 		return a.buf.Error()
// 	}
// 	return err
// }

// func NewFileAppender(filename string, formatter Formatter) Appender {
// 	w, _ := os.Create(filename)
// 	return &FileAppender{w, formatter, NewBuffer(w)}
// }

type StdoutAppender struct {
	formatter Encoder
	buf       *Buffer
}

func (a *StdoutAppender) Append(entry Entry) error {
	return a.formatter.Format(a.buf, entry)
}

func (a *StdoutAppender) Flush() error {
	a.buf.Flush()
	return a.buf.Error()
}

func (a *StdoutAppender) Close() error {
	a.buf.Flush()
	if a.buf.Error() != nil {
		return a.buf.Error()
	}

	return nil
}

func NewStdoutAppender(formatter Encoder) Appender {
	return &StdoutAppender{formatter, NewBuffer(ioutil.Discard)}
}

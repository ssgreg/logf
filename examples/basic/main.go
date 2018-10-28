package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"unsafe"
)

// type Enabler func(logf.Level) bool

// func (f Enabler) Enabled(lvl logf.Level) bool {
// 	return f(lvl)
// }

// type Encoder struct{}

// func (e Encoder) Print(k, v string) {
// 	fmt.Println(k, v)
// }

// func To(i interface{}) func(string, string) interface{} {
// 	return i.(func(string, string) interface{})
// }

// func Field(k, v string, next func(*Encoder)) func(e *Encoder) {
// 	return func(e *Encoder) {
// 		e.Print(k, v)
// 		if next != nil {
// 			next(e)
// 		}
// 	}
// }

// func test(logger *logf.MyLogger) {
// 	for i := 0; i < 1000000; i++ {

// 		rnd := make([]int, 1000)
// 		for j := 0; j < 1000; j++ {
// 			rnd[j] = rand.Int()
// 		}

// 		if i%10000 == 0 {
// 			runtime.GC()
// 			fmt.Println("gc", i/10000)
// 		}

// 		logger.Info("test", logf.ConstInts("k", rnd))
// 	}

// }

func Get() uintptr {
	rnd := make([]int, 100000)
	for j := 0; j < 100000; j++ {
		rnd[j] = rand.Int()
	}

	addr := uintptr(unsafe.Pointer(&rnd))

	return addr
	// fmt.Println(*(*[unsafe.Sizeof(s2)]byte)(addr))
	// // addr1 := uintptr(addr) + 1
	// // fmt.Println(*(*byte)(unsafe.Pointer(addr1)))
	// // (*int16)(unsafe.Pointer(addr))

}

func GetPointer() []byte {
	rnd := make([]byte, 1000)
	rand.Read(rnd)
	rnd1 := string(rnd)

	return *(*[]byte)(unsafe.Pointer(&rnd1))

	// return addr
	// fmt.Println(*(*[unsafe.Sizeof(s2)]byte)(addr))
	// // addr1 := uintptr(addr) + 1
	// // fmt.Println(*(*byte)(unsafe.Pointer(addr1)))
	// // (*int16)(unsafe.Pointer(addr))

}

func main() {
	// addr := Get()
	for i := 0; i < 1000000; i++ {
		addr := GetPointer()
		runtime.GC()
		// fmt.Println(*(*[]int)(unsafe.Pointer(addr)))
		fmt.Println(*(*string)(unsafe.Pointer(&addr)))
	}

	// channel := logf.NewBufferedChannel(0)
	// logger := logf.NewLogger(logf.Info, channel)
	// defer channel.Close()
	// defer fmt.Println("test exited")
	// defer runtime.GC()

	// // its := []int{0, 1, 2, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 0, 1, 2, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	// test(logger)
}

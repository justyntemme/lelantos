package main

import (
	"C"
	"fmt"
	"os"
	"unicode/utf16"
	"unsafe"

	"github.com/go-ole/go-ole"
	"golang.org/x/sys/windows"
)

func main() {
	pePath := os.Args[0]

	var sa windows.SecurityAttributes
	sa.Length = uint32(unsafe.Sizeof(sa))
	sa.InheritHandle = 0

	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	handle, err := windows.CreateFile(
		StringToUTF16Ptr(pePath),
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ,
		&sa,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer windows.CloseHandle(handle)

	// Get the size of the file
	info, err := os.Stat(pePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	x := info.Size()
	size := uint32(x)

	// // Open the PE file
	// peFile, err := os.Open(pePath)
	// if err != nil {
	// 	fmt.Printf("Failed to open PE file: %v\n", err)
	// 	os.Exit(1)
	// }
	// defer peFile.Close()

	// Map the file into memory
	addr, err := windows.MapViewOfFile(handle, windows.FILE_MAP_READ, 0, 0, uintptr(size))
	if err != nil {
		fmt.Println("Error mapping file:", err)
		return
	}
	defer windows.UnmapViewOfFile(addr)

	// Read the mapped memory
	data := (*[1 << 30]byte)(unsafe.Pointer(addr))[:size]
	fmt.Println(string(data))

}

func StringToUTF16Ptr(str string) *uint16 {
	wchars := utf16.Encode([]rune(str + "\x00"))
	return &wchars[0]
}
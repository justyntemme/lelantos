package main

import (
	"C"
	"fmt"
	"os"
	"unsafe"

	"github.com/go-ole/go-ole"
	"golang.org/x/sys/windows"
)

func main() {
	pePath := os.Args[0]
	pePath = "C:\\calc.exe"

	var sa windows.SecurityAttributes
	sa.Length = uint32(unsafe.Sizeof(sa))
	sa.InheritHandle = 0

	ole.CoInitialize(0)
	defer ole.CoUninitialize()
	exePath := windows.StringToUTF16Ptr(pePath)

	handle, err := windows.CreateFile(
		exePath,
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ,
		&sa,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	fmt.Println(pePath)
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

	// Create a file mapping object
	mapping, err := windows.CreateFileMapping(handle, &sa, windows.PAGE_READONLY, 0, size, nil)
	if err != nil {
		fmt.Println("Error creating file mapping:", err)
		return
	}

	// Map the file into memory
	addr, err := windows.MapViewOfFile(mapping, windows.FILE_MAP_READ, 0, 0, uintptr(size))
	if err != nil {
		fmt.Println("Error mapping file:", err)
		return
	}
	defer windows.UnmapViewOfFile(addr)

	// Read the mapped memory
	data := (*[1 << 30]byte)(unsafe.Pointer(addr))[:size]
	fmt.Println(string(data))

	// Create a new process with the mapped memory as the image
	var si windows.StartupInfo
	var pi windows.ProcessInformation
	z := uint16(0)
	si.Cb = uint32(unsafe.Sizeof(si))
	if err = windows.CreateProcess(nil, windows.StringToUTF16Ptr("process"), &sa, &sa, false, windows.CREATE_SUSPENDED|windows.CREATE_NO_WINDOW, &z, &z, &si, &pi); err != nil {
		fmt.Println("Error creating process:", err)
		return
	}

	defer windows.CloseHandle(pi.Process)
	defer windows.CloseHandle(pi.Thread)

	// Copy the mapped memory into the new process's memory space
	var written uintptr
	windows.WriteProcessMemory(pi.Process, pi.PebBaseAddress, addr, size, &written)

	// Resume the new process's execution
	windows.ResumeThread(pi.Thread)

	// Wait for the new process to exit
	windows.WaitForSingleObject(pi.Process, windows.INFINITE)

	// Get the exit code
	var exitCode uint32
	windows.GetExitCodeProcess(pi.Process, &exitCode)
	fmt.Println("Process exited with code:", exitCode)
}

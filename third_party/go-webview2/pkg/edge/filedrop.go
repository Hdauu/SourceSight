package edge

// SourceSight patch: expose absolute paths of files dropped onto the page.
//
// The page calls chrome.webview.postMessageWithAdditionalObjects(FileDropMessage, files);
// WebView2 (runtime >= 1.0.1774.30) then hands the host ICoreWebView2File objects whose
// Path is the real location on disk — something a plain DOM File never reveals.
// Interface IDs are the shipped values from WebView2.h; the design spec in
// MicrosoftEdge/WebView2Feedback lists draft IIDs that do not match the runtime.

import (
	"unsafe"

	"github.com/jchv/go-webview2/internal/w32"
	"golang.org/x/sys/windows"
)

// FileDropMessage is the message string the page sends alongside the dropped files.
const FileDropMessage = "__sourcesight_files_dropped__"

// OnFilesDropped receives the absolute paths of dropped files. Nil disables the feature.
var OnFilesDropped func(paths []string)

var (
	iidWebMessageReceivedEventArgs2 = NewGUID("{06fc7ab7-c90c-4297-9389-33ca01cf6d5e}")
	iidObjectCollectionView         = NewGUID("{0f36fd87-4f69-4415-98da-888f89fb9a33}")
	iidFile                         = NewGUID("{f2c19559-6bc1-4583-a757-90021be9afec}")
)

type iCoreWebView2WebMessageReceivedEventArgs2Vtbl struct {
	_IUnknownVtbl
	GetSource                ComProc
	GetWebMessageAsJSON      ComProc
	TryGetWebMessageAsString ComProc
	GetAdditionalObjects     ComProc
}

type iCoreWebView2ObjectCollectionViewVtbl struct {
	_IUnknownVtbl
	GetCount        ComProc
	GetValueAtIndex ComProc
}

type iCoreWebView2FileVtbl struct {
	_IUnknownVtbl
	GetPath ComProc
}

// comObject is the common shape of every COM interface pointer: a pointer to its vtable.
type comObject struct {
	vtbl *_IUnknownVtbl
}

func queryInterface(obj unsafe.Pointer, iid *GUID) unsafe.Pointer {
	var out unsafe.Pointer
	hr, _, _ := (*comObject)(obj).vtbl.QueryInterface.Call(
		uintptr(obj),
		uintptr(unsafe.Pointer(iid)),
		uintptr(unsafe.Pointer(&out)),
	)
	if hr != 0 {
		return nil
	}
	return out
}

func release(obj unsafe.Pointer) {
	if obj != nil {
		_, _, _ = (*comObject)(obj).vtbl.Release.Call(uintptr(obj))
	}
}

// droppedFilePaths extracts ICoreWebView2File paths from a web message, or nil if the
// runtime is too old or the message carried no files.
func droppedFilePaths(args *iCoreWebView2WebMessageReceivedEventArgs) []string {
	args2 := queryInterface(unsafe.Pointer(args), iidWebMessageReceivedEventArgs2)
	if args2 == nil {
		debugf("filedrop: runtime sin ICoreWebView2WebMessageReceivedEventArgs2 (WebView2 viejo)")
		return nil
	}
	defer release(args2)

	var collection unsafe.Pointer
	vtbl := *(**iCoreWebView2WebMessageReceivedEventArgs2Vtbl)(args2)
	if hr, _, _ := vtbl.GetAdditionalObjects.Call(uintptr(args2), uintptr(unsafe.Pointer(&collection))); hr != 0 || collection == nil {
		debugf("filedrop: get_AdditionalObjects hr=%#x", hr)
		return nil
	}
	defer release(collection)

	view := queryInterface(collection, iidObjectCollectionView)
	if view == nil {
		debugf("filedrop: QueryInterface ObjectCollectionView fallo")
		return nil
	}
	defer release(view)
	viewVtbl := *(**iCoreWebView2ObjectCollectionViewVtbl)(view)

	var count uint32
	if hr, _, _ := viewVtbl.GetCount.Call(uintptr(view), uintptr(unsafe.Pointer(&count))); hr != 0 {
		debugf("filedrop: get_Count hr=%#x", hr)
		return nil
	}

	debugf("filedrop: %d objeto(s) adjunto(s)", count)
	var paths []string
	for i := uint32(0); i < count; i++ {
		var item unsafe.Pointer
		if hr, _, _ := viewVtbl.GetValueAtIndex.Call(uintptr(view), uintptr(i), uintptr(unsafe.Pointer(&item))); hr != 0 || item == nil {
			continue
		}
		file := queryInterface(item, iidFile)
		release(item)
		if file == nil {
			debugf("filedrop: objeto %d no es ICoreWebView2File", i)
			continue
		}
		var p *uint16
		fileVtbl := *(**iCoreWebView2FileVtbl)(file)
		if hr, _, _ := fileVtbl.GetPath.Call(uintptr(file), uintptr(unsafe.Pointer(&p))); hr == 0 && p != nil {
			paths = append(paths, w32.Utf16PtrToString(p))
			windows.CoTaskMemFree(unsafe.Pointer(p))
		}
		release(file)
	}
	return paths
}

// DebugLog, if set, receives diagnostics about file drops.
var DebugLog func(format string, args ...interface{})

func debugf(format string, args ...interface{}) {
	if DebugLog != nil {
		DebugLog(format, args...)
	}
}

package client

type ExportType int

func (t ExportType) String() string {
	switch t {
	case ExportDebug:
		return "--export-debug"
	case ExportRelease:
		return "--export-release"
	case ExportPack:
		return "--export-pack"
	default:
		panic("unknown export type")
	}
}

const (
	ExportDebug ExportType = iota
	ExportRelease
	ExportPack
)

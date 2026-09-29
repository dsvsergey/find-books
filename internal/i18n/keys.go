package i18n

// Key identifies one user-facing text.
type Key int

const (
	_ Key = iota

	// index
	KeyLibraryNotFound
	KeyLibraryExists
	KeySchemaMismatch

	// library, libman, scan
	KeyFolder
	KeyNotAFolder
	KeyVolumeUnknown
	KeyDiskNotMounted
	KeyNotOnVolume
	KeyLibraryDiskOffline
	KeyLibraryFolderUnavailable

	// cli
	KeyCmdRootShort
	KeyFlagDB
	KeyFlagLang
	KeyNoLibraries
	KeyArgsNone
	KeyArgsExact
	KeyArgsMax
	KeyInterrupted
	KeyErrorPrefix
	KeyCmdAddUse
	KeyCmdAddShort
	KeyFlagName
	KeyScanIncomplete
	KeyCmdListShort
	KeyListHeader
	KeyIndexing
	KeyScanReport
	KeyProblemFiles
	KeyAndMore
	KeyCmdRemoveUse
	KeyCmdRemoveShort
	KeyRemoved
	KeyCmdSearchUse
	KeyCmdSearchShort
	KeyNeedQuery
	KeyNothingFound
	KeyFlagAuthor
	KeyFlagLimit
	KeyFlagJSON
	KeySearchHeader
	KeyCmdUpdateUse
	KeyCmdUpdateShort
	KeyNeedNameOrAll
	KeySkipOffline
	KeyFlagAll
	KeyCmdConfigShort
	KeyCmdConfigLangShort
	KeyLangSaved
	KeyEnvOverridesLang

	// tui
	KeyChoosePrompt
	KeyErrorStatus
	KeyDialogUnsupported
	KeyStopping
	KeyCanceling
	KeyRemoving
	KeyCanceled
	KeyChooseInFinder
	KeyAddInterruptedRegistered
	KeyAddCanceled
	KeyAddExists
	KeyDiskOffline
	KeyUpdateCanceled
	KeyReportLine
	KeyLibHelpBase
	KeyLibHelpEmpty
	KeyLibTitle
	KeyWorksInIndex
	KeyNoLibrariesYet
	KeyLibStats
	KeyIndexingJob
	KeyFindingFiles
	KeyEscCancel
	KeyConfirmRemove
	KeyEscToSearch
	KeyEscQuit
	KeySearchPlaceholder
	KeyPathCopied
	KeyDiskOfflineOpen
	KeyFileNotFound
	KeySearchHelp
	KeyLangToggle
	KeyHitsCount
	KeySearchError
	KeyWhereFile
	KeyWhereIn
	KeyLabelBook
	KeyLabelPath
	KeyLabelFile
	KeyLabelContents
	KeyLangSaveFailed
	KeyPreviewLoading
	KeyPreviewUnsupported
	KeyPreviewFailed
	KeyPreviewStale
	KeyPreviewHelp
	KeyIllustration

	keyCount
)
